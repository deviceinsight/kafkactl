package get

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newGetTopicsCmd() *cobra.Command {

	var flags operations.GetTopicsFlags

	var cmdGetTopics = &cobra.Command{
		Use:   "topics",
		Short: "list available topics",
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&operations.TopicOperation{}).GetTopics(flags); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdGetTopics.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide|compact")

	return cmdGetTopics
}
