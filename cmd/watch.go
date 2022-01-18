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
	"os"
	"strings"
)

var (
	maxDepth    int
	include     []string
	exclude     []string
	force       bool
	skip        bool
	defMaxDepth = 255
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Adds git repositories for Dura to watch",
	Long: `The watch command adds the given paths, which are to be paths to directories representing git repositories, to the Dura configuration. 
If more than one path is provided, the include, exclude and max-depth watch configuration settings will be applied to each of the repositories provided.
If the force flag is not provided the CLI will prompt for the user to accept these changes.

This function loops through the paths provided one-by-one calling the appropriate watch command which returns an error each time, the paths will be accessed in the 
order they were provided but if one should produce an error, and the skip flag was not provided, the entire CLI will exit. Additionally, for each watch command execution, the configuration file is 
written and then re-read this is due to the fact that this command should *usually* be executed once per repository but the added functionality to specify more is
for convenience.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 && (len(include) > 0 || len(exclude) > 0) && !force {
			yn := "n"
			fmt.Scanf("Apply the same watch configuration to all of these repositories? (y/N): %s", &yn)
			if strings.ToLower(yn)[0:1] == "n" {
				return
			}
		}
		if maxDepth < 0 || maxDepth > 255 {
			fmt.Fprintln(os.Stderr, "Max depth must be between 0-255, setting value back to the default (255)")
			maxDepth = defMaxDepth
		}
		for _, path := range args {
			err = dura.GetConfig().SetWatch(path, dura.WatchConfig{
				Include:  include,
				Exclude:  exclude,
				MaxDepth: maxDepth,
			})
			if !skip {
				cobra.CheckErr(err)
			} else if err != nil {
				fmt.Println(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().IntVarP(&maxDepth, "max-depth", "d", defMaxDepth, "Set recursion max depth, value must be between 0-255. (default: 255)")
	watchCmd.Flags().StringSliceVarP(&include, "include", "i", []string{}, `A comma separated list of gitignore strings representing files/folders to explicitly include in watch routines.
Example: -i "**/.log,/dura,tests/theTests*.test"
(default: [])`)
	watchCmd.Flags().StringSliceVarP(&exclude, "exclude", "e", []string{}, `A comma separated list of gitignore strings representing files/folders to explicitly exclude in watch routines.
Example: -e "**/.log,/dura,tests/theTests*.test"
(default: [])`)
	watchCmd.Flags().BoolVarP(&force, "force", "f", false, "Forces the watch action without asking for input (in the case where multiple arguments are provided). (default: false)")
	watchCmd.Flags().BoolVarP(&skip, "skip", "s", false, "When this flag is present, if an error occurs while processing a repository, the watch command will print the error and continue rather than exiting. (default: false)")
}
