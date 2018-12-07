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

package consume

import (
	consumerTools "github.com/random-dwi/kafkactl/kafka/consumer"
	"github.com/random-dwi/kafkactl/util"
	"github.com/random-dwi/kafkactl/util/output"
	"github.com/spf13/cobra"
)

var flags consumerTools.ConsumerFlags

// ConsumeCmd represents the consume command
var ConsumeCmd = &cobra.Command{
	Use:   "consume",
	Short: "consume messages from a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {

		context := util.CreateClientContext()

		var topic = args[0]

		consumerContext := consumerTools.CreateConsumerContext(&context, topic, flags)

		partitions := consumerContext.FindPartitions()
		if len(partitions) == 0 {
			output.Failf("Found no partitions to consume")
		}

		defer consumerContext.Close()

		consumerContext.Consume(partitions)
	},
}

func init() {

	ConsumeCmd.Flags().BoolVarP(&flags.PrintKeys, "print-keys", "k", false, "print message printKeys")
	ConsumeCmd.Flags().BoolVarP(&flags.PrintTimestamps, "print-timestamps", "t", false, "print message printTimestamps")
	ConsumeCmd.Flags().StringVarP(&flags.ConsumerGroup, "consumer-group", "g", "", "consumer group to join")
	ConsumeCmd.Flags().StringArrayVarP(&flags.Offsets, "offset", "f", flags.Offsets, "define offsets for consumer")

	ConsumeCmd.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "Output format. One of: json|yaml")
}
