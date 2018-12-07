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

package get

import (
	"fmt"
	"github.com/spf13/cobra"
)

// getConsumerGroupsCmd represents the useContext command
var getConsumerGroupsCmd = &cobra.Command{
	Use:     "consumer-groups",
	Aliases: []string{"groups", "consumerGroups"},
	Short:   "list available consumer groups",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented, yet :/")
	},
}

func init() {

	getConsumerGroupsCmd.Flags().StringVarP(&outputFormat, "output", "o", outputFormat, "Output format. One of: json|yaml")
}
