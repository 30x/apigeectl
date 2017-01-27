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
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// DefaultRuntime is the default runtime selection for imported apps
const DefaultRuntime = "node:4"

// getApplicationsCmd represents the application command
var getApplicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "retrieve all applications in a appspace",
	Long: `This retrieves all of the applications in the configured appspace,
returning all available information.

Example of use:

$ shipyardctl get application --org org1`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		MakeBuildPath()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
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
	req, err := http.NewRequest("GET", clusterTarget+basePath, nil)
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
	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

// getApplicationCmd represents the application command
var getApplicationCmd = &cobra.Command{
	Use:   "application -n {name} -o {org}",
	Short: "retrieve all revisions of named application in a appspace",
	Long: `This retrieves all of the revisions of the named application in the configured appspace.

Example of use:

$ shipyardctl get application --org org1 --name exampleApp`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireAuthToken(); err != nil {
			return err
		}

		if err := RequireOrgName(); err != nil {
			return err
		}

		if err := RequireAppName(); err != nil {
			return err
		}

		MakeBuildPath()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		status := getApplication(appName, orgName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := getApplication(appName, orgName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func getApplication(name string, appspace string) int {
	nameSplit := strings.Split(name, ":")

	var req *http.Request
	var err error

	if len(nameSplit) > 1 {
		req, err = http.NewRequest("GET", clusterTarget+basePath+"/"+nameSplit[0]+"/version/"+nameSplit[1], nil)
	} else {
		req, err = http.NewRequest("GET", clusterTarget+basePath+"/"+nameSplit[0], nil)
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
	Use:   "application --name {name} --directory {dir} --org {org} --runtime {runtime}[:{version}]",
	Short: "imports application into Shipyard",
	Long: `This command is used to import an application into Shipyard
from a given, zipped application source archive. Currently, node is the only supported runtime.

Within the project zip, there must be a valid package.json.

Example of use:

$ shipyardctl import application --name "echo-app1" --directory . --org acme --runtime node:4`,
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

		if err := RequireDirectory(); err != nil {
			return err
		}

		MakeBuildPath()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		status := importApp(appName, directory)
		if !CheckIfAuthn(status) {
			// retry once more
			status = importApp(appName, directory)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func importApp(appName string, zipPath string) int {
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

	if runtime == "" {
		runtime = DefaultRuntime
	}

	runtimeSplit := strings.Split(runtime, ":")

	if !isSupportedRuntime(runtimeSplit[0]) {
		fmt.Printf("Provided runtime: \"%s\"\n", runtimeSplit[0])
		fmt.Printf("Supported runtimes: \"%s\"\n", supportedRuntimes)
		fmt.Println("Exiting")
		return -1
	}

	writer.WriteField("name", appName)
	writer.WriteField("runtime", runtime)

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", clusterTarget+basePath, body)
	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
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

		MakeBuildPath()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
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
	req, err := http.NewRequest("DELETE", clusterTarget+basePath+"/"+appName+"/version/"+revision, nil)
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
	getCmd.AddCommand(getApplicationsCmd)
	getApplicationsCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")

	getCmd.AddCommand(getApplicationCmd)
	getApplicationCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	getApplicationCmd.Flags().StringVarP(&appName, "name", "n", "", "application name to retrieve and optional revision, ex. my-app[:4]")

	importCmd.AddCommand(importAppCmd)
	importAppCmd.Flags().StringSliceVar(&envVars, "env-var", []string{}, "Environment variable to set in the built image \"KEY=VAL\" ")
	importAppCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	importAppCmd.Flags().StringVarP(&runtime, "runtime", "u", "node:4", "Runtime to use for application and optional version, ex. node[:5]")
	importAppCmd.Flags().StringVarP(&appName, "name", "n", "", "application name and optional revision")
	importAppCmd.Flags().StringVarP(&directory, "directory", "d", "", "directory of application source archive")

	deleteCmd.AddCommand(deleteAppCmd)
	deleteAppCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	deleteAppCmd.Flags().StringVarP(&appName, "name", "n", "", "application name and revision, ex. my-app:4")
}

func isSupportedRuntime(input string) bool {
	return strings.Contains(supportedRuntimes, input)
}
