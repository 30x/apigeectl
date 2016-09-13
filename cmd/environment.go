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

	"github.com/spf13/cobra"
)

type Environment struct {
	EnvironmentName string
	HostNames []string
}

type EnvironmentPatch struct {
	HostNames []string
}

// environmentCmd represents the environment command

var environmentCmd = &cobra.Command{
	Use:   "environment -o {orgName} -e {envName}",
	Short: "retrieves either active environment information",
	Long: `Given an environment name, this will retrieve the available information of the
active environment(s) in JSON format. Example usage looks like:

$ shipyardctl get environment -o org1 -e env1

OR

$ shipyardctl get environment -o org1 -e env1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)
		req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + name, nil)
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
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var deleteEnvCmd = &cobra.Command{
	Use:   "environment -o {orgName} -e {envName}",
	Short: "deletes an active environment",
	Long: `Given the name of an active environment, this will delete it.

Example of use:
$ shipyardctl delete environment -o org1 -e env1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)
		req, err := http.NewRequest("DELETE", clusterTarget + enroberPath + "/" + name, nil)
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
			fmt.Println("\nDeletion of " + envName + " was sucessful\n")
		}
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var createEnvCmd = &cobra.Command{
	Use:   "environment -o {orgName} -e {envName} <hostnames...>",
	Short: "creates a new environment with name and hostnames",
	Long: `An environment is created by providing an Apigee organization and environment name,
by which it will be identified, and a space separated list of accepted hostnames.

Example of use:
$ shipyardctl create environment -o org1 -e env1 "test.host.name1" "test.host.name2" --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)

		if len(args) < 1 {
			fmt.Println("Missing required arg(s) <hostnames...>")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		hostnames := args[0:]
		js, _ := json.Marshal(Environment{name, hostnames})

		req, err := http.NewRequest("POST", clusterTarget + enroberPath, bytes.NewBuffer(js))

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
			fmt.Println("\nCreation of " + envName + " was sucessful\n")
		}
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var patchEnvCmd = &cobra.Command{
	Use:   "environment -o {orgName} -e {envName} <hostnames...>",
	Short: "update an active environment",
	Long: `Given the name of an active environment and a space delimited
set of hostnames, the environment will be updated. A patch of the hostnames
will replace them entirely.

Example of use:
$ shipyardctl patch -o org1 -e env1 "test.host.name3" "test.host.name4" --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)

		if len(args) < 1 {
			fmt.Println("Missing required arg(s) <hostnames...>")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		hostnames := args[0:]
		js, _ := json.Marshal(EnvironmentPatch{hostnames})

		req, err := http.NewRequest("PATCH", clusterTarget + enroberPath + "/" + name, bytes.NewBuffer(js))

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
			fmt.Println("\nPatch of " + envName + " was sucessful\n")
		}
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	getCmd.AddCommand(environmentCmd)
	environmentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")

	deleteCmd.AddCommand(deleteEnvCmd)
	deleteEnvCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	createCmd.AddCommand(createEnvCmd)
	createEnvCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	patchCmd.AddCommand(patchEnvCmd)
	patchEnvCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
}
