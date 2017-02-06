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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type EnvVar struct {
	Name      string
	Value     string
	ValueFrom *EnVarSource `json:"valueFrom,omitempty"`
}

type EnVarSource struct {
	EdgeConfigRef ConfigRef `json:"edgeConfigRef,omitempty"`
}

type ConfigRef struct {
	Name string
	Key  string
}

type Deployment struct {
	DeploymentName string
	Revision       int32
	Replicas       int32
	EnvVars        []EnvVar
}

type deploymentPatch struct {
	Revision *int32   `json:"revision,omitempty"`
	Replicas *int32   `json:"replicas,omitempty"`
	EnvVars  []EnvVar `json:"envVars,omitempty"`
}

const (
	NAME  = 0
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

		if !all {
			if err := RequireAppName(); err != nil {
				return err
			}
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if err := RequireEnvName(); err != nil {
			return err
		}

		if format == "" { // default to json for deployment retrieval
			format = "json"
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shipyardEnv := orgName + ":" + envName

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
	req, err := http.NewRequest("GET", clusterTarget+enroberPath+"/"+envName+"/deployments/"+depName, nil)
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

	// dump response body to stdout
	defer response.Body.Close()

	failure := fmt.Sprintf("\nThere was a problem retrieving %s in %s", depName, envName)

	outputBasedOnStatus("", failure, response.Body, response.StatusCode, format)

	return response.StatusCode
}

func getDeploymentAll(envName string) int {
	req, err := http.NewRequest("GET", clusterTarget+enroberPath+"/"+envName+"/deployments", nil)
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

	failure := fmt.Sprintf("\nThere was a problem retrieving deplopyments in %s", envName)

	outputBasedOnStatus("", failure, response.Body, response.StatusCode, format)

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
		shipyardEnv := orgName + ":" + envName

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
	req, err := http.NewRequest("DELETE", clusterTarget+enroberPath+"/"+envName+"/deployments/"+depName, nil)
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

	// dump response body to stdout
	defer response.Body.Close()

	success := fmt.Sprintf("\nUndeployment of %s in %s was successful", depName, envName)
	failure := fmt.Sprintf("\nThere was a problem undeploying %s in %s", depName, envName)

	outputBasedOnStatus(success, failure, response.Body, response.StatusCode, format)

	return response.StatusCode
}

// deployment creation command
var deployApplicationCmd = &cobra.Command{
	Use:   "application -o {org} -e {env} -n {name}:{revision}",
	Short: "creates a new deployment in the given environment with given app name",
	Long: `A deployment requires the application name and the organization and environment information.
Example of use:
$ shipyardctl deploy application -o acme -e test -n example:4

This command can also update an active deployment, with the --force flag.

#Update application reivision
$ shipyardctl deploy application -o acme -e test -n example:5 --force

#Update environment variable
$ shipyardctl deploy application -o acme -e test -n example --force --env-var="EXISTING_KEY=NEW_VAL"

#Force fresh deployment of an active revision, a.k.a bouncing a deployment
$ shipyardctl deploy application -o acme -e test -n example --force`,
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
		vars := parseEnvVars()
		vars = append(vars, parseConfigRefs()...)
		shipyardEnv := orgName + ":" + envName
		replicas32 := int32(defaultReplicas)

		nameSplit := strings.Split(appName, ":")

		if force {
			updateData := deploymentPatch{}

			// optionally provide revision
			if len(nameSplit) > 1 {
				revision, err := strconv.Atoi(nameSplit[1])
				if err != nil {
					log.Fatal(err)
				}

				revision32 := int32(revision)
				updateData.Revision = &revision32
			}

			if len(vars) > 0 {
				updateData.EnvVars = vars
			}

			status := updateDeployment(shipyardEnv, nameSplit[0], updateData)
			if !CheckIfAuthn(status) {
				// retry once more
				status := updateDeployment(shipyardEnv, nameSplit[0], updateData)
				if status == 401 {
					fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
					fmt.Println("Command failed.")
				}
			}
		} else {
			if len(nameSplit) < 2 {
				fmt.Println("Missing required revision number.")
				fmt.Println("\nIf you are trying to update an active deployment, please use the --force flag.")
				return
			}

			revision, err := strconv.Atoi(nameSplit[1])
			if err != nil {
				log.Fatal(err)
			}

			revision32 := int32(revision)
			status := deployApplication(shipyardEnv, nameSplit[NAME], revision32, replicas32, vars)
			if !CheckIfAuthn(status) {
				// retry once more
				status := deployApplication(envName, nameSplit[NAME], revision32, replicas32, vars)
				if status == 401 {
					fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
					fmt.Println("Command failed.")
				}
			}
		}
	},
}

func deployApplication(envName string, depName string, revision int32, replicas int32, vars []EnvVar) int {
	// prepare arguments in a Deployment struct and Marshal into JSON
	js, err := json.Marshal(Deployment{depName, revision, replicas, vars})
	if err != nil {
		log.Fatal(err)
	}

	// build API call with request body (deployment information)
	req, err := http.NewRequest("POST", clusterTarget+enroberPath+"/"+envName+"/deployments", bytes.NewBuffer(js))

	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
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

	success := fmt.Sprintf("\nCreation of %s in %s was successful", depName, envName)
	failure := fmt.Sprintf("\nThere was a problem deploying %s in %s", depName, envName)

	outputBasedOnStatus(success, failure, response.Body, response.StatusCode, format)

	return response.StatusCode
}

func updateDeployment(envName string, depName string, updateData deploymentPatch) int {
	data, err := json.Marshal(updateData)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PATCH", clusterTarget+enroberPath+"/"+envName+"/deployments/"+depName, bytes.NewBuffer(data))

	req.Header.Set("Authorization", "Bearer "+authToken)
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

	success := fmt.Sprintf("\nUpdate of %s in %s was successful", depName, envName)
	failure := fmt.Sprintf("\nThere was a problem updating %s in %s", depName, envName)

	outputBasedOnStatus(success, failure, response.Body, response.StatusCode, format)

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
		shipyardEnv := orgName + ":" + envName

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
		req, err = http.NewRequest("GET", clusterTarget+enroberPath+"/"+envName+"/deployments/"+depName+"/logs?previous=true", nil)
	} else {
		req, err = http.NewRequest("GET", clusterTarget+enroberPath+"/"+envName+"/deployments/"+depName+"/logs", nil)
	}
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
	getDeploymentCmd.Flags().StringVarP(&format, "format", "f", "", "output format for response: json, yaml, raw")

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
	deployApplicationCmd.Flags().StringVarP(&appName, "name", "n", "", "name and revision of application to deploy, ex. \"hello:3\"")
	deployApplicationCmd.Flags().StringSliceVar(&edgeConfigs, "edge-config", []string{}, "Edge-based configuration value exposed in deployment")
	deployApplicationCmd.Flags().BoolVar(&force, "force", false, "used to force an update of an active deployment")
	deployApplicationCmd.Flags().StringVarP(&format, "format", "f", "", "output format for response: json, yaml, raw")

}

func parseEnvVars() (parsed []EnvVar) {
	var temp string

	if len(envVars) > 0 {
		for i := range envVars {
			temp = envVars[i]
			split := strings.Split(temp, "=")
			parsed = append(parsed, EnvVar{Name: split[NAME], Value: split[VALUE]})
		}
	} else {
		return []EnvVar{}
	}

	return parsed
}

func parseConfigRefs() (parsed []EnvVar) {
	var temp string

	if len(edgeConfigs) > 0 {
		for i := range edgeConfigs {
			temp = edgeConfigs[i]
			split := strings.Split(temp, "=")
			valueSplit := strings.Split(split[VALUE], ":")
			parsed = append(parsed, EnvVar{Name: split[NAME], ValueFrom: &EnVarSource{ConfigRef{valueSplit[NAME], valueSplit[VALUE]}}})
		}
	} else {
		return []EnvVar{}
	}

	return parsed
}
