# shipyardctl

This project is a command line interface that wraps the Shipyard build and deploy APIs.

**While the usage is similar to `kubectl`, this is not meant to replace `kubectl`, but merely to wrap the many available API resources of Shipyard**

### Installation
Download the proper binary from the releases section of the repo, [here](https://github.com/30x/shipyardctl/releases).

```sh
> wget https://github.com/30x/shipyardctl/releases/download/v1.3.0/shipyardctl-1.3.0.darwin.amd64.go1.6.tar.gz
> tar -xvf shipyardctl-1.3.0.darwin.amd64.go1.6.tar.gz
> mv shipyardctl /usr/local/bin # might need sudo access
```

### Configuration and Environment

**Configurable values**

Here are some of the values that `shipyardctl` uses. A few of them have meaningful defaults that will not need to be changed for regular use. If there is no default, it is a user supplied value.
Each value has a few ways to be configured. Take notice of the different options you have (CLI Flag, environment variable and config file).

| Env Var | CLI Flag | In config file? | Default | Description |
| ------- |:--------:| ---------------:| -------:| -----------:|
|`APIGEE_ORG`|`--org -o`| no | n/a | Your Apigee org name|
|`APIGEE_ENVIRONMENT_NAME`|`--envName -e`| no | n/a | Your Apigee env name|
|`APIGEE_TOKEN` |`--token -t`| yes | n/a | Your JWT access token generated from Apigee credentials|
|`CLUSTER_TARGET`| n/a | yes | "https://shipyard.apigee.com" | The _protocol_ and _hostname_ of the k8s cluster |
|`SSO_LOGIN_URL`| n/a | yes | "https://login.apigee.com" | The _protocol_ and _hostname_ of the SSO target |

**Configuration resolution hierarchy**

The above variables will resolve in the following order where supported:
* CLI Flag
* Enviroment variable
* Config file

**When to use what**

Often times the values that are available to the configuration file should be managed in the config file. Using environment variables can be cumbersome and tricky to debug if you forget there is one set.
However, if you want to briefly change a value, take the token used to authenticate your `shipyardctl` calls for example, using the environment variable or CLI flag is useful and easy to undo.

The values that are not currently available to the configuration file (i.e. `org` and `envName`) should be configured with CLI flags when switching between combinations often and in envrionment variables when working in one combo for a while, to reduce command verbosity.

**Example config file**

Upon first use of `shipyarctl` it will write a configuration file to `$HOME/.shipyardctl/config`. The config file looks something like this on creation:
```yaml
currentcontext: default
contexts:
- name: default
  clusterinfo:
    name: default
    cluster: https://shipyard.apigee.com # CLUSTER_TARGET
    sso: https://login.apigee.com # SSO_LOGIN_URL
  userinfo:
    username: ""
    token: "" # APIGEE_TOKEN
```
`currentcontext`: name of the context to be referencing in `shipyardctl` use
`contexts`: set of named contexts containing cluster information and user credentials
> _Note: The `userinfo` property of a new context will be blank until you login._

**What is a context?**

A context contains the information about the cluster you are targetting with `shipyardctl` and user info that you are currently logged in as. When consume Shipyard regularly, the `default` context is all you will need.
If you are, however, running your own instance(s) of Shipyard, then having multiple contexts to easily switch your target is necessary.

### Usage

The list of available commands is as follows:
```
  ▾ shipyardctl
    ▾ login
    ▾ config
        view
        new-context
        use-context
    ▾ image
        create
        get
        delete
    ▾ applications
        get
    ▾ environment
        create
        get
        patch
        delete
    ▾ deployment
        create
        get
        patch
        delete
    ▾ bundle
        create
```

All commands support verbose output with the `-v` or `--verbose` flag.

Please also see `shipyardctl --help` for more information on the available commands and their arguments.

### Managing your config file

The config file shouldn't need to be changed much, unless you are developing on Shipyard or running your own cluster. Regardless, here are the available config management commands:

**Viewing your config file**
```sh
> shipyarctl config view
```
Prints the config file to stdout.

**Creating a new context**
```sh
> shipyarctl config new-context "e2e" --cluster-target=https://my.e2e.shipyard.com --sso-target=https://my.apigee.sso.com
New context e2e added!
Please switch contexts and login.
```
This creates a new cluster context. As mentioned before, this is helpful when you are developing Shipyard or running a
separate instance of Shipyard on a different cluster.
_Note: should any of the flags shown above be excluded, the default value will be used._

**Switching contexts**
```sh
> shipyardctl config use-context "e2e"
```
This switches the `currentcontext` property so that all following `shipyardctl` commands reference it.

## Walk through

During this walk through, we will go through the steps of building, deploying and managing a Node.js applicaion on Shipyard.

**1. Login**
```sh
> shipyardctl login --username orgAdmin@gmail.com
No config file present. Creating one now.
Creating configuration directory at: /my/home/directory/.shipyardctl
Creating configuration file at: /my/home/directory/.shipyardctl/config
Created new config file.

Enter password for username 'orgAdmin@gmail.com':

Enter your MFA token or just press 'enter' to skip:
1234

Writing credentials to config file
Successfully wrote credentials to /my/home/directory/.shipyardctl/config
```
This logs you in to a `shipyardctl` session by retrieving an auth token with your Apigee credentials and saving it to a
configuration file placed in your home directory.

> _Note: this token expires quickly, so make sure to refresh it about every 30 minutes._

**2. Build an image of a Node.js app**

This command consumes the Node.js application zip, builds it into an image, stores the image and provides the URL to retrieve its spec.

```sh
> shipyardctl create image "example" 1 "9000:/example" "./example-app.zip"
> export PTS_URL="<copy the Pod Template Spec URL generated and output by the build image command>"
```
The build command takes the name of your application, the revision number, the public port/path to reach your application
and the path to your zipped Node app.

**This command defaults to using Node.js LTS (v4) unless otherwise specified with the `--node-version` flag.**
**A list of available versions can be found [here](https://github.com/mhart/alpine-node#minimal-nodejs-docker-images-18mb-or-67mb-compressed). Provide the desired image tag as the `--node-version`.**

> _Note: there must be a valid package.json in the root of zipped application_

**3. Verify image creation**
```sh
> shipyardctl get image example 1
```
This retrieves the available information for the image specified by the application name and revision number

**4. Create a new environment**

This command will create the environment that will host your deployed Node.js applications.

```sh
> shipyardctl create environment "org1:env1" "<org name>-test.apigee.net" "<org name>-prod.apigee.net"
> export PUBLIC_KEY="<copy public key in creation response here>"
```
Here we create a new environment with the name "org1:env1" and the accepted hostnames of "orgName-test.apigee.net"
and "orgName-prod.apigee.net", a space delimited list.

> _Note: the naming convention used for hostnames is not strictly enforced, but will make Apigee Edge integration easier_

**5. Retrieve the newly created environment by name**
```sh
> shipyardctl get environment "org1:env1"
```
Here we have retrieved the newly created environment, by name.

**6. Update the environment's set of accepted hostnames**
```sh
> shipyardctl patch environment "org1:env1" "test.host.name3" "test.host.name4"
```
The environment "org1:env1" will be updated to accept traffic from the following hostnames, explicitly.

**7. Create a new deployment**

This command will create the deployment artifact that is used to manage your deployed Node.js application.

```sh
> export PUBLIC_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> export PRIVATE_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> shipyardctl create deployment "org1:env1" "example" $PUBLIC_HOST $PRIVATE_HOST 1 $PTS_URL --env "NAME1=VALUE1" -e "NAME2=VALUE2"
```
This creates a new deployment within the "org1:env1" environment with the previously generated PTS URL. The number 1 represents the number
of replicas to be made and "example" is the name of the deployment.

**8. Retrieve newly created deployment by name**
```sh
> shipyardctl get deployment "org1:env1" "example"
```
The response will include all available information on the active deployment in the given environment.

**9. Update the deployment**
```sh
> shipyardctl patch deployment "org1:env1" "example" '{"replicas": 3, "publicHosts": "replacement.host.name"}'
```
Updating a deployment by name, in a given environment, takes a JSON string that includes the properties to be changed.
This includes:
- number of replicas
- public host
- private host
- pod template spec URL
- pod template spec

**10. Create Apigee Edge Proxy bundle**
```sh
> shipyardctl create bundle "myEnvironment" --save ~/Desktop
```
This command, given the desired proxy name, will generate a valid proxy bundle for the environment deployed on Shipyard. It will be able to service
all applications deployed to the Shipyard enviroment associated with the working Edge organization and environment.
Zip this folder (named "apiproxy"), name it with your proxy name, and upload this to Apigee Edge.
Make sure to deploy the proxy after uploading it.

> _Note: you can customize the proxy base path with the `--basePath` flag. We recommend that you first create a proxy with the default base path of `/` for the_
> _entire environemnt, then make individual proxies with specific base paths **when necessary**._

> We are unable to zip the bundle for you as the zip generated by the native Go lang `archive/zip` package is not compatible
> with native Java zip packages. See [this forum](http://webmail.dev411.com/p/gg/golang-nuts/155g3s6g53/go-nuts-re-zip-files-created-with-archive-zip-arent-recognised-as-zip-files-by-java-util-zip) for an explanation.

**11. Delete the deployment**
```sh
> shipyardctl delete deployment "org1:env1" "example"
```
This deletes the named deployment.

**12. Delete the environment**
```sh
> shipyardctl delete environment "org1:env1"
```
This deletes the named environment.

**13. Delete the image**
```sh
> shipyardctl delete image "example" 1
```
This deletes the built application image, specified by the given app name and reivsion number.
