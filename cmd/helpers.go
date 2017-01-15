package cmd

import (
  "os"
  "fmt"
)


// RequireOrgName used to short circuit commands
// requiring the Apigee org name if it is not present
func RequireOrgName() {
	if orgName == "" {
		if orgName = os.Getenv("APIGEE_ORG"); orgName == "" {
			fmt.Println("Missing required flag '--org', or place in environment as APIGEE_ORG.")
			os.Exit(1)
		}
	}

	return
}

// RequireEnvName used to short circuit commands
// requiring the Apigee env name if it is not present
func RequireEnvName() {
	if envName == "" {
		if envName = os.Getenv("APIGEE_ENV"); envName == "" {
			fmt.Println("Missing required flag '--env', or place in environment as APIGEE_ENV.")
			os.Exit(1)
		}
	}

	return
}

// RequireAppName used to short circuit commands
// requiring the app name if it is not present
func RequireAppName() {
	if appName == "" {
		fmt.Println("Missing required flag '--name'.")
		os.Exit(1)
	}

	return
}

// RequireAppPath used to short circuit commands
// requiring the app path if it is not present
func RequireAppPath() {
	if appPath == "" {
		fmt.Println("Missing required flag '--path'.")
		os.Exit(1)
	}

	return
}

// RequireDirectory used to short circuit commands
// requiring a directory, if it is not present
func RequireDirectory() {
	if directory == "" {
		fmt.Println("Missing required flag '--directory'.")
		os.Exit(1)
	}

	return
}

// RequirePTSURL used to short circuit commands
// requiring a PTS URL, if it is not present
func RequirePTSURL() {
	if ptsUrl == "" {
		fmt.Println("Missing required flag '--pts-url'.")
		os.Exit(1)
	}

	return
}

// RequireHostnames used to short circuit commands
// requiring hostnames, if it is not present
func RequireHostnames() {
	if hostnames == "" {
		fmt.Println("Missing required flag '--hostnames'.")
		os.Exit(1)
	}

	return
}

// RequireZipPath used to short circuit commands
// requiring the path to a bundle zip, if it is not present
func RequireZipPath() {
	if bundlePath == "" {
		fmt.Println("Missing required flag '--zip-path'.")
		os.Exit(1)
	}

	return
}

// MakeBuildPath make build service path with given orgName
func MakeBuildPath() {
	basePath = fmt.Sprintf("/imagespaces/%s/images", orgName)
}
