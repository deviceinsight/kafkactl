package get

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/spf13/cobra"
)

var flags operations.GetTopicsFlags

var cmdGetTopics = &cobra.Command{
	Use:   "topics",
	Short: "list available topics",
	Run: func(cmd *cobra.Command, args []string) {
		(&operations.TopicOperation{}).GetTopics(flags)
	},
}

func init() {
	cmdGetTopics.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide|compact")
}
