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
)

// currentContextCmd represents the currentContext command
var CurrentContextCmd = &cobra.Command{
	Use:     "current-context",
	Aliases: []string{"currentContext"},
	Short:   "show current context",
	Long:    `Displays the name of context that is currently active`,
	Run: func(cmd *cobra.Command, args []string) {
		context := viper.GetString("current-context")
		fmt.Println(context)
	},
}

func init() {
}
