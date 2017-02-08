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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// applicationCmd represents the application command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status check of the Shipyard platform",
	Long: `This checks the status of the integral Shipyard components.

Example of use:

$ shipyardctl get status`,
	Run: func(cmd *cobra.Command, args []string) {

		// get kiln status
		kilnReq, err := http.NewRequest("GET", clusterTarget+"/organizations/status", nil)
		if debug {
			PrintDebugRequest(kilnReq)
		}

		kilnRes, err := http.DefaultClient.Do(kilnReq)
		if err != nil {
			log.Fatal(err)
		}

		if debug {
			PrintDebugResponse(kilnRes)
		}

		defer kilnRes.Body.Close()
		fmt.Print("Build service status: ")
		_, err = io.Copy(os.Stdout, kilnRes.Body)
		if err != nil {
			log.Fatal(err)
		}

		enroberReq, err := http.NewRequest("GET", clusterTarget+"/environments/status", nil)
		if debug {
			PrintDebugRequest(enroberReq)
		}

		enroberRes, err := http.DefaultClient.Do(enroberReq)
		if err != nil {
			log.Fatal(err)
		}

		if debug {
			PrintDebugResponse(enroberRes)
		}

		defer enroberRes.Body.Close()
		fmt.Print("\nDeployment service status: ")
		_, err = io.Copy(os.Stdout, enroberRes.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	getCmd.AddCommand(statusCmd)
}
