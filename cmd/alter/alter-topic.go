package alter

import (
	"github.com/deviceinsight/kafkactl/cmd/validation"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/spf13/cobra"
)

var flags operations.AlterTopicFlags

var cmdAlterTopic = &cobra.Command{
	Use:   "topic",
	Short: "alter a topic",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		(&operations.TopicOperation{}).AlterTopic(args[0], flags)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validation.ValidateAtLeastOneRequiredFlag(cmd)
	},
}

func init() {
	cmdAlterTopic.Flags().Int32VarP(&flags.Partitions, "partitions", "p", flags.Partitions, "number of partitions")
	cmdAlterTopic.Flags().StringArrayVarP(&flags.Configs, "config", "c", flags.Configs, "configs in format `key=value`")
	cmdAlterTopic.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	validation.MarkFlagAtLeastOneRequired(cmdAlterTopic.Flags(), "partitions")
	validation.MarkFlagAtLeastOneRequired(cmdAlterTopic.Flags(), "config")
}
