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
	"strings"

	"github.com/spf13/cobra"
)

type EnvVar struct {
	Name string
	Value string
}

type Deployment struct {
	DeploymentName string
	Replicas int64
	PtsUrl string
	EnvVars []EnvVar
}

type deploymentPatch struct {
	PublicHosts  *string      `json:"publicHosts,omitempty"`
	PrivateHosts *string      `json:"privateHosts,omitempty"`
	Replicas     *int32       `json:"replicas,omitempty"`
	PtsURL       string       `json:"ptsURL"`
}

const (
	NAME = 0
	VALUE = 1
)

var previous bool

// represents the get deployment command
var getDeploymentCmd = &cobra.Command{
	Use:   "deployment -o {org} -e {env} -n {name}",
	Short: "retrieves an active deployment's available information'",
	Long: `Given the name of an active deployment, this will retrieve the currently
available information in JSON format.

Example of use:
$ shipyardctl get deployment -o acme -e test -n example`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireAppName(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName+":"+envName

		// get all of the active deployments
		if all {
			status := getDeploymentAll(shipyardEnv)
			if !CheckIfAuthn(status) {
				// retry once more
				status := getDeploymentAll(shipyardEnv)
				if status == 401 {
					fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
					fmt.Println("Command failed.")
				}
			}
		} else { // get active deployment by name

			status := getDeploymentNamed(shipyardEnv, appName)
			if !CheckIfAuthn(status) {
				// retry once more
				status := getDeploymentNamed(shipyardEnv, appName)
				if status == 401 {
					fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
					fmt.Println("Command failed.")
				}
			}
		}
	},
}

func getDeploymentNamed(envName string, depName string) int {
	// build API call
	req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName, nil)
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

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

func getDeploymentAll(envName string) int {
	req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments" , nil)
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

var undeployApplicationCmd = &cobra.Command{
	Use:   "application --name {name} --org {org} --env {env}",
	Short: "undeploys an active deployment",
	Long: `Given the name of an active deployment and the environment it belongs to,
this will undeploy it.

Example of use:
$ shipyardctl undeploy application -n example -o acme -e test`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireAppName(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName+":"+envName

		status := undeployApplication(shipyardEnv, appName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := undeployApplication(shipyardEnv, appName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func undeployApplication(envName string, depName string) int {
	// build API call URL
	req, err := http.NewRequest("DELETE", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName, nil)
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
		fmt.Println("\nUndeployment of " + depName + " in " + envName + " was successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

// deployment creation command
var deployApplicationCmd = &cobra.Command{
	Use:   "application -o {org} -e {env} -n {name} --pts-url {ptsUrl}",
	Short: "creates a new deployment in the given environment with given name",
	Long: `A deployment requires a name, the number of replicas and the URL that locates
the appropriate Pod Template Spec imported to Shipyard.

Example of use:
$ shipyardctl deploy application -o acme -e test -n example --pts-url "https://pts.url.com"`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireAppName(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		if err := RequirePTSURL(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vars := parseEnvVars()
		shipyardEnv := orgName+":"+envName
		replicas64 := int64(replicas)

		status := deployApplication(shipyardEnv, appName, replicas64, ptsUrl, vars)
		if !CheckIfAuthn(status) {
			// retry once more
			status := deployApplication(envName, appName, replicas64, ptsUrl, vars)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func deployApplication(envName string, depName string, replicas int64, ptsUrl string, vars []EnvVar) int {
	// prepare arguments in a Deployment struct and Marshal into JSON
	js, err := json.Marshal(Deployment{depName, replicas, ptsUrl, vars})
	if err != nil {
		log.Fatal(err)
	}

	// build API call with request body (deployment information)
	req, err := http.NewRequest("POST", clusterTarget + enroberPath + "/" + envName + "/deployments", bytes.NewBuffer(js))

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
		fmt.Println("\nCreation of " + depName + " in " + envName + " was successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

// patch/update deployment command
var updateDeploymentCmd = &cobra.Command{
	Use:   "deployment -o {org} -e {env} -n {name} --replicas {num}",
	Short: "updates an active deployment",
	Long: `Once deployed, a deployment can be updated by passing a JSON object
with the corresponding mutations. All properties, except for the deployment name are mutable.
That includes, the public or private hosts, replicas, PTS URL entirely, or the PTS itself.

Example of use:
$ shipyardctl update deployment shipyardctl update deployment --org acme --env test --name example --replicas 4`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireAppName(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName+":"+envName
		if replicas == -1 && ptsUrl == "" {
			fmt.Println("Nothing to update. Ending.")
			return
		}

		updateData := deploymentPatch{}
		updateData.PtsURL = ptsUrl

		if replicas != -1 {
			replicas32 := int32(replicas)
			updateData.Replicas = &replicas32
		}

		status := updateDeployment(shipyardEnv, appName, updateData)
		if !CheckIfAuthn(status) {
			// retry once more
			status := updateDeployment(shipyardEnv, depName, updateData)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func updateDeployment(envName string, depName string, updateData deploymentPatch) int {
	data, err := json.Marshal(updateData)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PATCH", clusterTarget + enroberPath + "/" + envName + "/deployments/"+depName, bytes.NewBuffer(data))

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
		fmt.Println("\nPatch of " + depName + " in " + envName + " was successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

var logsCmd = &cobra.Command{
	Use:   "logs -o {org} -e {env} -n {name}",
	Short: "retrieves an active deployment's available logs",
	Long: `Given the name of an active deployment, this will retrieve the currently
available logs.

Example of use:
$ shipyardctl get logs -o acme -e test -n example`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireAppName(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName+":"+envName

		status := getDeploymentLogs(shipyardEnv, appName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := getDeploymentLogs(shipyardEnv, appName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func getDeploymentLogs(envName string, depName string) int {
	var req *http.Request
	var err error
	// build API call
	if previous {
		req, err = http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName + "/logs?previous=true", nil)
	} else {
		req, err = http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName + "/logs", nil)
	}
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

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

func init() {
	getCmd.AddCommand(getDeploymentCmd)
	getDeploymentCmd.Flags().BoolVarP(&all, "all", "a", false, "Retrieve all deployments")
	getDeploymentCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	getDeploymentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	getDeploymentCmd.Flags().StringVarP(&appName, "name", "n", "", "name of application deployment to retrieve")

	getCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&previous, "previous", "p", false, "used to retrieve previous container's logs")
	logsCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	logsCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	logsCmd.Flags().StringVarP(&appName, "name", "n", "", "name of application deployment to retrieve logs from")

	undeployCmd.AddCommand(undeployApplicationCmd)
	undeployApplicationCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	undeployApplicationCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	undeployApplicationCmd.Flags().StringVarP(&appName, "name", "n", "", "name of application deployment to undeploy")

	deployCmd.AddCommand(deployApplicationCmd)
	deployApplicationCmd.Flags().StringSliceVar(&envVars, "env-var", []string{}, "Environment variables to set in the deployment")
	deployApplicationCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	deployApplicationCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	deployApplicationCmd.Flags().StringVarP(&appName, "name", "n", "", "name of application deployment to deploy")
	deployApplicationCmd.Flags().StringVarP(&ptsUrl, "pts-url", "p", "", "URL of the Pod Template Spec given by import")
	deployApplicationCmd.Flags().IntVarP(&replicas, "replicas", "r", -1, "Number of application replicas to deploy")

	updateCmd.AddCommand(updateDeploymentCmd)
	updateDeploymentCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	updateDeploymentCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	updateDeploymentCmd.Flags().StringVarP(&appName, "name", "n", "", "name of application deployment to deploy")
	updateDeploymentCmd.Flags().IntVarP(&replicas, "replicas", "r", -1, "number of replicas to scale to")
	updateDeploymentCmd.Flags().StringVarP(&ptsUrl, "pts-url", "p", "", "URL of the Pod Template Spec to update with")
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
