package create

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func newCreateTopicCmd() *cobra.Command {

	var flags topic.CreateTopicFlags

	var cmdCreateTopic = &cobra.Command{
		Use:   "topic TOPIC",
		Short: "create a topic",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&topic.Operation{}).CreateTopics(args, flags)
		},
	}

	cmdCreateTopic.Flags().Int32VarP(&flags.Partitions, "partitions", "p", 1, "number of partitions")
	cmdCreateTopic.Flags().Int16VarP(&flags.ReplicationFactor, "replication-factor", "r", -1, "replication factor")
	cmdCreateTopic.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")
	cmdCreateTopic.Flags().StringVarP(&flags.File, "file", "f", "", "file with topic description")
	cmdCreateTopic.Flags().StringArrayVarP(&flags.Configs, "config", "c", flags.Configs, "configs in format `key=value`")

	return cmdCreateTopic
}
