package create

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/acl"
	"github.com/deviceinsight/kafkactl/v5/internal/consumergroups"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"
	"github.com/spf13/cobra"
)

func newCreateACLCmd() *cobra.Command {

	var flags acl.CreateACLFlags

	var cmdCreateACL = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "create an acl",
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&acl.Operation{}).CreateACL(flags)
		},
		ValidArgsFunction: acl.CompleteCreateACL,
	}

	cmdCreateACL.Flags().StringVarP(&flags.Principal, "principal", "p", "", "principal to be authenticated")
	cmdCreateACL.Flags().StringArrayVarP(&flags.Hosts, "host", "", nil, "hosts to allow")
	cmdCreateACL.Flags().StringArrayVarP(&flags.Operations, "operation", "o", nil, "operations of acl")
	cmdCreateACL.Flags().StringVarP(&flags.PatternType, "pattern", "", "literal", "pattern type. one of (match, prefixed, literal)")

	// specify permissionType
	cmdCreateACL.Flags().BoolVarP(&flags.Allow, "allow", "a", false, "acl of permissionType 'allow' (choose this or 'deny')")
	cmdCreateACL.Flags().BoolVarP(&flags.Deny, "deny", "d", false, "acl of permissionType 'deny' (choose this or 'allow')")

	// specify resource type
	cmdCreateACL.Flags().StringVarP(&flags.Topic, "topic", "t", "", "create acl for a topic")
	cmdCreateACL.Flags().StringVarP(&flags.Group, "group", "g", "", "create acl for a consumer group")
	cmdCreateACL.Flags().BoolVarP(&flags.Cluster, "cluster", "c", false, "create acl for the cluster")

	cmdCreateACL.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	_ = cmdCreateACL.MarkFlagRequired("principal")
	_ = cmdCreateACL.MarkFlagRequired("operation")

	_ = cmdCreateACL.RegisterFlagCompletionFunc("pattern", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"match", "prefixed", "literal"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdCreateACL.RegisterFlagCompletionFunc("operation", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "all", "read", "write", "create", "delete", "alter", "describe", "clusteraction", "describeconfigs", "alterconfigs", "idempotentwrite"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdCreateACL.RegisterFlagCompletionFunc("topic", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return topic.CompleteTopicNames(cmd, args, toComplete)
	})

	_ = cmdCreateACL.RegisterFlagCompletionFunc("group", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return consumergroups.CompleteConsumerGroups(cmd, args, toComplete)
	})

	return cmdCreateACL
}
