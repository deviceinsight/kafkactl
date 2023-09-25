package deletion

import (
	"github.com/deviceinsight/kafkactl/internal/acl"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDeleteACLCmd() *cobra.Command {

	var flags acl.DeleteACLFlags

	var cmdDeleteACL = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "delete an acl",
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&acl.Operation{}).DeleteACL(flags); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdDeleteACL.Flags().StringVarP(&flags.Operation, "operation", "o", "", "operation of acl")
	cmdDeleteACL.Flags().StringVarP(&flags.PatternType, "pattern", "", "", "pattern type. one of (any, match, prefixed, literal)")

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

	_ = cmdDeleteACL.RegisterFlagCompletionFunc("operation", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "all", "read", "write", "create", "delete", "alter", "describe", "clusteraction", "describeconfigs", "alterconfigs", "idempotentwrite"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdDeleteACL.RegisterFlagCompletionFunc("pattern", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "match", "prefixed", "literal"}, cobra.ShellCompDirectiveDefault
	})

	return cmdDeleteACL
}
