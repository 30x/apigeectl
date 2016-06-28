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
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool

// global variables used by most commands
var all bool
var envName string
var orgName string
var clusterTarget string
var authToken string
var depName string
var apiPath string
var buildPath string
var imagePath string
var pubKey string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "shipyardctl",
	Short: "A CLI wrapper for Enrober API",
	Long: `shipyardctl is a CLI wrapper for the deployment management API known as Enrober.
This API is used for managing and creating environments and deployments
in the Apigee Kuberenetes cluster solution. It is to be used in conjunction with Shipyard,
for the image building.

Pair this command with any of the available functions for environments or deployments.`,
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.apigeectl.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print environment variables used and API calls made")

	// check apigeectl required environment variables
	if clusterTarget = os.Getenv("CLUSTER_TARGET"); clusterTarget == "" {
		clusterTarget = "https://shipyard.apigee.com"
	}

	if authToken = os.Getenv("APIGEE_TOKEN"); authToken == "" {
		fmt.Println("Missing required environment variable APIGEE_TOKEN")
		os.Exit(-1)
	}

	if orgName = os.Getenv("APIGEE_ORG"); orgName == "" {
		fmt.Println("Missing required environment variable APIGEE_ORG")
		os.Exit(-1)
	}

	pubKey = os.Getenv("PUBLIC_KEY")
	envName = os.Getenv("APIGEE_ENVIRONMENT_NAME");

	// Enrober API path, appended to clusterTarget before each API call
	apiPath = "/beeswax/deploy/api/v1"
	imagePath = "/beeswax/images/api/v1/namespaces/"
	buildPath = "/beeswax/images/api/v1/builds/"
}

// NOTE: this is auto-generated code from Cobra, not sure it's actually doing anything
// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".apigeectl") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func PrintVerboseRequest(req *http.Request) {
	fmt.Println("Current environment:")
	fmt.Println("CLUSTER_TARGET="+clusterTarget)
	fmt.Println("APIGEE_ORG="+orgName)

	if envName != "" {
		fmt.Println("APIGEE_ENVIRONMENT_NAME="+envName)
	}

	if pubKey != "" {
		fmt.Println("PUBLIC_KEY="+pubKey)
	}

	dump, err := httputil.DumpRequest(req, true)
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
