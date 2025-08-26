package broker

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Broker struct {
	ID      int32             `json:",omitempty" yaml:",omitempty"`
	Address string            `json:",omitempty" yaml:",omitempty"`
	Configs []internal.Config `json:",omitempty" yaml:",omitempty"`
}

type GetBrokersFlags struct {
	OutputFormat string
}

type DescribeBrokerFlags struct {
	AllConfigs   bool
	OutputFormat string
}

type Operation struct {
}

type AlterBrokerFlags struct {
	ValidateOnly bool
	Configs      []string
}

func (operation *Operation) AlterBroker(id string, flags AlterBrokerFlags) error {
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

	if id == "default" {
		id = ""
	}

	var broker *sarama.Broker

	for _, aBroker := range client.Brokers() {
		if strconv.Itoa(int(aBroker.ID())) == id {
			broker = aBroker
			break
		}
	}

	if id != "" && broker == nil {
		return errors.Errorf("cannot find broker with id: %s", id)
	}

	var existingConfigs []sarama.ConfigEntry
	brokerConfig := sarama.ConfigResource{
		Type: sarama.BrokerResource,
		Name: id,
	}

	if existingConfigs, err = admin.DescribeConfig(brokerConfig); err != nil {
		return errors.Wrap(err, "failed to describe config for broker")
	}

	mergedConfigEntries := make(map[string]*string)
	for _, existingConfig := range existingConfigs {
		if !existingConfig.ReadOnly && !existingConfig.Default {
			mergedConfigEntries[existingConfig.Name] = &(existingConfig.Value)
		}
	}

	for _, config := range flags.Configs {
		configParts := strings.Split(config, "=")

		if len(configParts) == 2 {
			if len(configParts[1]) == 0 {
				delete(mergedConfigEntries, configParts[0])
			} else {
				mergedConfigEntries[configParts[0]] = &configParts[1]
			}
		}
	}

	if err = admin.AlterConfig(sarama.BrokerResource, id, mergedConfigEntries, flags.ValidateOnly); err != nil {
		return errors.Errorf("could not alter broker config '%s': %v", id, err)
	}

	if !flags.ValidateOnly {
		output.Infof("config has been altered")
		return nil
	}

	return operation.printConfig(mergedConfigEntries)
}

func (operation *Operation) printConfig(configs map[string]*string) error {

	keys := make([]string, 0, len(configs))
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	configTableWriter := output.CreateTableWriter()
	if err := configTableWriter.WriteHeader("CONFIG", "VALUE"); err != nil {
		return err
	}

	for _, key := range keys {
		if configs[key] != nil {
			if err := configTableWriter.Write(key, *configs[key]); err != nil {
				return err
			}
		}
	}

	if err := configTableWriter.Flush(); err != nil {
		return err
	}
	output.PrintStrings("")
	return nil
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

		if configs, err = internal.ListConfigs(&admin, brokerConfig, false); err != nil {
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

func (operation *Operation) DescribeBroker(id string, flags DescribeBrokerFlags) error {
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

	if id == "default" {
		id = ""
	}

	var broker *sarama.Broker

	for _, aBroker := range client.Brokers() {
		if strconv.Itoa(int(aBroker.ID())) == id {
			broker = aBroker
			break
		}
	}

	if id != "" && broker == nil {
		return errors.Errorf("cannot find broker with id: %s", id)
	}

	var configs []internal.Config

	brokerConfig := sarama.ConfigResource{
		Type: sarama.BrokerResource,
		Name: id,
	}

	if configs, err = internal.ListConfigs(&admin, brokerConfig, flags.AllConfigs); err != nil {
		return err
	}

	brokerInfo := Broker{Configs: configs}
	if broker != nil {
		brokerInfo.ID = broker.ID()
		brokerInfo.Address = broker.Addr()
	}

	if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
		return output.PrintObject(brokerInfo, flags.OutputFormat)
	} else if flags.OutputFormat != "" && flags.OutputFormat != "wide" {
		return errors.Errorf("unknown outputFormat: %s", flags.OutputFormat)
	}

	tableWriter := output.CreateTableWriter()

	if broker != nil {
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
	}

	// write config table
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

	return nil
}

func (operation *Operation) listBrokerIDs() ([]string, error) {

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

	var brokerIDs = make([]string, 0)

	for _, broker := range client.Brokers() {
		brokerIDs = append(brokerIDs, fmt.Sprint(broker.ID()))
	}

	return brokerIDs, nil
}

func CompleteBrokerIDs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	brokerIDs, err := (&Operation{}).listBrokerIDs()

	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return brokerIDs, cobra.ShellCompDirectiveNoFileComp
}

func FromYaml(yamlString string) (Broker, error) {
	var broker Broker
	err := yaml.Unmarshal([]byte(yamlString), &broker)
	return broker, err
}
