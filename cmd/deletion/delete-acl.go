package deletion

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/acl"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newDeleteACLCmd() *cobra.Command {

	var flags acl.DeleteACLFlags

	var cmdDeleteACL = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "delete an acl",
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&acl.Operation{}).DeleteACL(flags)
		},
	}

	cmdDeleteACL.Flags().StringVarP(&flags.Operation, "operation", "o", "", "operation of acl")
	cmdDeleteACL.Flags().StringVarP(&flags.PatternType, "pattern", "", "", "pattern type. one of (any, match, prefixed, literal)")

	cmdDeleteACL.Flags().StringVarP(&flags.Principal, "principal", "p", "", "principal of acl")
	cmdDeleteACL.Flags().StringVarP(&flags.Host, "host", "", "", "host of acl")

	// specify permissionType
	cmdDeleteACL.Flags().BoolVarP(&flags.Allow, "allow", "a", false, "acl of permissionType 'allow'")
	cmdDeleteACL.Flags().BoolVarP(&flags.Deny, "deny", "d", false, "acl of permissionType 'deny'")

	// specify resource type
	cmdDeleteACL.Flags().BoolVarP(&flags.Topics, "topics", "t", false, "delete acl for a topic")
	cmdDeleteACL.Flags().BoolVarP(&flags.Groups, "groups", "g", false, "delete acl for a consumer group")
	cmdDeleteACL.Flags().BoolVarP(&flags.Cluster, "cluster", "c", false, "delete acl for the cluster")

	cmdDeleteACL.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	_ = cmdDeleteACL.MarkFlagRequired("operation")
	_ = cmdDeleteACL.MarkFlagRequired("pattern")

	_ = cmdDeleteACL.RegisterFlagCompletionFunc("operation", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "all", "read", "write", "create", "delete", "alter", "describe", "clusteraction", "describeconfigs", "alterconfigs", "idempotentwrite"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdDeleteACL.RegisterFlagCompletionFunc("pattern", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "match", "prefixed", "literal"}, cobra.ShellCompDirectiveDefault
	})

	return cmdDeleteACL
}
