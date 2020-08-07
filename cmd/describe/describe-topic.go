package describe

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDescribeTopicCmd() *cobra.Command {

	var flags operations.DescribeTopicFlags

	var cmdDescribeTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "describe a topic",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&operations.TopicOperation{}).DescribeTopic(args[0], flags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: operations.CompleteTopicNames,
	}

	cmdDescribeTopic.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide")
	cmdDescribeTopic.Flags().BoolVarP(&flags.PrintConfigs, "print-configs", "c", true, "print configs")

	return cmdDescribeTopic
}
