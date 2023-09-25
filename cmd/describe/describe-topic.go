package describe

import (
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDescribeTopicCmd() *cobra.Command {

	var flags topic.DescribeTopicFlags

	var cmdDescribeTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "describe a topic",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&topic.Operation{}).DescribeTopic(args[0], flags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdDescribeTopic.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide")
	cmdDescribeTopic.Flags().BoolVarP(&flags.PrintConfigs, "print-configs", "c", true, "print configs")
	cmdDescribeTopic.Flags().BoolVarP(&flags.SkipEmptyPartitions, "skip-empty", "s", false, "show only partitions that have a messages")

	return cmdDescribeTopic
}
