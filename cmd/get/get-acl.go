package get

import (
	"github.com/deviceinsight/kafkactl/operations/acl"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newGetAclCmd() *cobra.Command {

	var flags acl.GetAclFlags

	var cmdGetAcls = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "list available acls",
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&acl.AclOperation{}).GetAcl(flags); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdGetAcls.Flags().StringVarP(&flags.Operation, "operation", "", "any", "operation of acl")
	cmdGetAcls.Flags().StringVarP(&flags.PatternType, "pattern", "", "any", "pattern type. one of (any, match, prefixed, literal)")

	// specify permissionType
	cmdGetAcls.Flags().BoolVarP(&flags.Allow, "allow", "a", false, "acl of permissionType 'allow'")
	cmdGetAcls.Flags().BoolVarP(&flags.Deny, "deny", "d", false, "acl of permissionType 'deny'")

	// specify resource type
	cmdGetAcls.Flags().BoolVarP(&flags.Topics, "topics", "t", false, "list acl for topics")
	cmdGetAcls.Flags().BoolVarP(&flags.Groups, "groups", "g", false, "list acl for consumer groups")
	cmdGetAcls.Flags().BoolVarP(&flags.Cluster, "cluster", "c", false, "list acl for the cluster")

	cmdGetAcls.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml")

	_ = cmdGetAcls.RegisterFlagCompletionFunc("operation", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "all", "read", "write", "create", "delete", "alter", "describe", "clusteraction", "describeconfigs", "alterconfigs", "idempotentwrite"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdGetAcls.RegisterFlagCompletionFunc("pattern", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "match", "prefixed", "literal"}, cobra.ShellCompDirectiveDefault
	})

	return cmdGetAcls
}
