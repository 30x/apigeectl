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
	"log"
	"io"
	"os"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

type Environment struct {
	EnvironmentName string
	HostNames []string
}

type EnvironmentUpdate struct {
	HostNames []string
}

// environmentCmd represents the environment command

var environmentCmd = &cobra.Command{
	Use:   "environment -o {org} -e {env}",
	Short: "retrieves either active environment information",
	Long: `Given an environment name, this will retrieve the available information of the
active environment(s) in JSON format. Example usage looks like:

$ shipyardctl get environment -o acme -e test

OR

$ shipyardctl get environment org1:env1 --token <token>`,
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
		shipyardEnv := orgName+":"+envName
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
	req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName, nil)
	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer " + authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	defer response.Body.Close()

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

var updateEnvCmd = &cobra.Command{
	Use:   "environment -o {org} -e {env}",
	Short: "update an active environment",
	Long: `Given the name of an active environment and a space delimited
set of hostnames, the environment will be updated. A update of the hostnames
will replace them entirely.

Example of use:
$ shipyardctl update -o acme -e test --hostnames="test.host.name3,test.host.name1"`,
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

		if err := RequireHostnames(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName+":"+envName
		status := updateEnv(shipyardEnv, strings.Split(hostnames, ","))
		if !CheckIfAuthn(status) {
			// retry once more
			status := updateEnv(shipyardEnv, strings.Split(hostnames, ","))
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func updateEnv(envName string, hostnames []string) int {
	js, _ := json.Marshal(EnvironmentUpdate{hostnames})

	req, err := http.NewRequest("PATCH", clusterTarget + enroberPath + "/" + envName, bytes.NewBuffer(js))

	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer " + authToken)
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

	updateCmd.AddCommand(updateEnvCmd)
	updateEnvCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	updateEnvCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	updateEnvCmd.Flags().StringVarP(&hostnames, "hostnames", "s", "", "Accepted hostnames for the environment")
}
