package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	var cmdCompletion = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "generate shell auto-completion file",
		Long: `To load completions:

Bash:

$ source <(kafkactl completion bash)

# To load completions for each session, execute once:
Linux:
  $ kafkactl completion bash > /etc/bash_completion.d/kafkactl
MacOS:
  $ kafkactl completion bash > /usr/local/etc/bash_completion.d/kafkactl

Zsh:

$ source <(kafkactl completion zsh)

# To load completions for each session, execute once:
$ kafkactl completion zsh > "${fpath[1]}/_kafkactl"

Fish:

$ kafkactl completion fish | source

# To load completions for each session, execute once:
$ kafkactl completion fish > ~/.config/fish/completions/kafkactl.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {

			var err error

			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				err = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				err = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletion(os.Stdout)
			default:
				err = errors.Errorf("unsupported shell type %q", args[0])
			}

			if err != nil {
				return errors.Wrap(err, "unable to generate shell completion")
			}
			return nil
		},
	}

	return cmdCompletion
}
