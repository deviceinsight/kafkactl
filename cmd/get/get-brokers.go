package get

import (
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newGetBrokersCmd() *cobra.Command {

	var flags operations.GetBrokersFlags

	var cmdGetBrokers = &cobra.Command{
		Use:   "brokers",
		Short: "list brokers",
		Run: func(cmd *cobra.Command, args []string) {
			if !(&k8s.K8sOperation{}).TryRun(cmd, args) {
				if err := (&operations.BrokerOperation{}).GetBrokers(flags); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdGetBrokers.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|compact")

	return cmdGetBrokers
}
