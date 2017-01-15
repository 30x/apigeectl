package mgmt

import (
  "fmt"
  "os"
  "net/http"
  "io/ioutil"
)

// UploadProxyBundle uploads a zipped proxy bundle
func UploadProxyBundle(target string, org string, env string, token string, bundlePath string, name string, verbose bool) error {
  url := fmt.Sprintf("%s/v1/o/%s/apis?action=import&validate=fales&name=%s", target, org, name)

  zip, err := os.Open(bundlePath)
  if err != nil { return err }

  req, err := http.NewRequest("POST", url, zip)
  if err != nil {return err}

  req.Header.Set("Authorization", "Bearer " + token)

  resp, err := http.DefaultClient.Do(req)
  if err != nil { return err }

  data, err := ioutil.ReadAll(resp.Body)

  if resp.StatusCode != 200 && resp.StatusCode != 201 {
    return fmt.Errorf("Error retrieving proxy list: %s %s", resp.Status, string(data))
  }

  if verbose {
    fmt.Printf("%s\n", string(data))
  }

  return nil
}