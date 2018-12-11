package operations

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

type ConsumerOperation struct {
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

func (operation *ConsumerOperation) Consume(topic string, flags ConsumerFlags) {

	operation.init(topic, flags)

	partitions := operation.findPartitions()
	if len(partitions) == 0 {
		output.Failf("Found no partitions to consume")
	}

	defer operation.Close()

	operation.consume(partitions)
}

func (operation *ConsumerOperation) init(topic string, args ConsumerFlags) {

	var err error

	clientContext := createClientContext()

	operation.topic = topic
	operation.args = args
	operation.timeout = time.Duration(0)

	operation.offsets, err = parseOffsets(args.Offsets)
	if err != nil {
		output.Failf("Failed to parse offsets: %s", err)
	}

	if operation.client, err = createClient(&clientContext); err != nil {
		output.Failf("failed to create client: %v", err)
	}

	if operation.offsetManager, err = createOffsetManager(operation.client, operation.args.ConsumerGroup); err != nil {
		output.Failf("failed to create offset manager: %v", err)
	}

	if operation.consumer, err = sarama.NewConsumerFromClient(operation.client); err != nil {
		output.Failf("failed to create consumer: %v", err)
	}
}

func (operation *ConsumerOperation) findPartitions() []int32 {
	var (
		partitions []int32
		res        []int32
		err        error
	)
	if partitions, err = operation.consumer.Partitions(operation.topic); err != nil {
		output.Failf("failed to read partitions for topic %v: %v", operation.topic, err)
	}

	if _, hasDefault := operation.offsets[-1]; hasDefault {
		return partitions
	}

	for _, p := range partitions {
		if _, ok := operation.offsets[p]; ok {
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

func (operation *ConsumerOperation) consume(partitions []int32) {
	var (
		wg sync.WaitGroup
	)

	wg.Add(len(partitions))
	for _, p := range partitions {
		go func(p int32) { defer wg.Done(); operation.consumePartition(p) }(p)
	}
	wg.Wait()
}

func (operation *ConsumerOperation) consumePartition(partition int32) {
	var (
		offsets Interval
		err     error
		pcon    sarama.PartitionConsumer
		start   int64
		end     int64
		ok      bool
	)

	if offsets, ok = operation.offsets[partition]; !ok {
		offsets, ok = operation.offsets[-1]
	}

	if start, err = operation.resolveOffset(offsets.start, partition); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read start offset for partition %v err=%v\n", partition, err)
		return
	}

	if end, err = operation.resolveOffset(offsets.end, partition); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read end offset for partition %v err=%v\n", partition, err)
		return
	}

	if pcon, err = operation.consumer.ConsumePartition(operation.topic, partition, start); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to consume partition %v err=%v\n", partition, err)
		return
	}

	operation.partitionLoop(pcon, partition, end)
}

func (operation *ConsumerOperation) resolveOffset(o offset, partition int32) (int64, error) {
	if !o.relative {
		return o.start, nil
	}

	var (
		res int64
		err error
	)

	if o.start == sarama.OffsetNewest || o.start == sarama.OffsetOldest {
		if res, err = operation.client.GetOffset(operation.topic, partition, o.start); err != nil {
			return 0, err
		}

		if o.start == sarama.OffsetNewest {
			res = res - 1
		}

		return res + o.diff, nil
	} else if o.start == offsetResume {
		if operation.args.ConsumerGroup == "" {
			return 0, fmt.Errorf("cannot resume without consumer-group argument")
		}
		pom := operation.getPOM(partition)
		next, _ := pom.NextOffset()
		return next, nil
	}

	return o.start + o.diff, nil
}

func (operation *ConsumerOperation) getPOM(p int32) sarama.PartitionOffsetManager {
	operation.Lock()
	if operation.poms == nil {
		operation.poms = map[int32]sarama.PartitionOffsetManager{}
	}
	pom, ok := operation.poms[p]
	if ok {
		operation.Unlock()
		return pom
	}

	pom, err := operation.offsetManager.ManagePartition(operation.topic, p)
	if err != nil {
		operation.Unlock()
		output.Failf("failed to create partition offset manager err=%v", err)
	}
	operation.poms[p] = pom
	operation.Unlock()
	return pom
}

func (operation *ConsumerOperation) Close() {
	operation.Lock()

	for p, pom := range operation.poms {
		if err := pom.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close partition offset manager for partition %v err=%v", p, err)
		}
	}

	if err := operation.consumer.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close consumer err=%v", err)
	}

	if err := operation.client.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close client err=%v", err)
	}

	operation.Unlock()
}

func (operation *ConsumerOperation) partitionLoop(pc sarama.PartitionConsumer, p int32, end int64) {
	//defer logClose(fmt.Sprintf("partition consumer %v", p), pc)
	var (
		timer   *time.Timer
		pom     sarama.PartitionOffsetManager
		timeout = make(<-chan time.Time)
	)

	if operation.args.ConsumerGroup != "" {
		pom = operation.getPOM(p)
	}

	for {
		if operation.timeout > 0 {
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(operation.timeout)
			timeout = timer.C
		}

		select {
		case <-timeout:
			fmt.Fprintf(os.Stderr, "consuming from partition %v timed out after %s\n", p, operation.timeout)
			return
		case err := <-pc.Errors():
			fmt.Fprintf(os.Stderr, "partition %v consumer encountered err %s", p, err)
			return
		case msg, ok := <-pc.Messages():
			if !ok {
				fmt.Fprintf(os.Stderr, "unexpected closed messages chan")
				return
			}

			m := operation.newConsumedMessage(msg, operation.encodeKey, operation.encodeValue)

			if operation.args.OutputFormat == "" {
				var row []string

				if operation.args.PrintKeys {
					if m.Key != nil {
						row = append(row, *m.Key)
					} else {
						row = append(row, "")
					}
				}
				if operation.args.PrintTimestamps {
					if m.Timestamp != nil {
						row = append(row, (*m.Timestamp).Format(time.RFC3339))
					} else {
						row = append(row, "")
					}
				}

				row = append(row, *m.Value)

				output.PrintStrings(strings.Join(row[:], "#"))

			} else {
				output.PrintObject(m, operation.args.OutputFormat)
			}

			//ctx := printContext{output: m, done: make(chan struct{})}
			//out <- ctx
			//<-ctx.done

			if operation.args.ConsumerGroup != "" {
				pom.MarkOffset(msg.Offset+1, "")
			}

			if end > 0 && msg.Offset >= end {
				return
			}
		}
	}
}

func (operation *ConsumerOperation) newConsumedMessage(m *sarama.ConsumerMessage, encodeKey, encodeValue string) consumedMessage {

	var key *string
	var timestamp *time.Time

	if operation.args.PrintKeys {
		key = encodeBytes(m.Key, encodeKey)
	}

	if operation.args.PrintTimestamps && !m.Timestamp.IsZero() {
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
