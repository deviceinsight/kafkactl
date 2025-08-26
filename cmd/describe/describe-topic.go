package describe

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func newDescribeTopicCmd() *cobra.Command {

	var flags topic.DescribeTopicFlags

	var cmdDescribeTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "describe a topic",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&topic.Operation{}).DescribeTopic(args[0], flags)
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	cmdDescribeTopic.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide")
	cmdDescribeTopic.Flags().BoolVarP(&flags.AllConfigs, "all-configs", "a", false, "print all configs including defaults")
	cmdDescribeTopic.Flags().BoolVarP(&flags.SkipEmptyPartitions, "skip-empty", "s", false, "show only partitions that have a messages")

	return cmdDescribeTopic
}
