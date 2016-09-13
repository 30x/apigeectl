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
	"net/http"
	"os"
	"io"
	"log"
	"encoding/json"
	"bytes"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type EnvVar struct {
	Name string
	Value string
}

type Deployment struct {
	DeploymentName string
	PublicHosts string
	PrivateHosts string
	Replicas int64
	PtsUrl string
	EnvVars []EnvVar
}

type DeploymentPatch struct {
	PublicHosts string
	PrivateHosts string
	Replicas int64
	PtsUrl string
}

const (
	NAME = 0
	VALUE = 1
)

// represents the get deployment command
var deploymentCmd = &cobra.Command{
	Use:   "deployment -o {orgName} -e {envName} <deploymentName>",
	Short: "retrieves an active deployment's available information'",
	Long: `Given the name of an active deployment, this will retrieve the currently
available information in JSON format.

Example of use:
$ shipyardctl get deployment -o org1 -e env1 dep1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)

		// get all of the active deployments
		if all {
			req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + name + "/deployments" , nil)
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
		} else { // get active deployment by name
			if len(args) < 1 {
				fmt.Println("Missing required arg <deplymentName>\n")
				fmt.Println("Usage:\n\t" + cmd.Use + "\n")
				return
			}

			// get deployment name from arguments
			depName = args[0]

			// build API call
			req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + name + "/deployments/" + depName, nil)
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

			// dump response body to stdout
			defer response.Body.Close()
			_, err = io.Copy(os.Stdout, response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

var deleteDeploymentCmd = &cobra.Command{
	Use:   "deployment -o {orgName} -e {envName} <deploymentName>",
	Short: "deletes an active deployment",
	Long: `Given the name of an active deployment and the environment it belongs to,
this will delete it.

Example of use:
$ shipyardctl delete deployment -o org1 -e env1 dep1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)

		if len(args) < 1 {
			fmt.Println("Missing required arg <deplymentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		depName = args[0]

		// build API call URL
		req, err := http.NewRequest("DELETE", clusterTarget + enroberPath + "/" + name + "/deployments/" + depName, nil)
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

		// dump response body to stdout
		defer response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			fmt.Println("\nDeletion of " + depName + " in " + name + " was sucessful\n")
		}
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// deployment creation command
var createDeploymentCmd = &cobra.Command{
	Use:   "deployment -o {orgName} -e {envName} <depName> <publicHost> <privateHost> <replicas> <ptsUrl>",
	Short: "creates a new deployment in the given environment with given name",
	Long: `A deployment requires a name, accepted public and private hosts, the number
of replicas and the URL that locates the appropriate Pod Template Spec built by Shipyard.
It also requires an active environment to deploy to.

Example of use:
$ shipyardctl create deployment -o org1 -e env1 dep1 "test.host.name" "test.host.name" 2 "https://pts.url.com" --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)

		// check and pull required args
		if len(args) < 5 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		depName = args[0]
		publicHost := args[1]
		privateHost := args[2]
		replicas, err := strconv.ParseInt(args[3], 0, 64)
		if err != nil {
			log.Fatal(err)
		}
		ptsUrl := args[4]
		vars := parseEnvVars()

		// prepare arguments in a Deployment struct and Marshal into JSON
		js, err := json.Marshal(Deployment{depName, publicHost, privateHost, replicas, ptsUrl, vars})
		if err != nil {
			log.Fatal(err)
		}

		// build API call with request body (deployment information)
		req, err := http.NewRequest("POST", clusterTarget + enroberPath + "/" + name + "/deployments", bytes.NewBuffer(js))

		if verbose {
			PrintVerboseRequest(req)
		}

		req.Header.Set("Authorization", "Bearer " + authToken)
		req.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatal(err)
		}

		if verbose {
			PrintVerboseResponse(response)
		}

		// dump response to stdout
		defer response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			fmt.Println("\nCreation of " + depName + " in " + name + " was sucessful\n")
		}
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// patch/update deployment command
var patchDeploymentCmd = &cobra.Command{
	Use:   "deployment -o {orgName} -e {envName} <depName> <updateData>",
	Short: "updates an active deployment",
	Long: `Once deployed, a deployment can be updated by passing a JSON object
with the corresponding mutations. All properties, except for the deployment name are mutable.
That includes, the public or private hosts, replicas, PTS URL entirely, or the PTS itself.

Example of use:
$ shipyardctl patch deployment -o org1 -e env1 dep1 '{"replicas": 3, "publicHosts": "test.host.name.patch"}' --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)

		// check and pull required args
		if len(args) < 2 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		depName = args[0]
		updateData := args[1]

		// build API call
		// the update data will come in from command line as a JSON string
		req, err := http.NewRequest("PATCH", clusterTarget + enroberPath + "/" + name + "/deployments/"+depName, bytes.NewBuffer([]byte(updateData)))

		req.Header.Set("Authorization", "Bearer " + authToken)
		if verbose {
			PrintVerboseRequest(req)
		}

		req.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatal(err)
		}

		if verbose {
			PrintVerboseResponse(response)
		}

		defer response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			fmt.Println("\nPatch of " + depName + " in " + name + " was sucessful\n")
		}
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs -o {orgName} -e {envName} <deploymentName>",
	Short: "retrieves an active deployment's available logs",
	Long: `Given the name of an active deployment, this will retrieve the currently
available logs.

Example of use:
$ shipyardctl get logs -o org1 -e env1 dep1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireOrgName()
		RequireEnvName()
		RequireAuthToken()

		name := fmt.Sprintf("%s:%s", orgName, envName)

		if len(args) < 1 {
			fmt.Println("Missing required arg <deplymentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		// get deployment name from arguments
		depName = args[0]

		// build API call
		req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + name + "/deployments/" + depName + "/logs", nil)
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

		// dump response body to stdout
		defer response.Body.Close()
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	getCmd.AddCommand(deploymentCmd)
	deploymentCmd.Flags().BoolVarP(&all, "all", "a", false, "Retrieve all deployments")
	deploymentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")

	getCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")

	deleteCmd.AddCommand(deleteDeploymentCmd)
	deleteDeploymentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	createCmd.AddCommand(createDeploymentCmd)
	createDeploymentCmd.Flags().StringSliceVar(&envVars, "envVar", []string{}, "Environment variables to set in the deployment")
	createDeploymentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")

	patchCmd.AddCommand(patchDeploymentCmd)
	patchDeploymentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
}

func parseEnvVars() (parsed []EnvVar) {
	var temp string

	if len(envVars) > 0 {
		for i := range envVars {
			temp = envVars[i]
			split := strings.Split(temp, "=")
			parsed = append(parsed, EnvVar{split[NAME], split[VALUE]})
		}
	} else {
		return []EnvVar{}
	}

	return parsed
}
