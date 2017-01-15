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
	"strings"
	"bytes"
	"mime/multipart"
	"path/filepath"
	"fmt"

	"github.com/spf13/cobra"
)

var nodeLTS = "4"

// getApplicationCmd represents the application command
var getApplicationCmd = &cobra.Command{
	Use:   "applications",
	Short: "retrieve all applications in a appspace",
	Long: `This retrieves all of the applications in the configured appspace,
returning all available information.

Example of use:

$ shipyardctl get application --org org1`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()
		RequireOrgName()
		MakeBuildPath()

		status := getApplications()
		if !CheckIfAuthn(status) {
			// retry once more
			status := getApplications()
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func getApplications() int {
	req, err := http.NewRequest("GET", clusterTarget + basePath, nil)
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

// importAppCmd represents the import application command
var importAppCmd = &cobra.Command{
	Use:   "application --name {name}[:{revision/version}] --path {port}:{path} --directory {dir} --org {org} --runtime {runtime}[:{version}]",
	Short: "imports application into Shipyard",
	Long: `This command is used to import an application into Shipyard
from a given, zipped application source archive. Currently, node is the only supported runtime.

Within the project zip, there must be a valid package.json.

Example of use:

$ shipyardctl import application --name "echo-app1[:1]" --path "9000:/echo-app" --directory . --org acme --runtime node:4`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()
		RequireOrgName()
		RequireAppName()
		RequireAppPath()
		RequireDirectory()
		MakeBuildPath()

		status := importApp(appName, appPath, directory)
		if !CheckIfAuthn(status) {
			// retry once more
			status = importApp(appName, appPath, directory)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func importApp(appName string, appPath string, zipPath string) int {
	zip, err := os.Open(zipPath)
	if err != nil {
		log.Fatal(err)
	}
	defer zip.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(part, zip)

	if len(envVars) > 0 {
		for i := range envVars {
			writer.WriteField("envVar", envVars[i])
		}
	}

	runtimeVersion := nodeLTS
	runtimeSplit := strings.Split(runtime, ":")
	if len(runtimeSplit) > 1 {
		runtimeVersion = runtimeSplit[1]
	}

	if !isSupportedRuntime(runtimeSplit[0]) {
		fmt.Printf("Provided runtime: \"%s\"\n", runtimeSplit[0])
		fmt.Printf("Supported runtimes: \"%s\"\n", supportedRuntimes)
		fmt.Println("Exiting")
		return -1
	}

	appVersion := "1"
	nameSplit := strings.Split(appName, ":")
	if len(nameSplit) > 1 {
		appVersion = nameSplit[1]
	}

	writer.WriteField("revision", appVersion)
	writer.WriteField("name", nameSplit[0])
	writer.WriteField("publicPath", appPath)
	writer.WriteField("nodeVersion", runtimeVersion)

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", clusterTarget + basePath, body)
	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer " + authToken)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	// dump response to stdout
	defer response.Body.Close()
	if response.StatusCode == 201 {
		fmt.Println("\nApplication import successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

var deleteAppCmd = &cobra.Command{
	Use:   "application --name {name}:{revision}",
	Short: "deletes an application revision thats been imported",
	Long: `This command deletes the application revision specified.

The application must've be imported by a successful 'shipyardctl import application' command

Example of use:

$ shipyardctl delete application -n example:1 --org org1`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()
		RequireOrgName()
		RequireAppName()
		MakeBuildPath()

		nameSplit := strings.Split(appName, ":")
		if len(nameSplit) < 2 {
			fmt.Println("Application revision required")
			return
		}

		status := deleteApp(nameSplit[0], nameSplit[1])
		if !CheckIfAuthn(status) {
			// retry once more
			status := deleteApp(nameSplit[0], nameSplit[1])
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func deleteApp(appName string, revision string) int {
	req, err := http.NewRequest("DELETE", clusterTarget + basePath + "/" + appName + "/version/"+revision, nil)
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

	if response.StatusCode == 200 {
		fmt.Println("Deletion of application revision successful.")
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
	getCmd.AddCommand(getApplicationCmd)
	getApplicationCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")

	importCmd.AddCommand(importAppCmd)
	importAppCmd.Flags().StringSliceVar(&envVars, "env-var", []string{}, "Environment variable to set in the built image \"KEY=VAL\" ")
	importAppCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	importAppCmd.Flags().StringVarP(&runtime, "runtime", "u", "node:4", "Runtime to use for application and optional version, ex. node[:5]")
	importAppCmd.Flags().StringVarP(&appName, "name", "n", "", "application name and optional revision, ex. my-app[:4]")
	importAppCmd.Flags().StringVarP(&appPath, "path", "p", "", "application port and base path, ex. 9000:/hello")
	importAppCmd.Flags().StringVarP(&directory, "directory", "d", "", "directory of application source archive")

	deleteCmd.AddCommand(deleteAppCmd)
	deleteAppCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	deleteAppCmd.Flags().StringVarP(&appName, "name", "n", "", "application name and revision, ex. my-app:4")
}

func isSupportedRuntime(input string) bool {
	return strings.Contains(supportedRuntimes, input)
}
