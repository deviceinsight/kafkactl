package alter

import (
	"github.com/deviceinsight/kafkactl/cmd/validation"
	"github.com/deviceinsight/kafkactl/operations/partitions"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
	"strconv"
)

var partitionFlags partitions.AlterPartitionFlags

var cmdAlterPartition = &cobra.Command{
	Use:   "partition TOPIC PARTITION",
	Short: "alter a topic",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		var partition int32

		if i, err := strconv.ParseInt(args[1], 10, 64); err != nil {
			output.Failf("argument 2 needs to be a partition %s", args[1])
		} else {
			partition = int32(i)
		}

		(&partitions.PartitionOperation{}).AlterPartition(args[0], partition, partitionFlags)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validation.ValidateAtLeastOneRequiredFlag(cmd)
	},
}

func init() {
	cmdAlterPartition.Flags().Int32SliceVarP(&partitionFlags.Replicas, "replicas", "r", nil, "set replicas for a partition")

	if err := validation.MarkFlagAtLeastOneRequired(cmdAlterPartition.Flags(), "replicas"); err != nil {
		output.Failf("internal error: %v", err)
	}
}
