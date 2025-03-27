package deletion

import (
	"github.com/deviceinsight/kafkactl/v5/cmd/validation"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func newDeleteRecordsCmd() *cobra.Command {

	var flags topic.DeleteRecordsFlags

	var cmdDeleteRecords = &cobra.Command{
		Use:   "records TOPIC",
		Short: "delete a records from a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&topic.Operation{}).DeleteRecords(args[0], flags)
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdDeleteRecords.Flags().StringArrayVarP(&flags.Offsets, "offset", "", flags.Offsets, "offsets in format `partition=offset`. records with smaller offset will be deleted.")

	if err := validation.MarkFlagAtLeastOneRequired(cmdDeleteRecords.Flags(), "offset"); err != nil {
		panic(err)
	}

	return cmdDeleteRecords
}
