package attach

import (
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/spf13/cobra"
)

func NewAttachCmd() *cobra.Command {

	var cmdAttach = &cobra.Command{
		Use:   "attach",
		Short: "run kafkactl pod in kubernetes and attach to it",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return k8s.NewOperation().Attach()
		},
	}

	return cmdAttach
}
