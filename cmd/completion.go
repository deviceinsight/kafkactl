package cmd

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"os"
)

const bashCompletionFunc = `
__kafkactl_get_topics()
{
    local kafkactl_topics_output out
    if kafkactl_topics_output=$(kafkactl get topics -o compact 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${kafkactl_topics_output[*]}" -- "$cur" ) )
    fi
}

__kafkactl_get_consumer_groups()
{
    local kafkactl_cg_output out
    if kafkactl_cg_output=$(kafkactl get cg -o compact 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${kafkactl_cg_output[*]}" -- "$cur" ) )
    fi
}

__kafkactl_get_contexts()
{
    local kafkactl_contexts_output out
    if kafkactl_contexts_output=$(kafkactl config get-contexts -o compact 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${kafkactl_contexts_output[*]}" -- "$cur" ) )
    fi
}

__kafkactl_custom_func() {
    case ${last_command} in
        kafkactl_consume | kafkactl_produce | kafkactl_delete_topic | kafkactl_describe_topic | kafkactl_alter_topic)
            __kafkactl_get_topics
            return
            ;;
		kafkactl_describe_consumer-group)
            __kafkactl_get_consumer_groups
            return
            ;;
		kafkactl_config_use-context)
            __kafkactl_get_contexts
            return
            ;;
        *)
            ;;
    esac
}
`

const zshCompletion = `#compdef kafkactl

_arguments \
  '1: :->level1' \
  '2: :->level2' \
  '3: :->level3' \
  '4: :_files'
case $state in
  level1)
    case $words[1] in
      kafkactl)
        _arguments '1: :(completion config consume create delete describe get help produce version)'
      ;;
      *)
        _arguments '*: :_files'
      ;;
    esac
  ;;
  level2)
    case $words[2] in
      delete)
        _arguments '2: :(topic)'
      ;;
      describe)
        _arguments '2: :(topic consumer-group)'
      ;;
      get)
        _arguments '2: :(topics consumer-groups)'
      ;;
      completion)
        _arguments '2: :(bash zsh fish)'
      ;;
      config)
        _arguments '2: :(current-context get-contexts use-context view)'
      ;;
      create)
        _arguments '2: :(topic)'
      ;;
      consume)
        _alternative "topic:topic names:($(kafkactl get topics -o compact))"
      ;;
      produce)
        _alternative "topic:topic names:($(kafkactl get topics -o compact))"
      ;;
      *)
        _arguments '*: :_files'
      ;;
    esac
  ;;
  level3)
    case $words[3] in
      topic)
        _alternative "topic:topic names:($(kafkactl get topics -o compact))"
      ;;
      consumer-group)
        _alternative "consumer-group:consumer-group names:($(kafkactl get cg -o compact))"
      ;;
      use-context)
        _alternative "context:context names:($(kafkactl config get-contexts -o compact))"
      ;;
      *)
        _arguments '*: :_files'
      ;;
    esac
  ;;
  *)
    _arguments '*: :_files'
  ;;
esac
`

func newCompletionCmd() *cobra.Command {
	var cmdCompletion = &cobra.Command{
		Use:                   "completion SHELL",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Short:                 "Output shell completion code for the specified shell (bash,zsh,fish)",
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				output.Fail(errors.Errorf("shell not specified"))
			}
			if len(args) > 1 {
				output.Fail(errors.Errorf("Too many arguments. Expected only the shell type"))
			}

			shell := args[0]

			var err error

			if shell == "bash" {
				err = runCompletionBash(cmd.Root(), os.Stdout)
			} else if shell == "zsh" {
				err = runCompletionZsh(os.Stdout)
			} else if shell == "fish" {
				err = runCompletionFish(cmd.Root(), os.Stdout)
			} else {
				err = errors.Errorf("Unsupported shell type %q.", shell)
			}

			if err != nil {
				output.Fail(errors.Wrap(err, "failed to generate completion"))
			}
		},
	}

	return cmdCompletion
}

func runCompletionBash(root *cobra.Command, out io.Writer) error {
	return root.GenBashCompletion(out)
}

func runCompletionZsh(out io.Writer) error {
	// wait until this is merged: https://github.com/spf13/cobra/pull/646
	// for now just use hardcoded completion script
	//return rootCmd.GenZshCompletion(out)
	_, err := io.WriteString(out, zshCompletion)
	return err
}

func runCompletionFish(cmd *cobra.Command, out io.Writer) error {
	// wait until this is merged: https://github.com/spf13/cobra/pull/754
	return util.GenFishCompletion(cmd, out)
}
