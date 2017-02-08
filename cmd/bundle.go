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
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/30x/shipyardctl/mgmt"
	"github.com/30x/zipper"

	"github.com/spf13/cobra"
)

type Bundle struct {
	Name       string
	BasePath   string
	TargetPath string
}

var savePath string
var base string
var targetPath string
var fileMode os.FileMode

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle --name {bundleName}",
	Short: "generate an Edge proxy bundle",
	Long: `This generates the appropriate API proxy bundle for an
environment built and deployed on Shipyard.

Example of use:

$ shipyardctl create bundle -n exampleName`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RequireBundleName(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// make a temp dir
		zipDir, tmpdir, err := MakeProxyBundle(bundleName)
		defer os.RemoveAll(tmpdir)
		checkError(err, "Problmem making proxy bundle")

		// move zip to designated savePath
		if savePath != "" {
			err = os.Rename(zipDir, filepath.Join(savePath, bundleName+".zip"))
			if debug {
				fmt.Println("Moving proxy folder to " + savePath)
			}
			checkError(err, "Unable to move apiproxy to target save directory")
		} else { // move apiproxy from tmpdir to cwd
			cwd, err := os.Getwd()
			err = os.Rename(zipDir, filepath.Join(cwd, bundleName+".zip"))
			if debug {
				fmt.Println("Moving proxy folder to CWD")
			}
			checkError(err, "Unable to move apiproxy bundle to cwd")
		}

		if debug {
			fmt.Println("Deleting tmpdir")
		}
	},
}

var deployProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "deploys a given proxy bundle Edge",
	Long: `This command consumes a Node.js application archive, deploys it to Shipyard,
and creates an appropriate Edge proxy.

$ shipyardctl deploy proxy -o acme -e test -z /path/to/bundle/zip `,
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
		var err error
		var tmpdir string

		if bundlePath == "" {
			bundlePath, tmpdir, err = MakeProxyBundle(appName)
			defer os.RemoveAll(tmpdir)
			checkError(err, "Problem building proxy bundle")
		}

		err = mgmt.UploadProxyBundle(config.GetCurrentMgmtAPITarget(), orgName, envName, config.GetCurrentToken(), bundlePath, appName, debug)
		checkError(err, "")
	},
}

func MakeProxyBundle(name string) (string, string, error) {
	// make a temp dir
	tmpdir, err := ioutil.TempDir("", orgName+"_"+envName)
	if err != nil {
		return "", "", err
	}

	if debug {
		fmt.Println("Creating tmpdir at: " + tmpdir)
	}

	// make apiproxy directory structure
	dir := filepath.Join(tmpdir, "apiproxy")
	err = os.Mkdir(dir, fileMode)

	if debug {
		fmt.Println("Creating folder 'apiproxy' at: " + dir)
	}

	if err != nil {
		return "", "", err
	}

	proxiesDirPath := filepath.Join(dir, "proxies")
	err = os.Mkdir(proxiesDirPath, fileMode)
	if debug {
		fmt.Println("Creating subfolder 'proxies' at: " + proxiesDirPath)
	}
	if err != nil {
		return "", "", err
	}

	targetsDirPath := filepath.Join(dir, "targets")
	err = os.Mkdir(targetsDirPath, fileMode)
	if debug {
		fmt.Println("Creating subfolder 'targets' at: " + targetsDirPath)
	}
	if err != nil {
		return "", "", err
	}

	policiesDirPath := filepath.Join(dir, "policies")
	err = os.Mkdir(policiesDirPath, fileMode)
	if debug {
		fmt.Println("Creating subfolder 'policies' at: " + policiesDirPath)
	}
	if err != nil {
		return "", "", err
	}

	// bundle user info for templates
	if base == "" {
		base = fmt.Sprintf("/%s", name)
	}

	if targetPath == "" {
		targetPath = fmt.Sprintf("/%s", name)
	}

	bundle := Bundle{name, base, targetPath}

	// example.xml --> ./apiproxy/
	proxy_xml, err := os.Create(filepath.Join(dir, name+".xml"))
	err = proxy_xml.Chmod(fileMode)
	if debug {
		fmt.Println("Creating file '" + name + ".xml'")
	}
	if err != nil {
		return "", "", err
	}

	proxyTmpl, err := template.New("PROXY").Parse(PROXY_XML)
	if err != nil {
		return "", "", err
	}

	err = proxyTmpl.Execute(proxy_xml, bundle)
	if err != nil {
		return "", "", err
	}

	// AddCors.xml --> ./apiproxy/policies
	add_cors_xml, err := os.Create(filepath.Join(policiesDirPath, "AddCors.xml"))
	err = add_cors_xml.Chmod(fileMode)
	if debug {
		fmt.Println("Creating file 'policies/AddCors.xml'")
	}
	if err != nil {
		return "", "", err
	}

	addCors, err := template.New("ADD_CORS").Parse(ADD_CORS)
	if err != nil {
		return "", "", err
	}

	err = addCors.Execute(add_cors_xml, bundle)
	if err != nil {
		return "", "", err
	}

	// default.xml --> ./apiproxy/proxies && ./apiproxy/targets
	proxy_default_xml, err := os.Create(filepath.Join(proxiesDirPath, "default.xml"))
	err = proxy_default_xml.Chmod(fileMode)
	if debug {
		fmt.Println("Creating file 'proxies/default.xml'")
	}
	if err != nil {
		return "", "", err
	}

	target_default_xml, err := os.Create(filepath.Join(targetsDirPath, "default.xml"))
	err = target_default_xml.Chmod(fileMode)
	if debug {
		fmt.Println("Creating file 'targets/default.xml'")
	}
	if err != nil {
		return "", "", err
	}

	proxyEndpoint, err := template.New("PROXY_ENDPOINT").Parse(PROXY_ENDPOINT)
	if err != nil {
		return "", "", err
	}

	err = proxyEndpoint.Execute(proxy_default_xml, bundle)
	if err != nil {
		return "", "", err
	}

	targetEndpoint, err := template.New("TARGET_ENDPOINT").Parse(TARGET_ENDPOINT)
	err = targetEndpoint.Execute(target_default_xml, bundle)
	if err != nil {
		return "", "", err
	}

	zipDir := filepath.Join(tmpdir, name+".zip")
	err = zipper.Archive(dir, zipDir, zipper.Options{})
	if err != nil {
		return "", "", err
	}

	return zipDir, tmpdir, nil
}

func checkError(err error, customMsg string) {
	if err != nil {
		if customMsg != "" {
			fmt.Println(customMsg)
		}

		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
}

func init() {
	createCmd.AddCommand(bundleCmd)
	bundleCmd.Flags().StringVarP(&bundleName, "name", "n", "", "Proxy bundle name")
	bundleCmd.Flags().StringVarP(&savePath, "save", "s", "", "Save path for proxy bundle")
	bundleCmd.Flags().StringVarP(&base, "basePath", "b", "/", "Proxy base path. Defaults to /{name}")
	bundleCmd.Flags().StringVarP(&targetPath, "targetPath", "p", "/", "Application public path. Defaults to /{name}")

	deployCmd.AddCommand(deployProxyCmd)
	deployProxyCmd.Flags().StringVarP(&orgName, "org", "o", "", "Apigee organization name")
	deployProxyCmd.Flags().StringVarP(&envName, "env", "e", "", "Apigee environment name")
	deployProxyCmd.Flags().StringVarP(&appName, "name", "n", "", "name of proxy to be deployed")
	deployProxyCmd.Flags().StringVarP(&base, "basePath", "b", "", "Proxy base path. Defaults to /{name}")
	deployProxyCmd.Flags().StringVarP(&targetPath, "targetPath", "p", "/", "Target base path. Defaults to /{name}")
	deployProxyCmd.Flags().StringVarP(&bundlePath, "zip-path", "z", "", "path to the proxy bundle zip")

	fileMode = 0755
}
