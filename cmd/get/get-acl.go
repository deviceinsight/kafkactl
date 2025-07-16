package get

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/acl"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newGetACLCmd() *cobra.Command {

	var flags acl.GetACLFlags

	var cmdGetAcls = &cobra.Command{
		Use:     "access-control-list",
		Aliases: []string{"acl"},
		Short:   "list available acls",
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&acl.Operation{}).GetACL(flags)
		},
	}

	cmdGetAcls.Flags().StringVarP(&flags.Operation, "operation", "", "any", "operation of acl")
	cmdGetAcls.Flags().StringVarP(&flags.PatternType, "pattern", "", "any", "pattern type. one of (any, match, prefixed, literal)")

	cmdGetAcls.Flags().StringVarP(&flags.ResourceName, "resource-name", "r", "", "resource name of acl (e.g. topic name)")
	cmdGetAcls.Flags().StringVarP(&flags.Principal, "principal", "p", "", "principal of acl")
	cmdGetAcls.Flags().StringVarP(&flags.Host, "host", "", "", "host of acl")

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
