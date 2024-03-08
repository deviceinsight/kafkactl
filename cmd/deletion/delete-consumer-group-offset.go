package deletion

import (
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroupoffsets"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

func newDeleteConsumerGroupOffsetCmd() *cobra.Command {

	var offsetFlags consumergroupoffsets.DeleteConsumerGroupOffsetFlags

	var cmdDeleteConsumerGroup = &cobra.Command{
		Use:     "consumer-group-offset CONSUMER-GROUP --topic=TOPIC --partition=PARTITION",
		Aliases: []string{"cgo", "offset"},
		Short:   "delete a consumer-group-offset",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&consumergroupoffsets.ConsumerGroupOffsetOperation{}).DeleteConsumerGroupOffset(args[0], offsetFlags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: consumergroups.CompleteConsumerGroups,
	}

	cmdDeleteConsumerGroup.Flags().Int32VarP(&offsetFlags.Partition, "partition", "p", -1, "delete offset for this partition. -1 stands for all partitions")
	cmdDeleteConsumerGroup.Flags().StringVarP(&offsetFlags.Topic, "topic", "t", offsetFlags.Topic, "delete offset for this topic")

	return cmdDeleteConsumerGroup
}
