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

package get

import (
	"github.com/Shopify/sarama"
	topicTools "github.com/random-dwi/kafkactl/kafka/topics"
	"github.com/random-dwi/kafkactl/util/output"

	"github.com/random-dwi/kafkactl/util"
	"github.com/spf13/cobra"
)

// getTopicsCmd represents the useContext command
var getTopicsCmd = &cobra.Command{
	Use:   "topics",
	Short: "list available topics",
	Run: func(cmd *cobra.Command, args []string) {

		context := util.CreateClientContext()

		var (
			err    error
			client sarama.Client
			topics []string
		)

		if client, err = util.CreateClient(&context); err != nil {
			output.Failf("failed to create client err=%v", err)
		}

		if topics, err = client.Topics(); err != nil {
			output.Failf("failed to read topics err=%v", err)
		}

		if outputFormat == "" {
			output.PrintStrings(topics...)
			return
		}

		for _, topic := range topics {
			var t, _ = topicTools.ReadTopic(&client, topic, true, false, true)
			output.PrintObject(t, outputFormat)
		}
	},
}

func init() {

	getTopicsCmd.Flags().StringVarP(&outputFormat, "output", "o", outputFormat, "Output format. One of: json|yaml")
}
