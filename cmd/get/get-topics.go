package get

import (
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/output"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/spf13/cobra"
)

func newGetTopicsCmd() *cobra.Command {

	var flags topic.GetTopicsFlags

	var cmdGetTopics = &cobra.Command{
		Use:   "topics",
		Short: "list available topics",
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&topic.Operation{}).GetTopics(flags); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdGetTopics.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide|compact")

	return cmdGetTopics
}
