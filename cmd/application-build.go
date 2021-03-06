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
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"bufio"

	"github.com/30x/zipper"
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

		if format == "" {
			format = "get-apps"
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
	if debug {
		PrintDebugRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if debug {
		PrintDebugResponse(response)
	}

	defer response.Body.Close()
	success := fmt.Sprint("\nAvailable applications:\n")
	failure := fmt.Sprintf("\nThere was an error retrieving your imported applications")

	outputBasedOnStatus(success, failure, response.Body, response.StatusCode, format)

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
		if format == "" {
			format = "get-app-rev"
		}

		req, err = http.NewRequest("GET", clusterTarget+basePath+"/"+nameSplit[0]+"/version/"+nameSplit[1], nil)
	} else {
		if format == "" {
			format = "get-app"
		}

		req, err = http.NewRequest("GET", clusterTarget+basePath+"/"+nameSplit[0], nil)
	}

	if debug {
		PrintDebugRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if debug {
		PrintDebugResponse(response)
	}

	success := fmt.Sprintf("\nAvailable info for %s in %s:\n", name, appspace)
	failure := fmt.Sprintf("\nThere was an error retrieving %s from %s", name, appspace)

	outputBasedOnStatus(success, failure, response.Body, response.StatusCode, format)

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

func importApp(appName string, directory string) int {
	tmpdir, err := ioutil.TempDir("", appName)
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmpdir)
	zipPath := filepath.Join(tmpdir, appName+".zip")

	err = zipper.ArchiveUnprocessed(directory, zipPath, zipper.Options{
		ExcludeBaseDir: true,
	})

	if err != nil {
		log.Fatal(err)
	}

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

	if debug {
		PrintDebugRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if debug {
		PrintDebugResponse(response)
	}

	// dump response to stdout
	defer response.Body.Close()
	if response.StatusCode == 201 {
		fmt.Println("\nBeginning application import. This could take a minute.")
		err = handleBuildStream(response.Body, verbose)
		if err != nil {
			log.Fatal(err)
		}
	} else if response.StatusCode != 401 {
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

		if force {
			if err := RequireEnvName(); err != nil {
				return err
			}
		}

		MakeBuildPath()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		if !force {
			promptResponse, err := PromptAppDeletion(appName)
			if err != nil {
				log.Fatal(err)
			}

			if !promptResponse {
				fmt.Println("Chose to cancel. Aborting.")
				return
			}
		} else {
			shipyardEnv := orgName + ":" + envName
			fmt.Printf("Undeploying any active deployment of %s in %s\n", appName, shipyardEnv)

			undeployApplication(shipyardEnv, appName)

		}

		status := deleteApp(appName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := deleteApp(appName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		} else if status == http.StatusConflict {
			fmt.Println("Please use the --force flag or use the undeploy command first if you wish to undeploy and delete the application")
		}
	},
}

func deleteApp(appName string) int {
	req, err := http.NewRequest("DELETE", clusterTarget+basePath+"/"+appName, nil)
	if debug {
		PrintDebugRequest(req)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if debug {
		PrintDebugResponse(response)
	}

	defer response.Body.Close()

	if response.StatusCode == 200 {

	}

	success := fmt.Sprintf("\nDeletion of application %s successful.", appName)
	failure := fmt.Sprintf("\nThere was an error deleting %s.", appName)

	outputBasedOnStatus(success, failure, response.Body, response.StatusCode, format)

	return response.StatusCode
}

func init() {
	getCmd.AddCommand(getApplicationsCmd)
	getApplicationsCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	getApplicationsCmd.Flags().StringVar(&format, "format", "", "output format: json,yaml,raw")

	getCmd.AddCommand(getApplicationCmd)
	getApplicationCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	getApplicationCmd.Flags().StringVarP(&appName, "name", "n", "", "application name to retrieve and optional revision, ex. my-app[:4]")
	getApplicationCmd.Flags().StringVar(&format, "format", "", "output format: json,yaml,raw")

	importCmd.AddCommand(importAppCmd)
	importAppCmd.Flags().StringSliceVar(&envVars, "env-var", []string{}, "Environment variable to set in the built image \"KEY=VAL\" ")
	importAppCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	importAppCmd.Flags().StringVarP(&runtime, "runtime", "u", "node:4", "Runtime to use for application and optional version, ex. node[:5]")
	importAppCmd.Flags().StringVarP(&appName, "name", "n", "", "application name and optional revision")
	importAppCmd.Flags().StringVarP(&directory, "directory", "d", "", "directory of application source archive")
	importAppCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "stream build output to console")

	deleteCmd.AddCommand(deleteAppCmd)
	deleteAppCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	deleteAppCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name (only necessary when forcing)")
	deleteAppCmd.Flags().StringVarP(&appName, "name", "n", "", "Name of application to be deleted")
	deleteAppCmd.Flags().BoolVar(&force, "force", false, "forces the deletion of all app revisions and any active deployments")
}

func isSupportedRuntime(input string) bool {
	return strings.Contains(supportedRuntimes, input)
}

func handleBuildStream(stream io.ReadCloser, outputToConsole bool) error {
	var line string
	var data bytes.Buffer
	scanner := bufio.NewScanner(stream)

	for scanner.Scan() {
		line = scanner.Text()

		if outputToConsole {
			fmt.Println(line)
		}

		if !outputToConsole {
			data.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if match, _ := regexp.MatchString("Organization: [a-z0-9]+ | Application: [a-z0-9]+ | Revision: [a-z0-9]+", line); !match {
		if !outputToConsole {
			return fmt.Errorf("There was a problem during the build. Build output:\n%s\nPlease refer to the above build output", data.String())
		}

		// else
		return fmt.Errorf("There was a problem during the build. Refer to the build stream")
	}

	if !outputToConsole {
		fmt.Println(line) // print last scanned line
	}

	return nil
}
