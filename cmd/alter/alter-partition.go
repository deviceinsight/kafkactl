package alter

import (
	"github.com/deviceinsight/kafkactl/cmd/validation"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/operations/partitions"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

func newAlterPartitionCmd() *cobra.Command {

	var flags partitions.AlterPartitionFlags

	var cmdAlterPartition = &cobra.Command{
		Use:   "partition TOPIC PARTITION",
		Short: "alter a partition",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {

				var partition int32

				if i, err := strconv.ParseInt(args[1], 10, 64); err != nil {
					output.Fail(errors.Errorf("argument 2 needs to be a partition %s", args[1]))
				} else {
					partition = int32(i)
				}

				if err := (&partitions.PartitionOperation{}).AlterPartition(args[0], partition, flags); err != nil {
					output.Fail(err)
				}
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validation.ValidateAtLeastOneRequiredFlag(cmd)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return operations.CompleteTopicNames(cmd, args, toComplete)
			} else if len(args) == 1 {
				return partitions.CompletePartitionIds(cmd, args, toComplete)
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
