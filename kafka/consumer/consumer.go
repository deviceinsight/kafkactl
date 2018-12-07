package consumer

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/random-dwi/kafkactl/util"
	"github.com/random-dwi/kafkactl/util/output"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConsumerContext struct {
	sync.Mutex

	topic   string
	offsets map[int32]Interval
	args    ConsumerFlags

	encodeValue string
	encodeKey   string

	timeout time.Duration

	client        sarama.Client
	consumer      sarama.Consumer
	offsetManager sarama.OffsetManager
	poms          map[int32]sarama.PartitionOffsetManager
}

type ConsumerFlags struct {
	PrintKeys       bool
	PrintTimestamps bool
	ConsumerGroup   string
	OutputFormat    string
	Offsets         []string
}

type offset struct {
	relative bool
	start    int64
	diff     int64
}

type Interval struct {
	start offset
	end   offset
}

type consumedMessage struct {
	Partition int32
	Offset    int64
	Key       *string `json:",omitempty" yaml:",omitempty"`
	Value     *string
	Timestamp *time.Time `json:",omitempty" yaml:",omitempty"`
}

var offsetResume int64 = -3

func CreateConsumerContext(clientContext *util.ClientContext, topic string, args ConsumerFlags) ConsumerContext {

	var context ConsumerContext

	var err error

	context.topic = topic
	context.args = args
	context.timeout = time.Duration(0)

	context.offsets, err = parseOffsets(args.Offsets)
	if err != nil {
		output.Failf("Failed to parse offsets: %s", err)
	}

	if context.client, err = util.CreateClient(clientContext); err != nil {
		output.Failf("failed to create client: %v", err)
	}

	if context.offsetManager, err = util.CreateOffsetManager(context.client, context.args.ConsumerGroup); err != nil {
		output.Failf("failed to create offset manager: %v", err)
	}

	if context.consumer, err = sarama.NewConsumerFromClient(context.client); err != nil {
		output.Failf("failed to create consumer: %v", err)
	}

	return context
}

func (context *ConsumerContext) FindPartitions() []int32 {
	var (
		partitions []int32
		res        []int32
		err        error
	)
	if partitions, err = context.consumer.Partitions(context.topic); err != nil {
		output.Failf("failed to read partitions for topic %v: %v", context.topic, err)
	}

	if _, hasDefault := context.offsets[-1]; hasDefault {
		return partitions
	}

	for _, p := range partitions {
		if _, ok := context.offsets[p]; ok {
			res = append(res, p)
		}
	}

	return res
}

func parseOffset(str string) (offset, error) {
	result := offset{}
	re := regexp.MustCompile("(oldest|newest|resume)?(-|\\+)?(\\d+)?")
	matches := re.FindAllStringSubmatch(str, -1)

	if len(matches) == 0 || len(matches[0]) < 4 {
		return result, fmt.Errorf("Could not parse offset [%v]", str)
	}

	startStr := matches[0][1]
	qualifierStr := matches[0][2]
	intStr := matches[0][3]

	var err error
	if result.start, err = strconv.ParseInt(intStr, 10, 64); err != nil && len(intStr) > 0 {
		return result, fmt.Errorf("invalid offset [%v]", str)
	}

	if len(qualifierStr) > 0 {
		result.relative = true
		result.diff = result.start
		result.start = sarama.OffsetOldest
		if qualifierStr == "-" {
			result.start = sarama.OffsetNewest
			result.diff = -result.diff
		}
	}

	switch startStr {
	case "newest":
		result.relative = true
		result.start = sarama.OffsetNewest
	case "oldest":
		result.relative = true
		result.start = sarama.OffsetOldest
	case "resume":
		result.relative = true
		result.start = offsetResume
	}

	return result, nil
}

func parseOffsets(offsets []string) (map[int32]Interval, error) {
	defaultInterval := Interval{
		start: offset{relative: true, start: sarama.OffsetNewest},
		end:   offset{start: 1<<63 - 1},
	}

	if len(offsets) == 0 {
		return map[int32]Interval{-1: defaultInterval}, nil
	}

	result := map[int32]Interval{}
	for _, partitionInfo := range offsets {
		re := regexp.MustCompile("(all|\\d+)?=?([^:]+)?:?(.+)?")
		matches := re.FindAllStringSubmatch(strings.TrimSpace(partitionInfo), -1)
		if len(matches) != 1 || len(matches[0]) < 3 {
			return result, fmt.Errorf("invalid partition info [%v]", partitionInfo)
		}

		var partition int32
		start := defaultInterval.start
		end := defaultInterval.end
		partitionMatches := matches[0]

		// partition
		partitionStr := partitionMatches[1]
		if partitionStr == "all" || len(partitionStr) == 0 {
			partition = -1
		} else {
			i, err := strconv.Atoi(partitionStr)
			if err != nil {
				return result, fmt.Errorf("invalid partition [%v]", partitionStr)
			}
			partition = int32(i)
		}

		// start
		if len(partitionMatches) > 2 && len(strings.TrimSpace(partitionMatches[2])) > 0 {
			startStr := strings.TrimSpace(partitionMatches[2])
			o, err := parseOffset(startStr)
			if err == nil {
				start = o
			}
		}

		// end
		if len(partitionMatches) > 3 && len(strings.TrimSpace(partitionMatches[3])) > 0 {
			endStr := strings.TrimSpace(partitionMatches[3])
			o, err := parseOffset(endStr)
			if err == nil {
				end = o
			}
		}

		result[partition] = Interval{start, end}
	}

	return result, nil
}

func (context *ConsumerContext) Consume(partitions []int32) {
	var (
		wg sync.WaitGroup
	)

	wg.Add(len(partitions))
	for _, p := range partitions {
		go func(p int32) { defer wg.Done(); context.consumePartition(p) }(p)
	}
	wg.Wait()
}

func (context *ConsumerContext) consumePartition(partition int32) {
	var (
		offsets Interval
		err     error
		pcon    sarama.PartitionConsumer
		start   int64
		end     int64
		ok      bool
	)

	if offsets, ok = context.offsets[partition]; !ok {
		offsets, ok = context.offsets[-1]
	}

	if start, err = context.resolveOffset(offsets.start, partition); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read start offset for partition %v err=%v\n", partition, err)
		return
	}

	if end, err = context.resolveOffset(offsets.end, partition); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read end offset for partition %v err=%v\n", partition, err)
		return
	}

	if pcon, err = context.consumer.ConsumePartition(context.topic, partition, start); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to consume partition %v err=%v\n", partition, err)
		return
	}

	context.partitionLoop(pcon, partition, end)
}

func (context *ConsumerContext) resolveOffset(o offset, partition int32) (int64, error) {
	if !o.relative {
		return o.start, nil
	}

	var (
		res int64
		err error
	)

	if o.start == sarama.OffsetNewest || o.start == sarama.OffsetOldest {
		if res, err = context.client.GetOffset(context.topic, partition, o.start); err != nil {
			return 0, err
		}

		if o.start == sarama.OffsetNewest {
			res = res - 1
		}

		return res + o.diff, nil
	} else if o.start == offsetResume {
		if context.args.ConsumerGroup == "" {
			return 0, fmt.Errorf("cannot resume without consumer-group argument")
		}
		pom := context.getPOM(partition)
		next, _ := pom.NextOffset()
		return next, nil
	}

	return o.start + o.diff, nil
}

func (context *ConsumerContext) getPOM(p int32) sarama.PartitionOffsetManager {
	context.Lock()
	if context.poms == nil {
		context.poms = map[int32]sarama.PartitionOffsetManager{}
	}
	pom, ok := context.poms[p]
	if ok {
		context.Unlock()
		return pom
	}

	pom, err := context.offsetManager.ManagePartition(context.topic, p)
	if err != nil {
		context.Unlock()
		output.Failf("failed to create partition offset manager err=%v", err)
	}
	context.poms[p] = pom
	context.Unlock()
	return pom
}

func (context *ConsumerContext) Close() {
	context.Lock()

	for p, pom := range context.poms {
		if err := pom.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close partition offset manager for partition %v err=%v", p, err)
		}
	}

	if err := context.consumer.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close consumer err=%v", err)
	}

	if err := context.client.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close client err=%v", err)
	}

	context.Unlock()
}

func (context *ConsumerContext) partitionLoop(pc sarama.PartitionConsumer, p int32, end int64) {
	//defer logClose(fmt.Sprintf("partition consumer %v", p), pc)
	var (
		timer   *time.Timer
		pom     sarama.PartitionOffsetManager
		timeout = make(<-chan time.Time)
	)

	if context.args.ConsumerGroup != "" {
		pom = context.getPOM(p)
	}

	for {
		if context.timeout > 0 {
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(context.timeout)
			timeout = timer.C
		}

		select {
		case <-timeout:
			fmt.Fprintf(os.Stderr, "consuming from partition %v timed out after %s\n", p, context.timeout)
			return
		case err := <-pc.Errors():
			fmt.Fprintf(os.Stderr, "partition %v consumer encountered err %s", p, err)
			return
		case msg, ok := <-pc.Messages():
			if !ok {
				fmt.Fprintf(os.Stderr, "unexpected closed messages chan")
				return
			}

			m := context.newConsumedMessage(msg, context.encodeKey, context.encodeValue)

			if context.args.OutputFormat == "" {
				var row []string

				if context.args.PrintKeys {
					if m.Key != nil {
						row = append(row, *m.Key)
					} else {
						row = append(row, "")
					}
				}
				if context.args.PrintTimestamps {
					if m.Timestamp != nil {
						row = append(row, (*m.Timestamp).Format(time.RFC3339))
					} else {
						row = append(row, "")
					}
				}

				row = append(row, *m.Value)

				output.PrintStrings(strings.Join(row[:], "#"))

			} else {
				output.PrintObject(m, context.args.OutputFormat)
			}

			//ctx := printContext{output: m, done: make(chan struct{})}
			//out <- ctx
			//<-ctx.done

			if context.args.ConsumerGroup != "" {
				pom.MarkOffset(msg.Offset+1, "")
			}

			if end > 0 && msg.Offset >= end {
				return
			}
		}
	}
}

func (context *ConsumerContext) newConsumedMessage(m *sarama.ConsumerMessage, encodeKey, encodeValue string) consumedMessage {

	var key *string
	var timestamp *time.Time

	if context.args.PrintKeys {
		key = encodeBytes(m.Key, encodeKey)
	}

	if context.args.PrintTimestamps && !m.Timestamp.IsZero() {
		timestamp = &m.Timestamp
	}

	return consumedMessage{
		Partition: m.Partition,
		Offset:    m.Offset,
		Key:       key,
		Value:     encodeBytes(m.Value, encodeValue),
		Timestamp: timestamp,
	}
}

func encodeBytes(data []byte, encoding string) *string {
	if data == nil {
		return nil
	}

	var str string
	switch encoding {
	case "hex":
		str = hex.EncodeToString(data)
	case "base64":
		str = base64.StdEncoding.EncodeToString(data)
	default:
		str = string(data)
	}

	return &str
}
