package get

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/broker"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func newGetBrokersCmd() *cobra.Command {

	var flags broker.GetBrokersFlags

	var cmdGetBrokers = &cobra.Command{
		Use:   "brokers",
		Short: "list brokers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&broker.Operation{}).GetBrokers(flags)
		},
	}

	cmdGetBrokers.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|compact")

	return cmdGetBrokers
}
