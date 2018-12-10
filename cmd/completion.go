// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
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
    if kafkactl_contexts_output=$(kafkactl config get-contexts 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${kafkactl_contexts_output[*]}" -- "$cur" ) )
    fi
}

__custom_func() {
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

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "generate bash completion",
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenBashCompletion(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
