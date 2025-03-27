package alter

import (
	"strconv"

	"github.com/deviceinsight/kafkactl/v5/internal"

	"github.com/deviceinsight/kafkactl/v5/cmd/validation"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/partition"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newAlterPartitionCmd() *cobra.Command {

	var flags partition.AlterPartitionFlags

	var cmdAlterPartition = &cobra.Command{
		Use:   "partition TOPIC PARTITION",
		Short: "alter a partition",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}

			var (
				partitionID int64
				err         error
			)

			if partitionID, err = strconv.ParseInt(args[1], 10, 64); err != nil {
				return errors.Errorf("argument 2 needs to be a partition %s", args[1])
			}

			return (&partition.Operation{}).AlterPartition(args[0], int32(partitionID), flags)
		},
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return validation.ValidateAtLeastOneRequiredFlag(cmd)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return topic.CompleteTopicNames(cmd, args, toComplete)
			} else if len(args) == 1 {
				return partition.CompletePartitionIDs(cmd, args, toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmdAlterPartition.Flags().Int32SliceVarP(&flags.Replicas, "replicas", "r", nil, "set replicas for a partition")
	cmdAlterPartition.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	if err := validation.MarkFlagAtLeastOneRequired(cmdAlterPartition.Flags(), "replicas"); err != nil {
		panic(err)
	}
	return cmdAlterPartition
}
