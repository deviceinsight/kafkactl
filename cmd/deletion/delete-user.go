package deletion

import (
	"github.com/deviceinsight/kafkactl/v5/cmd/validation"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/user"
	"github.com/spf13/cobra"
)

func newDeleteUserCmd() *cobra.Command {

	var flags user.DeleteUserFlags

	var cmdDeleteUser = &cobra.Command{
		Use:   "user USERNAME",
		Short: "delete SCRAM user credentials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&user.Operation{}).DeleteUser(args[0], flags)
		},
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return validation.ValidateAtLeastOneRequiredFlag(cmd)
		},
	}

	cmdDeleteUser.Flags().StringVarP(&flags.Mechanism, "mechanism", "m", "SCRAM-SHA-256", "SCRAM mechanism to delete (SCRAM-SHA-256, SCRAM-SHA-512)")

	if err := validation.MarkFlagAtLeastOneRequired(cmdDeleteUser.Flags(), "mechanism"); err != nil {
		panic(err)
	}

	return cmdDeleteUser
}
