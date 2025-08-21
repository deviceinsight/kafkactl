package alter

import (
	"github.com/deviceinsight/kafkactl/v5/cmd/validation"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/broker"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newAlterBrokerCmd() *cobra.Command {

	var flags broker.AlterBrokerFlags

	var cmdAlterBroker = &cobra.Command{
		Use:   "broker BROKER",
		Short: "alter a broker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&broker.Operation{}).AlterBroker(args[0], flags)
		},
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return validation.ValidateAtLeastOneRequiredFlag(cmd)
		},
		ValidArgsFunction: broker.CompleteBrokerIDs,
	}

	cmdAlterBroker.Flags().StringArrayVarP(&flags.Configs, "config", "c", flags.Configs, "configs in format `key=value`")
	cmdAlterBroker.Flags().BoolVarP(&flags.ValidateOnly, "validate-only", "v", false, "validate only")

	if err := validation.MarkFlagAtLeastOneRequired(cmdAlterBroker.Flags(), "config"); err != nil {
		panic(err)
	}

	return cmdAlterBroker
}
