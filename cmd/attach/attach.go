package attach

import (
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

func NewAttachCmd() *cobra.Command {

	var cmdAttach = &cobra.Command{
		Use:   "attach",
		Short: "run kafkactl pod in kubernetes and attach to it",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			if err := k8s.NewOperation().Attach(); err != nil {
				output.Fail(err)
			}
		},
	}

	return cmdAttach
}
