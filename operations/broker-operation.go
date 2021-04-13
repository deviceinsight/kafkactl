package operations

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"sort"
	"strconv"
)

type Broker struct {
	Id      int32
	Address string
}

type GetBrokersFlags struct {
	OutputFormat string
}

type BrokerOperation struct {
}

func (operation *BrokerOperation) GetBrokers(flags GetBrokersFlags) error {

	var (
		err     error
		context ClientContext
		client  sarama.Client
		brokers []*sarama.Broker
	)

	if context, err = CreateClientContext(); err != nil {
		return err
	}

	if client, err = CreateClient(&context); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	brokers = client.Brokers()

	tableWriter := output.CreateTableWriter()

	if flags.OutputFormat == "" {
		if err := tableWriter.WriteHeader("ID", "ADDRESS"); err != nil {
			return err
		}
	} else if flags.OutputFormat == "compact" {
		tableWriter.Initialize()
	} else if flags.OutputFormat != "json" && flags.OutputFormat != "yaml" {
		return errors.Errorf("unknown outputFormat: %s", flags.OutputFormat)
	}

	brokerList := make([]Broker, 0, len(brokers))
	for _, broker := range brokers {
		brokerList = append(brokerList, Broker{Id: broker.ID(), Address: broker.Addr()})
	}

	sort.Slice(brokerList, func(i, j int) bool {
		return brokerList[i].Id < brokerList[j].Id
	})

	if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
		return output.PrintObject(brokerList, flags.OutputFormat)
	} else if flags.OutputFormat == "compact" {
		for _, t := range brokerList {
			if err := tableWriter.Write(t.Address); err != nil {
				return err
			}
		}
	} else {
		for _, t := range brokerList {
			if err := tableWriter.Write(strconv.Itoa(int(t.Id)), t.Address); err != nil {
				return err
			}
		}
	}

	if flags.OutputFormat == "compact" || flags.OutputFormat == "" {
		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}
