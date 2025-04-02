package reset

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroupoffsets"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newResetOffsetCmd() *cobra.Command {

	var offsetFlags consumergroupoffsets.ResetConsumerGroupOffsetFlags

	var cmdResetOffset = &cobra.Command{
		Use:     "consumer-group-offset GROUP",
		Aliases: []string{"cgo", "offset"},
		Short:   "reset a consumer group offset",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&consumergroupoffsets.ConsumerGroupOffsetOperation{}).ResetConsumerGroupOffset(offsetFlags, args[0])
		},
		ValidArgsFunction: consumergroups.CompleteConsumerGroups,
	}

	cmdResetOffset.Flags().BoolVarP(&offsetFlags.OldestOffset, "oldest", "", false, "set the offset to oldest offset (for all partitions or the specified partition)")
	cmdResetOffset.Flags().BoolVarP(&offsetFlags.NewestOffset, "newest", "", false, "set the offset to newest offset (for all partitions or the specified partition)")
	cmdResetOffset.Flags().BoolVarP(&offsetFlags.AllTopics, "all-topics", "", false, "do the operation for all topics in the consumer group")
	cmdResetOffset.Flags().Int64VarP(&offsetFlags.Offset, "offset", "", -1, "set offset to this value. offset with value -1 is ignored")
	cmdResetOffset.Flags().Int32VarP(&offsetFlags.Partition, "partition", "p", -1, "partition to apply the offset. -1 stands for all partitions")
	cmdResetOffset.Flags().StringArrayVarP(&offsetFlags.Topic, "topic", "t", offsetFlags.Topic, "one ore more topics to change offset for")
	cmdResetOffset.Flags().BoolVarP(&offsetFlags.Execute, "execute", "e", false, "execute the reset (as default only the results are displayed for validation)")
	cmdResetOffset.Flags().StringVarP(&offsetFlags.OutputFormat, "output", "o", offsetFlags.OutputFormat, "output format. One of: json|yaml")
	cmdResetOffset.Flags().StringVarP(&offsetFlags.ToDatetime, "to-datetime", "", "", "set the offset to offset of given timestamp")

	return cmdResetOffset
}
