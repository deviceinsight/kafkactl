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

package deletion

import (
	"github.com/Shopify/sarama"
	"github.com/random-dwi/kafkactl/util/output"

	"github.com/random-dwi/kafkactl/util"
	"github.com/spf13/cobra"
)

// deleteTopicCmd represents the useContext command
var deleteTopicCmd = &cobra.Command{
	Use:   "topic",
	Short: "delete a topic",
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

		for _, topic := range args {
			if err = admin.DeleteTopic(topic); err != nil {
				output.Failf("failed to delete topic: %v", err)
			} else {
				output.Infof("topic deleted: %s", topic)
			}
		}
	},
}

func init() {
}
