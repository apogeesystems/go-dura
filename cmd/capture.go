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

// captureCmd represents the capture command
var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Run a single backup of an entire repository.",
	Long:  `Run a single backup of an entire repository. This is the one single iteration of the 'serve'' control loop.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cs *dura.CaptureStatus
		if len(args) == 0 { // Use CWD
			fmt.Println(CWD)
			cs, err = dura.Capture(CWD)
			cobra.CheckErr(err)
			fmt.Println(cs.CommitHash)
		} else { // Use paths provided
			var path string
			for _, path = range args {
				cs, err = dura.Capture(path)
				cobra.CheckErr(err)
				fmt.Println(cs.CommitHash)
			}
		}
	},
	Version: "0.0.1",
}

func init() {
	rootCmd.AddCommand(captureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// captureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// captureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
