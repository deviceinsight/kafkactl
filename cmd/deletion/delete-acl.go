package deletion

import (
	"github.com/deviceinsight/kafkactl/operations/acl"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDeleteAclCmd() *cobra.Command {

	var flags acl.DeleteAclFlags

	var cmdDeleteAcl = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "delete an acl",
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := (&acl.AclOperation{}).DeleteAcl(flags); err != nil {
				output.Fail(err)
			}
		},
	}

	cmdDeleteAcl.Flags().StringVarP(&flags.Operation, "operation", "o", "", "operation of acl")
	cmdDeleteAcl.Flags().StringVarP(&flags.PatternType, "pattern", "", "", "pattern type. one of (any, match, prefixed, literal)")

	// specify permissionType
	cmdDeleteAcl.Flags().BoolVarP(&flags.Allow, "allow", "a", false, "acl of permissionType 'allow'")
	cmdDeleteAcl.Flags().BoolVarP(&flags.Deny, "deny", "d", false, "acl of permissionType 'deny'")

	// specify resource type
	cmdDeleteAcl.Flags().BoolVarP(&flags.Topics, "topics", "t", false, "delete acl for a topic")
	cmdDeleteAcl.Flags().BoolVarP(&flags.Groups, "groups", "g", false, "delete acl for a consumer group")
	cmdDeleteAcl.Flags().BoolVarP(&flags.Cluster, "cluster", "c", false, "delete acl for the cluster")

	cmdDeleteAcl.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	_ = cmdDeleteAcl.MarkFlagRequired("operation")
	_ = cmdDeleteAcl.MarkFlagRequired("pattern")

	_ = cmdDeleteAcl.RegisterFlagCompletionFunc("operation", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "all", "read", "write", "create", "delete", "alter", "describe", "clusteraction", "describeconfigs", "alterconfigs", "idempotentwrite"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdDeleteAcl.RegisterFlagCompletionFunc("pattern", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "match", "prefixed", "literal"}, cobra.ShellCompDirectiveDefault
	})

	return cmdDeleteAcl
}
