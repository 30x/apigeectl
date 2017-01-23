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
	"net/http/httputil"
	"os"

	"github.com/30x/shipyardctl/utils"
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
var sso_target string

var appName string
var appPath string
var runtime string
var directory string
var ptsUrl string
var replicasUpdate int
var replicasDeploy int
var hostnames string
var bundlePath string
var bundleName string

var supportedRuntimes = "node"

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
	context := config.GetCurrentContext()
	fmt.Println("Current context:")
	if ct := os.Getenv("CLUSTER_TARGET"); ct != "" {
		fmt.Printf("Cluster: %s (from environment variable)\n", ct)
	} else {
		fmt.Printf("Cluster: %s (from config file)\n", context.ClusterInfo.Cluster)
	}

	if st := os.Getenv("SSO_LOGIN_URL"); st != "" {
		fmt.Printf("SSO login: %s (from environment variable)\n", st)
	} else {
		fmt.Printf("SSO login: %s (from config file)\n", context.ClusterInfo.SSO)
	}

	if org := os.Getenv("APIGEE_ORG"); org != "" {
		fmt.Printf("Apigee org: %s (from environment variable)\n", org)
	} else if orgName != "" {
		fmt.Printf("Apigee org: %s (from CLI flag)\n", orgName)
	}

	if envName != "" {
		fmt.Printf("Environment name: %s\n", envName)
	}

	dump, err := httputil.DumpRequestOut(req, false) // not dump req body
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
func RequireAuthToken() error {
	if authToken == "" { // check flag first
		if authToken = os.Getenv("APIGEE_TOKEN"); authToken == "" { // check environment second
			if config != nil { // check config file last
				authToken = config.GetCurrentToken()

				if authToken == "" {
					Login()
					authToken = config.GetCurrentToken()
				}

				return nil
			} else {
				return fmt.Errorf("No config file loaded.\nMissing required auth token.\nRun shipyardctl login.")
			}
		}
	}

	return nil
}

// CheckIfAuthn checks if the API call was authenticated or not
func CheckIfAuthn(status int) bool {
	if status == 401 {
		fmt.Println("Your token has expired. Please login again.")
		username = config.GetCurrentUsername()
		Login()
		authToken = config.GetCurrentToken()
		return false
	}

	return true
}
