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
	"io"
	"mime/multipart"
	"bytes"
	"os"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

var nodeVersion string

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image <appName> <revision> <publicPath> <zipPath>",
	Short: "builds a Docker image with Shipyard",
	Long: `This command is used to build Docker images with Shipyard's build API
from a given, zipped Node.js application.

Within the project zip, there must be a valid package.json.

Example of use:

$ shipyardctl create image example 1 "9000:/example" "./path/to/zipped/app --org org1 --token <token>"`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()
		RequireOrgName()
		MakeBuildPath()

		if len(args) < 4 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t "+cmd.Use+"\n")
			return
		}

		appName := args[0]
		revision := args[1]
		publicPath := args[2]
		zipPath := args[3]

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

		writer.WriteField("revision", revision)
		writer.WriteField("name", appName)
		writer.WriteField("publicPath", publicPath)
		writer.WriteField("nodeVersion", nodeVersion)

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
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			fmt.Println("\nImage build successful\n")
		} else {
			CheckIfAuthn(response.StatusCode)
		}

		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var getImageCmd = &cobra.Command{
	Use:   "image <appName> <revision>",
	Short: "retrieves a built image's info'",
	Long: `This command retrieves the build specified by the application name and revision number.

The image must've be built by a successful 'shipyardctl build image' command

Example of use:

$ shipyardctl get image example 1

OR

$ shipyardctl get image example --all --org org1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()
		RequireOrgName()
		MakeBuildPath()

		if all {
			if len(args) < 1 {
				fmt.Println("Missing application name\n")
				return
			}

			appName := args[0]

			req, err := http.NewRequest("GET", clusterTarget + basePath + "/" + appName, nil)
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

			CheckIfAuthn(response.StatusCode)

			_, err = io.Copy(os.Stdout, response.Body)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			if len(args) < 2 {
				fmt.Println("Missing required args\n")
				fmt.Println("Usage:\n\t "+cmd.Use+"\n")
				return
			}

			appName := args[0]
			revision := args[1]

			req, err := http.NewRequest("GET", clusterTarget + basePath + "/" + appName + "/version/"+revision, nil)
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

			CheckIfAuthn(response.StatusCode)

			_, err = io.Copy(os.Stdout, response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

var deleteImageCmd = &cobra.Command{
	Use:   "image <appName> <revision>",
	Short: "deletes a built image'",
	Long: `This command deletes the image specified by the application name and revision number.

The image must've be built by a successful 'shipyardctl build image' command

Example of use:

$ shipyardctl delete image example 1 --org org1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()
		RequireOrgName()
		MakeBuildPath()

		if len(args) < 2 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t "+cmd.Use+"\n")
			return
		}

		appName := args[0]
		revision := args[1]

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

		CheckIfAuthn(response.StatusCode)

		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	createCmd.AddCommand(imageCmd)
	imageCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Environment variable to set in the built image \"KEY=VAL\" ")
	imageCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	imageCmd.Flags().StringVarP(&nodeVersion, "node-version", "n", "4", "Node version to use in base image.")

	getCmd.AddCommand(getImageCmd)
	getImageCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
	getImageCmd.Flags().BoolVarP(&all, "all", "a", false, "Retrieve all images for an application")

	deleteCmd.AddCommand(deleteImageCmd)
	deleteImageCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee org name")
}
