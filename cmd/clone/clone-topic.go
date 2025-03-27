package clone

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func newCloneTopicCmd() *cobra.Command {

	var cloneTopicCmd = &cobra.Command{
		Use:   "topic SOURCE_TOPIC TARGET_TOPIC",
		Short: "clone existing topic (number of partitions, replication factor, config entries) to new one",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&topic.Operation{}).CloneTopic(args[0], args[1])
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	return cloneTopicCmd
}
