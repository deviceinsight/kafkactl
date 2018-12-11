package cmd

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
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

__kafkactl_get_contexts()
{
    local kafkactl_contexts_output out
    if kafkactl_contexts_output=$(kafkactl config get-contexts -o compact 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${kafkactl_contexts_output[*]}" -- "$cur" ) )
    fi
}

__kafkactl_custom_func() {
    case ${last_command} in
        kafkactl_consume | kafkactl_produce | kafkactl_delete_topic | kafkactl_describe_topic)
            __kafkactl_get_topics
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

var cmdCompletion = &cobra.Command{
	Use:   "completion",
	Short: "generate bash completion",
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GenBashCompletion(os.Stdout)
		if err != nil {
			output.Failf("Failed to generate bash completion: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdCompletion)
}
