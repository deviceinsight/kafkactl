package broker

import (
	"fmt"
	"sort"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Broker struct {
	ID      int32
	Address string
	Configs []internal.Config `json:",omitempty" yaml:",omitempty"`
}

type GetBrokersFlags struct {
	OutputFormat string
}

type DescribeBrokerFlags struct {
	OutputFormat string
}

type Operation struct {
}

func (operation *Operation) GetBrokers(flags GetBrokersFlags) error {

	var (
		err     error
		context internal.ClientContext
		client  sarama.Client
		admin   sarama.ClusterAdmin
		brokers []*sarama.Broker
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if client, err = internal.CreateClient(&context); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
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

		var configs []internal.Config

		brokerConfig := sarama.ConfigResource{
			Type: sarama.BrokerResource,
			Name: fmt.Sprint(broker.ID()),
		}

		if configs, err = internal.ListConfigs(&admin, brokerConfig); err != nil {
			return err
		}

		brokerList = append(brokerList, Broker{ID: broker.ID(), Address: broker.Addr(), Configs: configs})
	}

	sort.Slice(brokerList, func(i, j int) bool {
		return brokerList[i].ID < brokerList[j].ID
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
			if err := tableWriter.Write(strconv.Itoa(int(t.ID)), t.Address); err != nil {
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

func (operation *Operation) DescribeBroker(id int32, flags DescribeBrokerFlags) error {

	var (
		err     error
		context internal.ClientContext
		client  sarama.Client
		admin   sarama.ClusterAdmin
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if client, err = internal.CreateClient(&context); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	var broker *sarama.Broker

	for _, aBroker := range client.Brokers() {
		if aBroker.ID() == id {
			broker = aBroker
			break
		}
	}

	if broker == nil {
		return errors.Errorf("cannot find broker with id: %d", id)
	}

	var configs []internal.Config

	brokerConfig := sarama.ConfigResource{
		Type: sarama.BrokerResource,
		Name: fmt.Sprint(broker.ID()),
	}

	if configs, err = internal.ListConfigs(&admin, brokerConfig); err != nil {
		return err
	}

	brokerInfo := Broker{ID: broker.ID(), Address: broker.Addr(), Configs: configs}

	if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
		return output.PrintObject(brokerInfo, flags.OutputFormat)
	} else if flags.OutputFormat != "" && flags.OutputFormat != "wide" {
		return errors.Errorf("unknown outputFormat: %s", flags.OutputFormat)
	} else {

		tableWriter := output.CreateTableWriter()

		// write broker info table
		if err := tableWriter.WriteHeader("ID", "ADDRESS"); err != nil {
			return err
		}

		if err := tableWriter.Write(fmt.Sprint(brokerInfo.ID), brokerInfo.Address); err != nil {
			return err
		}

		if err := tableWriter.Flush(); err != nil {
			return err
		}

		output.PrintStrings("")

		// first write config table
		if err := tableWriter.WriteHeader("CONFIG", "VALUE"); err != nil {
			return err
		}

		for _, c := range brokerInfo.Configs {
			if err := tableWriter.Write(c.Name, c.Value); err != nil {
				return err
			}
		}

		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (operation *Operation) listBrokerIds() ([]string, error) {

	var (
		err     error
		context internal.ClientContext
		client  sarama.Client
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return nil, err
	}

	if client, err = internal.CreateClient(&context); err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	var brokerIds = make([]string, 0)

	for _, broker := range client.Brokers() {
		brokerIds = append(brokerIds, fmt.Sprint(broker.ID()))
	}

	return brokerIds, nil
}

func CompleteBrokerIds(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	brokerIds, err := (&Operation{}).listBrokerIds()

	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return brokerIds, cobra.ShellCompDirectiveNoFileComp
}

func FromYaml(yamlString string) (Broker, error) {
	var broker Broker
	err := yaml.Unmarshal([]byte(yamlString), &broker)
	return broker, err
}
