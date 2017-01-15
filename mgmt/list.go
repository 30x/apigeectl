package mgmt

import (
  "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"
)

// ProxyList list of proxy names
type ProxyList []string

// ListProxies lsits the proxy in an org
func ListProxies(target string, org string, token string) (list ProxyList, err error) {
  url := fmt.Sprintf("%s/v1/o/%s/apis", target, org)

  req, err := http.NewRequest("GET", url, nil)
  if err != nil { return nil, err}

  req.Header.Set("Authorization", "Bearer " + token)

  resp, err := http.DefaultClient.Do(req)
  if err != nil { return nil, err }

  data, err := ioutil.ReadAll(resp.Body)

  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("Error retrieving proxy list: %s %s", resp.Status, string(data))
  }

  list = ProxyList{}
  err = json.Unmarshal(data, &list)
  if err != nil { return nil, err }

  return list, nil
}