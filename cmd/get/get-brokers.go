package get

import (
	"github.com/deviceinsight/kafkactl/internal/broker"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/internal/output"
	"github.com/spf13/cobra"
)

func newGetBrokersCmd() *cobra.Command {

	var flags broker.GetBrokersFlags

	var cmdGetBrokers = &cobra.Command{
		Use:   "brokers",
		Short: "list brokers",
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {
				if err := (&broker.Operation{}).GetBrokers(flags); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdGetBrokers.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|compact")

	return cmdGetBrokers
}
