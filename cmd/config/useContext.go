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

package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

// useContextCmd represents the useContext command
var useContextCmd = &cobra.Command{
	Use:     "use-context",
	Aliases: []string{"useContext"},
	Short:   "switch active context",
	Long:    `command to switch active context`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		context := strings.Join(args, " ")

		contexts := viper.GetStringMap("contexts")

		// check if it is an existing context
		if _, ok := contexts[context]; !ok {
			fmt.Println("not a valid context:", context)
			return
		}

		viper.Set("current-context", context)

		if err := viper.WriteConfig(); err != nil {
			fmt.Println("Unable to write config:", err)
		}
	},
}

func init() {

}
