package cmd

import (
  "fmt"
  "os"
  "log"

  "github.com/spf13/cobra"
  "github.com/30x/shipyardctl/utils"
)

var cluster string
var sso string

var useContextCmd = &cobra.Command{
	Use:   "use-context",
	Short: "switch context",
	Long: `Switch shipyardctl context to given context name.
You should login again after using this, to be sure the token is fresh.

Example of use:

$ shipyardctl config use-context e2e`,
	Run: func(cmd *cobra.Command, args []string) {
    if len(args) < 1 {
      fmt.Println("Missing required context name")
      os.Exit(-1)
    }

    contextName := args[0]

    if config == nil { // no config file
      fmt.Println("There is no config file present at:", utils.GetConfigPath())
    } else { // switch the current context to give name
      err := config.SetContext(contextName)
      if err != nil {
        fmt.Println(err)
        os.Exit(-1)
      }
    }

    return
	},
}

var newContextCmd = &cobra.Command{
	Use:   "new-context <name>",
	Short: "new-context",
	Long: `Create a new configuration context with given name.
Cluster target information will default to cluster-target=https://shipyard.apigee.com
and sso-target=https://login.apigee.com unless otherwise specified.

Example of use:

$ shipyardctl config new-context e2e`,
	Run: func(cmd *cobra.Command, args []string) {
    if len(args) < 1 {
      fmt.Println("Missing required context name")
      os.Exit(-1)
    }

    contextName := args[0]

    if config == nil { // no config file
      fmt.Println("There is no config file present at:", utils.GetConfigPath())
    } else { // switch the current context to give name
      err := config.NewContext(contextName, sso, cluster)
      if err != nil {
        fmt.Println(err)
        os.Exit(-1)
      }
    }

    fmt.Printf("New context %s added!\nPlease switch contexts and login.\n", contextName)

    return
	},
}

var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "view",
	Long: `Prints out the config file present at $HOME/.shipyardctl/config

Example of use:

$ shipyardctl config view`,
	Run: func(cmd *cobra.Command, args []string) {

    if config == nil { // no config file
      fmt.Println("There is no config file present at:", utils.GetConfigPath())
    } else { // dump config file to stdout
      err := config.DumpConfig()
      if err != nil {
        log.Fatal(err)
      }
    }

    return
	},
}

var ConfigCmd = &cobra.Command{
	Use:   "config <sub-command>",
	Short: "config based commands",
	Long: `Supports all of the configuration based commands for shipyardctl.

Example of use:

$ shipyardctl config use-context e2e

$ shipyardctl config view

$ shipyardctl config new-context prod --cluster-target=https://my.shipyard.com`,
}

func init() {
  ConfigCmd.AddCommand(viewConfigCmd)
	ConfigCmd.AddCommand(useContextCmd)
  ConfigCmd.AddCommand(newContextCmd)
  newContextCmd.Flags().StringVarP(&cluster, "cluster-target", "c", "https://shipyard.apigee.com", "Indicates the URL of the target cluster")
  newContextCmd.Flags().StringVarP(&sso, "sso-target", "s", "https://login.apigee.com", "Indicates the URL of the SSO target")
  RootCmd.AddCommand(ConfigCmd)
}