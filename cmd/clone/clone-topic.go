package clone

import (
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newCloneTopicCmd() *cobra.Command {

	var cloneTopicCmd = &cobra.Command{
		Use:   "topic SOURCE_TOPIC TARGET_TOPIC",
		Short: "clone existing topic (number of partitions, replication factor, config entries) to new one",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.Operation{}).TryRun(cmd, args) {
				if err := (&topic.Operation{}).CloneTopic(args[0], args[1]); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	return cloneTopicCmd
}
