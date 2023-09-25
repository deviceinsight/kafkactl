package alter

import (
	"github.com/deviceinsight/kafkactl/cmd/validation"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newAlterTopicCmd() *cobra.Command {

	var flags topic.AlterTopicFlags

	var cmdAlterTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "alter a topic",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&topic.Operation{}).AlterTopic(args[0], flags); err != nil {
					output.Fail(err)
				}
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validation.ValidateAtLeastOneRequiredFlag(cmd)
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdAlterTopic.Flags().Int32VarP(&flags.Partitions, "partitions", "p", flags.Partitions, "number of partitions")
	cmdAlterTopic.Flags().Int16VarP(&flags.ReplicationFactor, "replication-factor", "r", flags.ReplicationFactor, "replication factor")
	cmdAlterTopic.Flags().StringArrayVarP(&flags.Configs, "config", "c", flags.Configs, "configs in format `key=value`")
	cmdAlterTopic.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	if err := validation.MarkFlagAtLeastOneRequired(cmdAlterTopic.Flags(), "partitions"); err != nil {
		panic(err)
	}
	if err := validation.MarkFlagAtLeastOneRequired(cmdAlterTopic.Flags(), "config"); err != nil {
		panic(err)
	}
	if err := validation.MarkFlagAtLeastOneRequired(cmdAlterTopic.Flags(), "replication-factor"); err != nil {
		panic(err)
	}

	return cmdAlterTopic
}
