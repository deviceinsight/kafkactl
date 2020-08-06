package create

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newCreateTopicCmd() *cobra.Command {

	var flags operations.CreateTopicFlags

	var cmdCreateTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "create a topic",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&operations.TopicOperation{}).CreateTopics(args, flags); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdCreateTopic.Flags().Int32VarP(&flags.Partitions, "partitions", "p", 1, "number of partitions")
	cmdCreateTopic.Flags().Int16VarP(&flags.ReplicationFactor, "replication-factor", "r", 1, "replication factor")
	cmdCreateTopic.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")
	cmdCreateTopic.Flags().StringArrayVarP(&flags.Configs, "config", "c", flags.Configs, "configs in format `key=value`")

	return cmdCreateTopic
}
