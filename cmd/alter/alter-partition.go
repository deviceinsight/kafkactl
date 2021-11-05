package alter

import (
	"strconv"

	"github.com/deviceinsight/kafkactl/cmd/validation"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/partition"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newAlterPartitionCmd() *cobra.Command {

	var flags partition.AlterPartitionFlags

	var cmdAlterPartition = &cobra.Command{
		Use:   "partition TOPIC PARTITION",
		Short: "alter a partition",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.Operation{}).TryRun(cmd, args) {

				var partitionID int32

				if i, err := strconv.ParseInt(args[1], 10, 64); err != nil {
					output.Fail(errors.Errorf("argument 2 needs to be a partition %s", args[1]))
				} else {
					partitionID = int32(i)
				}

				if err := (&partition.Operation{}).AlterPartition(args[0], partitionID, flags); err != nil {
					output.Fail(err)
				}
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validation.ValidateAtLeastOneRequiredFlag(cmd)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return topic.CompleteTopicNames(cmd, args, toComplete)
			} else if len(args) == 1 {
				return partition.CompletePartitionIds(cmd, args, toComplete)
			} else {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
	}

	cmdAlterPartition.Flags().Int32SliceVarP(&flags.Replicas, "replicas", "r", nil, "set replicas for a partition")
	cmdAlterPartition.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	if err := validation.MarkFlagAtLeastOneRequired(cmdAlterPartition.Flags(), "replicas"); err != nil {
		panic(err)
	}
	return cmdAlterPartition
}
