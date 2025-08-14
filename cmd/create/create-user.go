package create

import (
	"github.com/deviceinsight/kafkactl/v5/cmd/validation"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/user"
	"github.com/spf13/cobra"
)

func newCreateUserCmd() *cobra.Command {

	var flags user.CreateUserFlags

	var cmdCreateUser = &cobra.Command{
		Use:   "user USERNAME",
		Short: "create a SCRAM user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&user.Operation{}).CreateUser(args[0], flags)
		},
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return validation.ValidateAtLeastOneRequiredFlag(cmd)
		},
	}

	cmdCreateUser.Flags().StringVarP(&flags.Mechanism, "mechanism", "m", "SCRAM-SHA-256", "SCRAM mechanism (SCRAM-SHA-256, SCRAM-SHA-512)")
	cmdCreateUser.Flags().StringVarP(&flags.Password, "password", "p", "", "user password")
	cmdCreateUser.Flags().StringVarP(&flags.Salt, "salt", "s", "", "custom salt (base64 encoded, generated if not provided)")
	cmdCreateUser.Flags().Int32VarP(&flags.Iterations, "iterations", "i", 4096, "SCRAM iterations")

	if err := validation.MarkFlagAtLeastOneRequired(cmdCreateUser.Flags(), "password"); err != nil {
		panic(err)
	}

	return cmdCreateUser
}
