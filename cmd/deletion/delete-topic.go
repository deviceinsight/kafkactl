package deletion

import (
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/output"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/spf13/cobra"
)

func newDeleteTopicCmd() *cobra.Command {

	var cmdDeleteTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "delete a topic",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&topic.Operation{}).DeleteTopics(args); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: topic.CompleteTopicNames,
	}

	return cmdDeleteTopic
}
