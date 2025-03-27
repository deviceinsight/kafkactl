package deletion

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func newDeleteTopicCmd() *cobra.Command {

	var cmdDeleteTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "delete a topic",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&topic.Operation{}).DeleteTopics(args)
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	return cmdDeleteTopic
}
