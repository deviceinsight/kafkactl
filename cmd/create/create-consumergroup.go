package create

import (
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroupoffsets"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

var cgFlags consumergroupoffsets.ResetConsumerGroupOffsetFlags

func newCreateConsumerGroupCmd() *cobra.Command {
	var cmdCreateConsumerGroup = &cobra.Command{
		Use:     "consumer-group GROUP",
		Aliases: []string{"cg"},
		Short:   "create a consumerGroup",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&consumergroupoffsets.ConsumerGroupOffsetOperation{}).CreateConsumerGroup(cgFlags, args[0]); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdCreateConsumerGroup.Flags().BoolVarP(&cgFlags.OldestOffset, "oldest", "", false, "set the offset to oldest offset (for all partitions or the specified partition)")
	cmdCreateConsumerGroup.Flags().BoolVarP(&cgFlags.NewestOffset, "newest", "", false, "set the offset to newest offset (for all partitions or the specified partition)")
	cmdCreateConsumerGroup.Flags().Int64VarP(&cgFlags.Offset, "offset", "", -1, "set offset to this value. offset with value -1 is ignored")
	cmdCreateConsumerGroup.Flags().Int32VarP(&cgFlags.Partition, "partition", "p", -1, "partition to create group for. -1 stands for all partitions")
	cmdCreateConsumerGroup.Flags().StringArrayVarP(&cgFlags.Topic, "topic", "t", cgFlags.Topic, "one or more topics to create group for")

	return cmdCreateConsumerGroup
}
