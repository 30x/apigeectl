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

// environmentCmd represents the environment command

var environmentCmd = &cobra.Command{
	Use:   "environment -o {org} -e {env}",
	Short: "retrieves either active environment information",
	Long: `Given an environment name, this will retrieve the available information of the
active environment(s) in JSON format. Example usage looks like:

$ shipyardctl get environment -o acme -e test`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if format == "" {
			format = "get-env"
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName + ":" + envName
		status := getEnvironment(shipyardEnv)
		if !CheckIfAuthn(status) {
			// retry once more
			status := getEnvironment(shipyardEnv)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func getEnvironment(envName string) int {
	req, err := http.NewRequest("GET", clusterTarget+enroberPath+"/"+envName, nil)
	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	defer response.Body.Close()

	success := fmt.Sprintf("\nAvailable information for %s:", envName)
	failure := fmt.Sprintf("\nThere was an error retrieving %s", envName)

	outputBasedOnStatus(success, failure, response.Body, response.StatusCode, format)

	return response.StatusCode
}

var syncEnvCmd = &cobra.Command{
	Use:   "environment -o {org} -e {env}",
	Short: "sync an active environment with Edge",
	Long: `Given the name of an active environmentit will be sync'd with Edge.

Example of use:
$ shipyardctl sync environment -o acme -e test`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName + ":" + envName
		status := syncEnv(shipyardEnv)
		if !CheckIfAuthn(status) {
			// retry once more
			status := syncEnv(shipyardEnv)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func syncEnv(envName string) int {
	req, err := http.NewRequest("PATCH", clusterTarget+enroberPath+"/"+envName, nil)

	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Println("\nPatch of " + envName + " was successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

func init() {
	getCmd.AddCommand(environmentCmd)
	environmentCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	environmentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	environmentCmd.Flags().StringVarP(&format, "format", "f", "", "output format: json,yaml,raw")

	syncCmd.AddCommand(syncEnvCmd)
	syncEnvCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	syncEnvCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
}
