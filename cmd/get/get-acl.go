package get

import (
	"github.com/deviceinsight/kafkactl/internal/acl"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/output"
	"github.com/spf13/cobra"
)

func newGetACLCmd() *cobra.Command {

	var flags acl.GetACLFlags

	var cmdGetAcls = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "list available acls",
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&acl.Operation{}).GetACL(flags); err != nil {
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

	_ = cmdGetAcls.RegisterFlagCompletionFunc("operation", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "all", "read", "write", "create", "delete", "alter", "describe", "clusteraction", "describeconfigs", "alterconfigs", "idempotentwrite"}, cobra.ShellCompDirectiveDefault
	})

	_ = cmdGetAcls.RegisterFlagCompletionFunc("pattern", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "match", "prefixed", "literal"}, cobra.ShellCompDirectiveDefault
	})

	return cmdGetAcls
}
