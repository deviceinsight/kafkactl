package get

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/user"
	"github.com/spf13/cobra"
)

func newGetUsersCmd() *cobra.Command {

	var flags user.GetUsersFlags

	var cmdGetUsers = &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "get SCRAM users",
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&user.Operation{}).GetUsers(flags)
		},
	}

	cmdGetUsers.Flags().StringVarP(&flags.OutputFormat, "output", "o", "", "output format. One of: json|yaml")

	return cmdGetUsers
}
