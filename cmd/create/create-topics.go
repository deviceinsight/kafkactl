// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package create

import (
	"github.com/Shopify/sarama"
	"github.com/random-dwi/kafkactl/util/output"
	"strings"

	"github.com/random-dwi/kafkactl/util"
	"github.com/spf13/cobra"
)

type CreateTopicFlags struct {
	partitions        int32
	replicationFactor int16
	validateOnly      bool
	configs           []string
}

var flags CreateTopicFlags

// createTopicCmd represents the useContext command
var createTopicCmd = &cobra.Command{
	Use:   "topic",
	Short: "create a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		context := util.CreateClientContext()

		var (
			err   error
			admin sarama.ClusterAdmin
		)

		if admin, err = util.CreateClusterAdmin(&context); err != nil {
			output.Failf("failed to create cluster admin: %v", err)
		}

		topicDetails := sarama.TopicDetail{
			NumPartitions:     flags.partitions,
			ReplicationFactor: flags.replicationFactor,
			ConfigEntries:     map[string]*string{},
		}

		for _, config := range flags.configs {
			configParts := strings.Split(config, "=")
			topicDetails.ConfigEntries[configParts[0]] = &configParts[1]
		}

		for _, topic := range args {
			if err = admin.CreateTopic(topic, &topicDetails, flags.validateOnly); err != nil {
				output.Failf("failed to create topic: %v", err)
			} else {
				output.Infof("topic created: %s", topic)
			}
		}
	},
}

func init() {
	createTopicCmd.Flags().Int32VarP(&flags.partitions, "partitions", "p", 1, "number of partitions")
	createTopicCmd.Flags().Int16VarP(&flags.replicationFactor, "replication-factor", "r", 1, "replication factor")
	createTopicCmd.Flags().BoolVarP(&flags.validateOnly, "validate-only", "V", false, "validate only")
	createTopicCmd.Flags().StringArrayVarP(&flags.configs, "config", "C", flags.configs, "configs in format `key=value`")
}
