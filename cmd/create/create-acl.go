package create

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/operations/acl"
	"github.com/deviceinsight/kafkactl/operations/consumergroups"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newCreateAclCmd() *cobra.Command {

	var flags acl.CreateAclFlags

	var cmdCreateAcl = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "create an acl",
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&acl.AclOperation{}).CreateAcl(flags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: acl.CompleteCreateAcl,
	}

	cmdCreateAcl.Flags().StringVarP(&flags.Principal, "principal", "p", "", "principal to be authenticated")
	cmdCreateAcl.Flags().StringArrayVarP(&flags.Hosts, "host", "", nil, "hosts to allow")
	cmdCreateAcl.Flags().StringArrayVarP(&flags.Operations, "operation", "o", nil, "operations of acl")
	cmdCreateAcl.Flags().StringVarP(&flags.PatternType, "pattern", "", "literal", "pattern type. one of (match, prefixed, literal)")

	// specify permissionType
	cmdCreateAcl.Flags().BoolVarP(&flags.Allow, "allow", "a", false, "acl of permissionType 'allow' (choose this or 'deny')")
	cmdCreateAcl.Flags().BoolVarP(&flags.Deny, "deny", "d", false, "acl of permissionType 'deny' (choose this or 'allow')")

	// specify resource type
	cmdCreateAcl.Flags().StringVarP(&flags.Topic, "topic", "t", "", "create acl for a topic")
	cmdCreateAcl.Flags().StringVarP(&flags.Group, "group", "g", "", "create acl for a consumer group")
	cmdCreateAcl.Flags().BoolVarP(&flags.Cluster, "cluster", "c", false, "create acl for the cluster")

	cmdCreateAcl.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	_ = cmdCreateAcl.MarkFlagRequired("principal")
	_ = cmdCreateAcl.MarkFlagRequired("operation")

	_ = cmdCreateAcl.RegisterFlagCompletionFunc("pattern", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"match", "prefixed", "literal"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdCreateAcl.RegisterFlagCompletionFunc("operation", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "all", "read", "write", "create", "delete", "alter", "describe", "clusteraction", "describeconfigs", "alterconfigs", "idempotentwrite"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdCreateAcl.RegisterFlagCompletionFunc("topic", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return operations.CompleteTopicNames(cmd, args, toComplete)
	})

	_ = cmdCreateAcl.RegisterFlagCompletionFunc("group", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return consumergroups.CompleteConsumerGroups(cmd, args, toComplete)
	})

	return cmdCreateAcl
}
