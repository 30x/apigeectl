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
	"text/template"
	"os"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Bundle struct {
	Org string
	Env string
	PublicKey string
	AppName string
}

var savePath string
var fileMode os.FileMode

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle <appName>",
	Short: "generate an Edge proxy bundle",
	Long: `This generates the appropriate API proxy bundle for an
application built and deployed on Shipyard.

The APIGEE_ORG, APIGEE_ENVIRONMENT_NAME and PUBLIC_KEY must be
in the environment.

Example of use:

$ shipyardctl create bundle exampleApp`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Missing required arg.")
			fmt.Println("Usage:\t\n"+cmd.Use)
			return
		}

		appName := args[0]

		if pubKey == "" {
			fmt.Println("Missing environment variable PUBLIC_KEY")
			return
		}

		if envName == "" {
			fmt.Println("Missing environment variable APIGEE_ENVIRONMENT_NAME")
			return
		}

		// make a temp dir
		tmpdir, err := ioutil.TempDir("", orgName+"_"+envName)
		if err != nil {
			fmt.Println("Failed to make a temporary directory")
			return
		}

		if verbose {
			fmt.Println("Creating tmpdir at: "+tmpdir)
		}

		defer os.RemoveAll(tmpdir)

		// make apiproxy directory structure
		dir := filepath.Join(tmpdir, "apiproxy")
		err = os.Mkdir(dir, fileMode)

		if verbose {
			fmt.Println("Creating folder 'apiproxy' at: "+dir)
		}
		checkError(err, "Unable to make root apiproxy dir")

		proxiesDirPath := filepath.Join(dir, "proxies")
		err = os.Mkdir(proxiesDirPath, fileMode)
		if verbose {
			fmt.Println("Creating subfolder 'proxies' at: "+proxiesDirPath)
		}
		checkError(err, "Unable to make proxies dir")

		targetsDirPath := filepath.Join(dir, "targets")
		err = os.Mkdir(targetsDirPath, fileMode)
		if verbose {
			fmt.Println("Creating subfolder 'targets' at: "+targetsDirPath)
		}
		checkError(err, "Unable to make targets dir")

		policiesDirPath := filepath.Join(dir, "policies")
		err = os.Mkdir(policiesDirPath, fileMode)
		if verbose {
			fmt.Println("Creating subfolder 'policies' at: "+policiesDirPath)
		}
		checkError(err, "Unable to make policies dir")

		// bundle user info for templates
		bundle := Bundle{orgName, envName, pubKey, appName}

		// example.xml --> ./apiproxy/
		proxy_xml, err := os.Create(filepath.Join(dir, appName+".xml"))
		err = proxy_xml.Chmod(fileMode)
		if verbose {
			fmt.Println("Creating file '"+appName+".xml'")
		}
		checkError(err, "Unable to make "+appName+".xml file")

		proxyTmpl, err := template.New("PROXY").Parse(PROXY_XML)
		if err != nil { panic(err) }
		err = proxyTmpl.Execute(proxy_xml, bundle)
		if err != nil { panic(err) }

		// AddCors.xml --> ./apiproxy/policies
		add_cors_xml, err := os.Create(filepath.Join(policiesDirPath, "AddCors.xml"))
		err = add_cors_xml.Chmod(fileMode)
		if verbose {
			fmt.Println("Creating file 'policies/AddCors.xml'")
		}
		checkError(err, "Unable to make AddCors.xml file")

		addCors, err := template.New("ADD_CORS").Parse(ADD_CORS)
		if err != nil { panic(err) }
		err = addCors.Execute(add_cors_xml, bundle)
		if err != nil { panic(err) }

		// RetainHostHeaders.xml --> ./apiproxy/policies
		retain_host_headers_xml, err := os.Create(filepath.Join(policiesDirPath, "RetainHostHeaders.xml"))
		err = retain_host_headers_xml.Chmod(fileMode)
		if verbose {
			fmt.Println("Creating file 'policies/RetainHostHeaders.xml'")
		}
		checkError(err, "Unable to make RetainHostHeaders.xml file")

		retainHost, err := template.New("RETAIN_HOST").Parse(RETAIN_HOST)
		if err != nil { panic(err) }
		err = retainHost.Execute(retain_host_headers_xml, bundle)
		if err != nil { panic(err) }

		// SetRoutingAPIKey.xml --> ./apiproxy/policies
		set_routing_key_xml, err := os.Create(filepath.Join(policiesDirPath, "SetRoutingAPIKey.xml"))
		err = set_routing_key_xml.Chmod(fileMode)
		if verbose {
			fmt.Println("Creating file 'policies/SetRoutingAPIKey.xml'")
		}
		checkError(err, "Unable to make SetRoutingAPIKey.xml file")

		routingKey, err := template.New("ROUTING_KEY").Parse(ROUTING_KEY)
		if err != nil { panic(err) }
		err = routingKey.Execute(set_routing_key_xml, bundle)
		if err != nil { panic(err) }


		// default.xml --> ./apiproxy/proxies && ./apiproxy/targets
		proxy_default_xml, err := os.Create(filepath.Join(proxiesDirPath, "default.xml"))
		err = proxy_default_xml.Chmod(fileMode)
		if verbose {
			fmt.Println("Creating file 'proxies/default.xml'")
		}
		checkError(err, "Unable to make default.xml file")

		target_default_xml, err := os.Create(filepath.Join(targetsDirPath, "default.xml"))
		err = target_default_xml.Chmod(fileMode)
		if verbose {
			fmt.Println("Creating file 'targets/default.xml'")
		}
		checkError(err, "Unable to make default.xml file")

		proxyEndpoint, err := template.New("PROXY_ENDPOINT").Parse(PROXY_ENDPOINT)
		if err != nil { panic(err) }
		err = proxyEndpoint.Execute(proxy_default_xml, bundle)
		if err != nil { panic(err) }

		targetEndpoint, err := template.New("TARGET_ENDPOINT").Parse(TARGET_ENDPOINT)
		err = targetEndpoint.Execute(target_default_xml, bundle)
		if err != nil { panic(err) }

		// move zip to designated savePath
		if savePath != "" {
			err = os.Rename(dir, filepath.Join(savePath, "apiproxy"))
			if verbose {
				fmt.Println("Moving proxy folder to "+savePath)
			}
			checkError(err, "Unable to move apiproxy to target save directory")
		} else { // move apiproxy from tmpdir to cwd
			cwd, err := os.Getwd()
			err = os.Rename(dir, filepath.Join(cwd, "apiproxy"))
			if verbose {
				fmt.Println("Moving proxy folder to CWD")
			}
			checkError(err, "Unable to move apiproxy bundle to cwd")
		}

		if verbose {
				fmt.Println("Deleting tmpdir")
			}
	},
}


func checkError(err error, customMsg string) {
	if err != nil {
		if customMsg != "" {
			fmt.Println(customMsg)
		}

		fmt.Println("\n%v\n", err)
		os.Exit(1)
	}
}

func init() {
	createCmd.AddCommand(bundleCmd)
	bundleCmd.Flags().StringVarP(&savePath, "save", "s", "", "Save path for proxy bundle")

	fileMode = 0755
	if orgName = os.Getenv("APIGEE_ORG"); orgName == "" {
		fmt.Println("Missing required environment variable APIGEE_ORG")
		os.Exit(-1)
	}
}
