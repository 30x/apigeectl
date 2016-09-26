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

package utils

import (
  "fmt"
  "os"
  "path/filepath"
  "io/ioutil"

  yaml "gopkg.in/yaml.v2"
)

const (
  // ShipyardctlConfigDir name of the config file directory for shipyardctl
  ShipyardctlConfigDir = ".shipyardctl"
  // ShipyardctlConfigFileName name of the config file for shipyardctl
  ShipyardctlConfigFileName = "config"
)

// InitNewConfigFile creates a new config file
func InitNewConfigFile(name string, sso string, clusterTarget string) error {

  home, err := homedir()
  if err != nil {
    return err
  }

  configDirPath := filepath.Join(home, ShipyardctlConfigDir)
  configFilePath := filepath.Join(configDirPath, ShipyardctlConfigFileName)

  // make sure the directory is there
  fmt.Println("Creating configuration directory at:", configDirPath)
  err = os.MkdirAll(configDirPath, 0755)
  if err != nil {
    return err
  }

  fmt.Println("Creating configuration file at:", configFilePath)

  // default config
  defaultConfig := MakeConfig(name, sso, clusterTarget)
  data, err := yaml.Marshal(defaultConfig)
  if err != nil {
    return err
  }

  return ioutil.WriteFile(configFilePath, data, 0755)
}

// MakeConfig creates a context named default based on the given environment
func MakeConfig(name string, sso string, clusterTarget string) *Config {
  cluster := Cluster{name, clusterTarget, sso}
  context := Context{name, cluster, User{}}

  return &Config{name, []Context{context}}
}

// GetCurrentToken retrieves the user token from the current active context
func (c *Config) GetCurrentToken() string {
  for _, con := range c.Contexts {
    if con.Name == c.CurrentContext {
      return con.UserInfo.Token
    }
  }

  return "" // couldn't find the current Context
}

// GetCurrentContext retrieves the current context
func (c *Config) GetCurrentContext() *Context {
  for _, con := range c.Contexts {
    if con.Name == c.CurrentContext {
      return &con
    }
  }

  return nil // couldn't find current Context
}

// SetContext switch current context to given context name
func (c *Config) SetContext(name string) error {
  for _, con := range c.Contexts {
    if con.Name == name { // valid context name
      c.CurrentContext = name // set current context
      return c.Save() // save change
    }
  }

  return fmt.Errorf("Invalid context name: %s", name)
}

// Save writes the config out to file
func (c *Config) Save() error {
  data, err := yaml.Marshal(c)
  if err != nil {
    return err
  }

  path, err := getConfigPath()
  if err != nil {
    return err
  }

  return ioutil.WriteFile(path, data, 0755)
}

// GetCurrentClusterTarget retrieves current context cluster target
func (c *Config) GetCurrentClusterTarget() string {
  context := c.GetCurrentContext()
  return context.ClusterInfo.Cluster
}

// GetCurrentSSOTarget retrieves current context sso target
func (c *Config) GetCurrentSSOTarget() string {
  context := c.GetCurrentContext()
  return context.ClusterInfo.SSO
}

// GetCurrentUsername retrieves the username of the current context
func (c *Config) GetCurrentUsername() string {
  context := c.GetCurrentContext()
  return context.UserInfo.Username
}

// NewContext used to create a new context
func (c *Config) NewContext(name string, sso string, clusterTarget string) error {
  c.Contexts = append(c.Contexts, Context{name, Cluster{name, clusterTarget, sso}, User{}})
  c.Save()

  return nil
}

// DumpConfig dumps the config to stdout
func (c *Config) DumpConfig() error {
  data, err := yaml.Marshal(c)
  if err != nil {
    return err
  }

  fmt.Println(string(data))

  return nil
}

// SaveToken writes the given username and token to the current context
func (c *Config) SaveToken(username string, token string) error {
  user := User{username, token}
  for ndx, con := range c.Contexts {
    if con.Name == c.CurrentContext {
      c.Contexts[ndx].UserInfo = user
      return c.Save()
    }
  }

  return fmt.Errorf("Could not find current context: %s", c.CurrentContext)
}
