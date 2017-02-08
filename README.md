# shipyardctl

This project is a command line interface that wraps the Shipyard build and deploy APIs.

**While the usage is similar to `kubectl`, this is not meant to replace `kubectl`, but merely to wrap the many available API resources of Shipyard**

### Installation
Download the proper binary from the releases section of the repo, [here](https://github.com/30x/shipyardctl/releases).

```sh
> wget https://github.com/30x/shipyardctl/releases/download/v1.3.1/shipyardctl-1.3.1.darwin.amd64.go1.7.zip
> unzip shipyardctl-1.3.1.darwin.amd64.go1.7.zip
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
  proxymgmtapi: https://api.enterprise.apigee.com
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
    ▾ version
    ▾ config
        view
        new-context
        use-context
    ▾ create
        bundle
    ▾ delete
        application
    ▾ deploy
        applicationp=
        proxy
    ▾ undeploy
        application
    ▾ get
        applications
        deployment
        environment
        logs
        status
    ▾ import
        application
    ▾ update
        environment
        deployment
```

All commands support debug output with the `-v` or `--debug` flag.

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

> _Note: this token expires quickly, so every dependant command will prompt you to refresh your login and rety the command, when necessary._

**2. Import an Node.js application source code**

This command consumes the Node.js application zip, stores the application revions and provides the URL to retrieve its spec.

```sh
> shipyardctl import application --name "echo-app1[:1]" --path "9000:/echo-app" --directory "./example.zip" --org acme --runtime node:4
> export PTS_URL="<copy the Pod Template Spec URL generated and output by the build image command>"
```
The build command takes the name of your application, the revision number, the public port/path to reach your application
and the path to your zipped Node app.

> _Note: there must be a valid package.json in the root of zipped application_

**3. Verify image creation**
```sh
> shipyardctl get applications --org acme
```
This retrieves all of the imported applications into an appspace.

**4. Retrieve your Shipyard environment**
```sh
> shipyardctl get environment --org acme --env test
```
Here we have retrieved the environment, by Apigee org & env name.

**5. Update the environment's set of accepted hostnames**
```sh
> shipyardctl update environment --org acme --env test "test.host.name3"
```
The environment "acme-test" will be updated to accept traffic from the following hostnames, explicitly.

**6. Create a new deployment**

This command will create the deployment artifact that is used to manage your deployed application.

```sh
> export PUBLIC_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> export PRIVATE_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> shipyardctl deploy application -o acme -e test -n example --pts-url "https://pts.url.com"
```
This creates a new deployment within the "acme-test" environment with the imported application spec provided by the PTS URL.

**7. Retrieve newly created deployment by name**
```sh
> shipyardctl get deployment --org acme --env test --name example
```
The response will include all available information on the active deployment in the given environment.

**8. Check your deployment's logs**
```sh
> shipyardctl get logs -o acme -e test -n example
```
This will dump all of the logs available from each replica belonging to the named deployment.

**9. Update the deployment**
```sh
> shipyardctl update deployment --org acme --env test --name example --replicas 4
```
Updating a deployment by name, in a given environment, with the flags of the properties to be changed.
This includes:
- number of replicas
- pod template spec URL

**10. Create Apigee Edge Proxy bundle**
```sh
> shipyardctl create bundle "myProxy" --save ~/Desktop
```
This command, given the desired proxy name, will generate a valid proxy bundle archive for the environment deployed on Shipyard. It will be able to service
all applications deployed to the Shipyard enviroment associated with the working Edge organization and environment.
Upload this to Apigee Edge. Make sure to deploy the proxy after uploading it.

> _Note: you can customize the proxy base path with the `--basePath` flag. We recommend that you first create a proxy with the default base path of `/` for the_
> _entire environemnt, then make individual proxies with specific base paths **when necessary**. When you do this, make sure to also use the `--publicPath` option_
> _in conjunction to specify the public path your deployment services. It defaults to `/`. The `publicPath` does not have to be the same as the proxy `basePath`._

**11. Undeploy an application deployment**
```sh
> shipyardctl undeploy application -n example -o acme -e test
```
This undeploys the named deployment.

**14. Delete the application import**
```sh
> shipyardctl delete application -n example:1 --org org1
```
This deletes the imported application, specified by the given app name and optional reivsion number.
