// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
	"net/http"
	"io"
	"os"
	"log"

	"github.com/spf13/cobra"
)

// applicationCmd represents the application command
var applicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "retrieve all applications in a namespace",
	Long: `This retrieves all of the applications in the configured namespace,
returning all available information.

Example of use:

$ shipyardctl get applications`,
	Run: func(cmd *cobra.Command, args []string) {
		req, err := http.NewRequest("GET", clusterTarget + imagePath + orgName + "/applications", nil)
		if verbose {
			PrintVerboseRequest(req)
		}

		req.Header.Set("Authorization", "Bearer " + authToken)
		response, err := http.DefaultClient.Do(req)

		if verbose {
			PrintVerboseResponse(response)
		}

		if err != nil {
			log.Fatal(err)
		}

		defer response.Body.Close()
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	getCmd.AddCommand(applicationsCmd)
}
