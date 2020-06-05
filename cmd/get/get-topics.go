package get

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

var flags operations.GetTopicsFlags

func newGetTopicsCmd() *cobra.Command {

	var cmdGetTopics = &cobra.Command{
		Use:   "topics",
		Short: "list available topics",
		Run: func(cmd *cobra.Command, args []string) {
			if err := (&operations.TopicOperation{}).GetTopics(flags); err != nil {
				output.Fail(err)
			}
		},
	}

	cmdGetTopics.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide|compact")

	return cmdGetTopics
}
