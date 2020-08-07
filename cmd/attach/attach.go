package attach

import (
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func NewAttachCmd() *cobra.Command {

	var cmdAttach = &cobra.Command{
		Use:   "attach",
		Short: "run kafkactl pod in kubernetes and attach to it",
		Args:  cobra.NoArgs,
		Run: func(cobraCmd *cobra.Command, args []string) {
			if err := (&k8s.K8sOperation{}).Attach(); err != nil {
				output.Fail(err)
			}
		},
	}

	return cmdAttach
}
