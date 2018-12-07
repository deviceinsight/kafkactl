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
	"fmt"
	"github.com/Shopify/sarama"
	topicTools "github.com/random-dwi/kafkactl/kafka/topics"
	"github.com/random-dwi/kafkactl/util/output"
	"os"
	"strings"
	"text/tabwriter"

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
			admin  sarama.ClusterAdmin
			topics []string
		)

		tabwriter := new(tabwriter.Writer)

		if admin, err = util.CreateClusterAdmin(&context); err != nil {
			output.Failf("failed to create cluster admin: %v", err)
		}

		if client, err = util.CreateClient(&context); err != nil {
			output.Failf("failed to create client err=%v", err)
		}

		if topics, err = client.Topics(); err != nil {
			output.Failf("failed to read topics err=%v", err)
		}

		if outputFormat == "" {
			tabwriter.Init(os.Stdout, 0, 0, 5, ' ', 0)
			fmt.Fprintln(tabwriter, "TOPIC\tPARTITIONS")
		} else if outputFormat == "compact" {
			output.PrintStrings(topics...)
			return
		} else if outputFormat == "wide" {
			tabwriter.Init(os.Stdout, 0, 0, 5, ' ', 0)
			fmt.Fprintln(tabwriter, "TOPIC\tPARTITIONS\tCONFIGS")
		}

		for _, topic := range topics {
			var t, _ = topicTools.ReadTopic(&client, &admin, topic, true, false, true, true)
			if outputFormat == "json" || outputFormat == "yaml" {
				output.PrintObject(t, outputFormat)
			} else if outputFormat == "wide" {
				printTable(tabwriter, t, true)
			} else {
				printTable(tabwriter, t, false)
			}
		}

		if outputFormat == "wide" || outputFormat == "" {
			tabwriter.Flush()
		}
	},
}

func printTable(tabwriter *tabwriter.Writer, topic topicTools.Topic, wide bool) {

	var configStrings []string

	for _, config := range topic.Configs {
		configStrings = append(configStrings, fmt.Sprintf("%s=%s", config.Name, config.Value))
	}

	var config = strings.Trim(strings.Join(configStrings, ","), "[]")

	if wide {
		fmt.Fprintln(tabwriter, fmt.Sprintf("%s\t%d\t%s", topic.Name, len(topic.Partitions), config))
	} else {
		fmt.Fprintln(tabwriter, fmt.Sprintf("%s\t%d", topic.Name, len(topic.Partitions)))
	}
}

func init() {

	getTopicsCmd.Flags().StringVarP(&outputFormat, "output", "o", outputFormat, "Output format. One of: json|yaml|wide|compact")
}
