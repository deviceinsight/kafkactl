package deletion

import (
	"github.com/deviceinsight/kafkactl/cmd/validation"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDeleteRecordsCmd() *cobra.Command {

	var flags topic.DeleteRecordsFlags

	var cmdDeleteRecords = &cobra.Command{
		Use:   "records TOPIC",
		Short: "delete a records from a topic",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&topic.Operation{}).DeleteRecords(args[0], flags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdDeleteRecords.Flags().StringArrayVarP(&flags.Offsets, "offset", "", flags.Offsets, "offsets in format `partition=offset`. records with smaller offset will be deleted.")

	if err := validation.MarkFlagAtLeastOneRequired(cmdDeleteRecords.Flags(), "offset"); err != nil {
		panic(err)
	}

	return cmdDeleteRecords
}
