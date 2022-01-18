/*
Copyright © 2022 Dane Nelson <apogeesystemsllc@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/apogeesystems/go-dura/cmd/dura"

	"github.com/spf13/cobra"
)

// unwatchCmd represents the unwatch command
var unwatchCmd = &cobra.Command{
	Use:   "unwatch",
	Short: "Removes git repositories from Dura watch routines",
	Long: `The unwatch command removes the given paths, which are to be paths to directories representing git repositories, from the Dura configuration.
These should be repositories already listed in the Dura configuration. If more than one path is provided, each path/repository will be evaluated in the order 
in which they were provided.

This function loops through the paths one-by-one calling the appropriate unwatch command which returns an error if something went wrong. If the skip flag was not 
set, an error will force the CLI to exit.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, path := range args {
			err = dura.GetConfig().SetUnwatch(path)
			if !skip {
				cobra.CheckErr(err)
			} else {
				fmt.Println(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(unwatchCmd)

	watchCmd.Flags().BoolVarP(&skip, "skip", "s", false, "When this flag is present, if an error occurs while processing a repository, the watch command will print the error and continue rather than exiting. (default: false)")
}
