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
	"os"
	"net/http"
	"net/http/httputil"

	"github.com/spf13/cobra"
)

var verbose bool

// global variables used by most commands
var all bool
var envName string
var orgName string
var clusterTarget string
var authToken string
var depName string
var enroberPath string
var basePath string
var pubKey string
var envVars []string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "shipyardctl",
	Short: "A CLI wrapper for Enrober API",
	Long: `shipyardctl is a CLI wrapper for the Shipyard build and deploy APIs.

Pair this command with any of the available functions for applications, images,
bundles, environments or deployments.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print environment variables used and API calls made")
	RootCmd.PersistentFlags().StringVarP(&authToken, "token", "t", "", "Apigee auth token. Required. Or place in APIGEE_TOKEN.")

	// check apigeectl required environment variables
	if clusterTarget = os.Getenv("CLUSTER_TARGET"); clusterTarget == "" {
		clusterTarget = "https://shipyard.apigee.com"
	}

	// Enrober API path, appended to clusterTarget before each API call
	enroberPath = "/environments"
}

func PrintVerboseRequest(req *http.Request) {
	fmt.Println("Current environment:")
	fmt.Println("CLUSTER_TARGET="+clusterTarget)
	fmt.Println("APIGEE_ORG="+orgName)

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Println("Request dump failed. Request state is unknown. Aborting.")
		os.Exit(1)
	}
	fmt.Println("\nRequest:")
	fmt.Printf("%s\n", string(dump))
}

func PrintVerboseResponse(res *http.Response) {
	if res != nil {
		fmt.Println("\nResponse:")
		dump, err := httputil.DumpResponse(res, false)
		if err != nil {
			fmt.Println("Could not dump response")
		}

		fmt.Printf("%s", string(dump))
	}
}

func RequireAuthToken() {
	if authToken == "" {
		if authToken = os.Getenv("APIGEE_TOKEN"); authToken == "" {
			fmt.Println("Missing required flag '--token', or place in environment as APIGEE_TOKEN.")
			os.Exit(1)
		}
	}

	return
}

func RequireOrgName() {
	if orgName == "" {
		if orgName = os.Getenv("APIGEE_ORG"); orgName == "" {
			fmt.Println("Missing required flag '--org', or place in environment as APIGEE_ORG.")
			os.Exit(1)
		}
	}

	return
}

func MakeBuildPath() {
	basePath = fmt.Sprintf("/imagespaces/%s/images", orgName)
}
