package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"bytes"

	"github.com/ryanuber/columnize"
	yaml "gopkg.in/yaml.v2"
)

// RequireOrgName used to short circuit commands
// requiring the Apigee org name if it is not present
func RequireOrgName() error {
	if orgName == "" {
		if orgName = os.Getenv("APIGEE_ORG"); orgName == "" {
			return fmt.Errorf("Missing required flag '--org', or place in environment as APIGEE_ORG.")
		}
	}

	return nil
}

// RequireEnvName used to short circuit commands
// requiring the Apigee env name if it is not present
func RequireEnvName() error {
	if envName == "" {
		if envName = os.Getenv("APIGEE_ENV"); envName == "" {
			return fmt.Errorf("Missing required flag '--env', or place in environment as APIGEE_ENV.")
		}
	}

	return nil
}

// RequireAppName used to short circuit commands
// requiring the app name if it is not present
func RequireAppName() error {
	if appName == "" {
		return fmt.Errorf("Missing required flag '--name'.")
	}

	return nil
}

// RequireBundleName used to short circuit commands
// requiring the bundle name be provided via the name flag
func RequireBundleName() error {
	if bundleName == "" {
		return fmt.Errorf("Missing required flag '--name'.")
	}

	return nil
}

// RequireDirectory used to short circuit commands
// requiring a directory, if it is not present
func RequireDirectory() error {
	if directory == "" {
		return fmt.Errorf("Missing required flag '--directory'.")
	}

	return nil
}

// RequireZipPath used to short circuit commands
// requiring the path to a bundle zip, if it is not present
func RequireZipPath() error {
	if bundlePath == "" {
		return fmt.Errorf("Missing required flag '--zip-path'.")
	}

	return nil
}

// MakeBuildPath make build service path with given orgName
func MakeBuildPath() {
	basePath = fmt.Sprintf("/organizations/%s/apps", orgName)
}

// PromptAppDeletion prompts the user trying to delete an app before they do it
func PromptAppDeletion(name string) (bool, error) {
	consolereader := bufio.NewReader(os.Stdin)
	fmt.Printf("You are about to delete all revisions of \"%s\". Are you sure? [Y/n]: ", name)

	input, err := consolereader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(input)

	if input == "Y" {
		return true, nil
	}

	return false, nil
}

func templateParseRevision(labels map[string]interface{}) interface{} {
	return labels["edge/app.rev"]
}

func columnizeOutput(format string, data interface{}, temp string) ([]byte, error) {
	funcMap := template.FuncMap{
		"revision": templateParseRevision,
	}

	tempGen, err := template.New("tempGen").Funcs(funcMap).Parse(temp)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tempGen.Execute(&buffer, data)
	if err != nil {
		return nil, err
	}

	out := columnize.SimpleFormat(strings.Split(buffer.String(), "\n"))

	return []byte(out), nil
}

func formatOutput(format string, body io.ReadCloser) ([]byte, error) {
	var dat interface{}

	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if len(buf) == 0 {
		return nil, nil
	}

	if format != "yaml" { // unmarshal via json once so this code isn't copied several times
		err = json.Unmarshal(buf, &dat)
		if err != nil {
			return nil, err
		}
	}

	switch format {
	case "json":
		return json.MarshalIndent(dat, "", "  ")
	case "yaml":
		err = yaml.Unmarshal(buf, &dat)
		if err != nil {
			return nil, err
		}

		return yaml.Marshal(dat)
	case "raw":
		return buf, nil
	// human readables
	case "get-app":
		return columnizeOutput(format, dat, GET_APP)
	case "get-app-rev":
		return columnizeOutput(format, dat, GET_APP_REV)
	case "get-apps":
		return columnizeOutput(format, dat, GET_APPS)
	case "get-dep":
		return columnizeOutput(format, dat, GET_DEP)
	case "get-deps":
		return columnizeOutput(format, dat, GET_DEPS)
	default:
		return nil, nil
	}
}

func outputBasedOnStatus(success string, failure string, body io.ReadCloser, status int, format string) {
	if status == 200 || status == 201 || status == 204 {
		if success != "" && format == "" {
			fmt.Println(success)
		}

		out, err := formatOutput(format, body)
		if err != nil {
			fmt.Println(err)
		} else if out != nil {
			fmt.Println(string(out))
		}
	} else if status == 401 {
		return // we handle this special
	} else if status == 403 {
		if failure != "" {
			fmt.Println(failure)
		}

		out, err := formatOutput("raw", body)
		if err != nil {
			fmt.Println(err)
		}

		if out != nil {
			fmt.Println(string(out))
		}
	} else if status == 404 {
		if failure != "" {
			fmt.Println(failure)
		}

		fmt.Println("Received a 404. Resource not found.")
	} else {
		if failure != "" {
			fmt.Println(failure)
		}

		out, err := formatOutput(format, body)
		if err != nil {
			fmt.Println(err)
		}

		if out != nil {
			fmt.Println(string(out))
		}
	}
}
