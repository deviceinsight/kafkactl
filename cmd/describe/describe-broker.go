package describe

import (
	"github.com/deviceinsight/kafkactl/v5/internal"

	"github.com/deviceinsight/kafkactl/v5/internal/broker"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newDescribeBrokerCmd() *cobra.Command {

	var flags broker.DescribeBrokerFlags

	var cmdDescribeBroker = &cobra.Command{
		Use:   "broker ID",
		Short: "describe a broker",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&broker.Operation{}).DescribeBroker(args[0], flags)
		},
		ValidArgsFunction: broker.CompleteBrokerIDs,
	}

	cmdDescribeBroker.Flags().BoolVarP(&flags.AllConfigs, "all-configs", "a", false, "print all configs including defaults")
	cmdDescribeBroker.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide")

	return cmdDescribeBroker
}
