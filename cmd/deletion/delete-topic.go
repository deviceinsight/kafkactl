package deletion

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDeleteTopicCmd() *cobra.Command {

	var cmdDeleteTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "delete a topic",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := (&operations.TopicOperation{}).DeleteTopics(args); err != nil {
				output.Fail(err)
			}
		},
		ValidArgsFunction: operations.CompleteTopicNames,
	}

	return cmdDeleteTopic
}
