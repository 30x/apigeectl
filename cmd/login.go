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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/30x/shipyardctl/utils"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
)

var username string
var password string
var mfa string

type AuthResponse struct {
	Access_token string `json:"access_token"`
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "get new auth token",
	Long: `This retrieves a new JWT token based on Apigee credentials.

Example of use:

$ shipyardctl login -u orgAdmin@apigee.com`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := requireUsername(); err != nil {
			return err
		}

		if err := requirePassword(); err != nil {
			return err
		}

		if err := askForMFA(); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		Login()

		return
	},
}

func Login() {
	data := url.Values{}
	data.Add("username", username)
	data.Add("password", password)
	data.Add("grant_type", "password")
	payload := bytes.NewBufferString(data.Encode())
	clientAuth := "ZWRnZWNsaTplZGdlY2xpc2VjcmV0"

	var req *http.Request
	var err error
	if mfa == "" {
		req, err = http.NewRequest("POST", sso_target+"/oauth/token", payload)
	} else {
		req, err = http.NewRequest("POST", sso_target+"/oauth/token?mfa_token="+mfa, payload)
	}

	req.Header.Set("Authorization", "Basic "+clientAuth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Add("Accept", "application/json;charset=utf-8")

	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != 200 {
		fmt.Println("Invalid credentials. Failed to login.")
		os.Exit(-1)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	auth := AuthResponse{}
	err = json.Unmarshal(body, &auth)
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		fmt.Println("Authorization token:")
		fmt.Println(auth.Access_token)
	}

	fmt.Println("Writing credentials to current context")

	err = config.SaveToken(username, auth.Access_token)
	if err != nil {
		fmt.Print("Failed to write credentials to file.")
		log.Fatal(err)
	}

	fmt.Println("Successfully wrote credentials to", utils.GetConfigPath())
}

func init() {
	RootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Apigee org admin username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Apigee org admin password")
}

func requireUsername() error {
	if username == "" {
		if username = os.Getenv("APIGEE_USERNAME"); username == "" {
			consolereader := bufio.NewReader(os.Stdin)
			fmt.Println("Enter your Apigee username:")

			usr, err := consolereader.ReadString('\n')
			if err != nil {
				return err
			}

			username = strings.TrimSpace(usr)
		}
	}

	return nil
}

func requirePassword() error {
	if password == "" {
		if password = os.Getenv("APIGEE_PASSWORD"); password == "" {
			fmt.Println("Enter password for username '" + username + "':")
			pass, err := gopass.GetPasswd()
			if err != nil {
				return err
			}

			password = string(pass)
		}
	}

	return nil
}

func askForMFA() error {
	consolereader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter your MFA token or just press 'enter' to skip:")

	input, err := consolereader.ReadString('\n')
	if err != nil {
		return err
	}

	mfa = strings.TrimSpace(input)
	return nil
}
