package utils

import (
  "os"
  "os/user"
  "io/ioutil"
  "path/filepath"
  "fmt"

  yaml "gopkg.in/yaml.v2"
)

// ConfigExists checks if the config file exists
func ConfigExists() (bool, error) {
  path, err := getConfigPath()
  if err != nil {
    return false, err
  }
  return exists(path)
}

// GetConfigPath exported version of getConfigPath
func GetConfigPath() string {
  path, err := getConfigPath()
  if err != nil {
    fmt.Println(err)
    os.Exit(-1)
  }

  return path
}

// LoadConfig reads the config file into memory
func LoadConfig() (*Config, error) {
  path, err := getConfigPath()
  if err != nil {
    return nil, err
  }

  config := Config{}
  data, err := ioutil.ReadFile(path)
  if err != nil {
    return nil, err
  }

  err = yaml.Unmarshal(data, &config)
  if err != nil {
    return nil, err
  }

  return &config, nil
}

func exists(path string) (bool, error) {
  _, err := os.Stat(path)
  if err == nil { return true, nil }
  if os.IsNotExist(err) { return false, nil }
  return true, err
}

func homedir() (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }

  return usr.HomeDir, nil
}

func getConfigPath() (string, error) {
  home, err := homedir()
  if err != nil {
    return "", err
  }

  return filepath.Join(home, ShipyardctlConfigDir, ShipyardctlConfigFileName), err
}