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
	"github.com/30x/shipyardctl/utils"
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
var sso_target string

var config *utils.Config

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

	// check if there is a config file present
	check, err := utils.ConfigExists()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// read environment variables or use defaults
	checkEnvironmentOrDefault()

	// make a new config file because there wasn't one
	if !check {
		fmt.Println("No config file present. Creating one now.")

		err = utils.InitNewConfigFile("default", sso_target, clusterTarget)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		fmt.Printf("Created new config file.\n\n")
	}

	// read config into memory
	config, err = utils.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// environment overrides config, so check there first before setting vars based on config
	checkEnvironmentOrConfig()

	// Enrober API path, appended to clusterTarget before each API call
	enroberPath = "/environments"
}

// PrintVerboseRequest used to print the request when using verbose
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

// PrintVerboseResponse used to print the response when using verbose
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

func checkEnvironmentOrDefault() {
	if sso_target = os.Getenv("SSO_LOGIN_URL"); sso_target == "" {
		sso_target = "https://login.apigee.com"
	}

	if clusterTarget = os.Getenv("CLUSTER_TARGET"); clusterTarget == "" {
		clusterTarget = "https://shipyard.apigee.com"
	}
}

func checkEnvironmentOrConfig() {
	if os.Getenv("CLUSTER_TARGET") == "" {
		clusterTarget = config.GetCurrentClusterTarget()
	}

	if sso_target = os.Getenv("SSO_LOGIN_URL"); sso_target == "" {
		sso_target = config.GetCurrentSSOTarget()
	}
}

// RequireAuthToken used to load the auth token from:
// 1. --token flag
// 2. APIGEE_TOKEN env var
// 3. config file
// 4. Runs login sequence if there is no token at all
func RequireAuthToken() {
	if authToken == "" { // check flag first
		if authToken = os.Getenv("APIGEE_TOKEN"); authToken == "" { // check environment second
			if config != nil { // check config file last
				authToken = config.GetCurrentToken()

				if authToken == "" {
					Login()
					authToken = config.GetCurrentToken()
				}

				fmt.Println("Using auth token from config file.")
				return
			} else {
				fmt.Println("No config file loaded.")
				fmt.Println("Missing required auth token.")
				fmt.Println("Run shipyardctl login.")
				os.Exit(1)
			}
		}
	}

	return
}

// RequireOrgName used to short circuit commands
// requiring the Apigee org name if it is not present
func RequireOrgName() {
	if orgName == "" {
		if orgName = os.Getenv("APIGEE_ORG"); orgName == "" {
			fmt.Println("Missing required flag '--org', or place in environment as APIGEE_ORG.")
			os.Exit(1)
		}
	}

	return
}

// MakeBuildPath make build service path with given orgName
func MakeBuildPath() {
	basePath = fmt.Sprintf("/imagespaces/%s/images", orgName)
}
