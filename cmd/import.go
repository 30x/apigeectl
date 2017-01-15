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

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [command]",
	Short: "imports an artifact into Shipyard",
	Long: `This command, when paired with the proper subcommand, will import the respective artifact.

shipyardctl import application --name "echo-app1" --path "9000:/echo-app" --directory . --org acme --runtime node:4`,
}

func init() {
	RootCmd.AddCommand(importCmd)
}
