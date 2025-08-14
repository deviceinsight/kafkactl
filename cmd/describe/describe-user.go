package describe

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/user"
	"github.com/spf13/cobra"
)

func newDescribeUserCmd() *cobra.Command {

	var flags user.DescribeUserFlags

	var cmdDescribeUser = &cobra.Command{
		Use:   "user USERNAME",
		Short: "describe a SCRAM user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&user.Operation{}).DescribeUser(args[0], flags)
		},
	}

	cmdDescribeUser.Flags().StringVarP(&flags.OutputFormat, "output", "o", "", "output format. One of: json|yaml")

	return cmdDescribeUser
}
