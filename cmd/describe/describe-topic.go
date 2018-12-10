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

package describe

import (
	"github.com/Shopify/sarama"
	topicTools "github.com/random-dwi/kafkactl/kafka/topics"
	"github.com/random-dwi/kafkactl/util"
	"github.com/random-dwi/kafkactl/util/output"
	"github.com/spf13/cobra"
)

var describeTopicCmd = &cobra.Command{
	Use:   "topic",
	Short: "describe a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		context := util.CreateClientContext()

		var topic = args[0]

		var (
			client sarama.Client
			admin  sarama.ClusterAdmin
			err    error
		)

		if admin, err = util.CreateClusterAdmin(&context); err != nil {
			output.Failf("failed to create cluster admin: %v", err)
		}

		if client, err = util.CreateClient(&context); err != nil {
			output.Failf("failed to create client err=%v", err)
		}

		var t, _ = topicTools.ReadTopic(&client, &admin, topic, true, false, true, true)
		output.PrintObject(t, "yaml")
	},
}

func init() {
}
