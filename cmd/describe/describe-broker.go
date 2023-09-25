package describe

import (
	"strconv"

	"github.com/deviceinsight/kafkactl/internal/broker"
	"github.com/deviceinsight/kafkactl/internal/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newDescribeBrokerCmd() *cobra.Command {

	var flags broker.DescribeBrokerFlags

	var cmdDescribeBroker = &cobra.Command{
		Use:   "broker ID",
		Short: "describe a broker",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !k8s.NewOperation().TryRun(cmd, args) {

				id, err := strconv.ParseInt(args[0], 10, 32)
				if err != nil {
					output.Fail(err)
				}

				if err := (&broker.Operation{}).DescribeBroker(int32(id), flags); err != nil {
					output.Fail(err)
				}
			}
		},
		ValidArgsFunction: broker.CompleteBrokerIds,
	}

	cmdDescribeBroker.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide")

	return cmdDescribeBroker
}
