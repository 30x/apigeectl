#shipyardctl

This project is a command line interface that wraps the Shipyard build and deploy APIs.

**While the usage is similar to `kubectl`, this is not meant to replace `kubectl`, but merely to wrap the many available API resources of Shipyard**

###Installation
Download the proper binary from the releases section of the repo, [here](https://github.com/30x/shipyardctl/releases).

```sh
> wget https://github.com/30x/shipyardctl/releases/download/v1.2.1/shipyardctl-1.2.1.darwin.amd64.go1.6.tar.gz
> tar -xvf shipyardctl-1.3.0.darwin.amd64.go1.6.tar.gz
> mv shipyardctl /usr/local/bin # might need sudo access
```

###Configuration and Environment
Upon first use of `shipyarctl` it will write a configuration file to `$HOME/.shipyardctl/config`. The config file looks something like this on creation:
```yaml
currentcontext: default
contexts:
- name: default
  clusterinfo:
    name: default
    clustertarget: https://shipyard.apigee.com
    ssotarget: https://login.apigee.com
  userinfo:
    username: ""
    token: ""
```
`currentcontext`: name of the context to be referencing in `shipyardctl` use
`contexts`: set of named contexts containing cluster information and user credentials
_Note: The `userinfo` property of a new context will be blank until you login._

**Environment**

| Env Var | CLI Flag | Description |
| ------- |:--------:| -----------:|
|`APIGEE_ORG`|`--org -o`| Your Apigee org name|
|`APIGEE_ENVIRONMENT_NAME`|`--envName -e`| Your Apigee env name|
|`APIGEE_TOKEN` |`--token -t`|Your JWT access token generated from Apigee credentials|
|`CLUSTER_TARGET`| n/a |The _protocol_ and _hostname_ of the k8s cluster (**default:** "https://shipyard.apigee.com")|
|`SSO_LOGIN_URL`| n/a |The _protocol_ and _hostname_ of the SSO target (**default:** "https://login.apigee.com")|

**Configuration hierarchy**
The above variables will resolve in the following order:
* CLI Flag (when supported)
* Enviroment variable
* Config file

###Usage

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

###Managing your config file

The config file shouldn't need to be changed much, unless you are developing on Shipyard or running your own cluster. Regardless, here are the available config management commands:

**View**
```sh
> shipyarctl config view
```
Prints the config file to stdout.

**New Context**
```sh
> shipyarctl config new-context "e2e" --cluster-target=https://my.e2e.shipyard.com --sso-target=https://my.apigee.sso.com
New context e2e added!
Please switch contexts and login.
```
This creates a new cluster context. As mentioned before, this is helpful when you are developing Shipyard or running a
separate instance of Shipyard on a different cluster.
_Note: should any of the flags shown above be excluded, the default value will be used._

**Switch Context**
```sh
> shipyardctl config use-context "e2e"
```
This switches the `currentcontext` property so that all following `shipyardctl` commands reference it.

####Walk through

**Login**
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

_Note: this token expires quickly, so make sure to refresh it about every 30 minutes._

**Build an image of a Node.js app**
```sh
> shipyardctl create image "example" 1 "9000:/example" "./example-app.zip"
> export PTS_URL="<copy the Pod Template Spec URL generated and output by the build image command>"
```
The build command takes the name of your application, the revision number, the public port/path to reach your application
and the path to your zipped Node app.

**This command defaults to using Node.js LTS (v4) unless otherwise specified with the `--node-version` flag.**
**A list of available versions can be found [here](https://github.com/mhart/alpine-node#minimal-nodejs-docker-images-18mb-or-67mb-compressed). Provide the desired image tag as the `--node-version`.**

_Note: there must be a valid package.json in the root of zipped application_

**Verify image creation**
```sh
> shipyardctl get image example 1
```
This retrieves the available information for the image specified by the application name and revision number

**Create a new environment**
```sh
> shipyardctl create environment "org1:env1" "<org name>-test.apigee.net" "<org name>-prod.apigee.net"
> export PUBLIC_KEY="<copy public key in creation response here>"
```
Here we create a new environment with the name "org1:env1" and the accepted hostnames of "orgName-test.apigee.net"
and "orgName-prod.apigee.net", a space delimited list.

_Note: the naming convention used for hostnames is not strictly enforced, but will make Apigee Edge integration easier_

**Retrieve the newly created environment by name**
```sh
> shipyardctl get environment "org1:env1"
```
Here we have retrieved the newly created environment, by name.

**Update the environment's set of accepted hostnames**
```sh
> shipyardctl patch environment "org1:env1" "test.host.name3" "test.host.name4"
```
The environment "org1:env1" will be updated to accept traffic from the following hostnames, explicitly.

**Create a new deployment**
```sh
> export PUBLIC_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> export PRIVATE_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> shipyardctl create deployment "org1:env1" "example" $PUBLIC_HOST $PRIVATE_HOST 1 $PTS_URL --env "NAME1=VALUE1" -e "NAME2=VALUE2"
```
This creates a new deployment within the "org1:env1" environment with the previously generated PTS URL. The number 1 represents the number
of replicas to be made and "example" is the name of the deployment.

**Retrieve newly created deployment by name**
```sh
> shipyardctl get deployment "org1:env1" "example"
```
The response will include all available information on the active deployment in the given environment.

**Update the deployment**
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

**Create Apigee Edge Proxy bundle**
```sh
> shipyardctl create bundle "myEnvironment" --save ~/Desktop
```
This command, given the desired proxy name, will generate a valid proxy bundle for the environment deployed on Shipyard. It will be able to service
all applications deployed to the Shipyard enviroment associated with the working Edge organization and environment.
Zip this folder (named "apiproxy"), name it with your proxy name, and upload this to Apigee Edge.
Make sure to deploy the proxy after uploading it.

_Note: you can customize the proxy base path with the `--basePath` flag. We recommend that you first create a proxy with the default base path of `/` for the_
_entire environemnt, then make individual proxies with specific base paths **when necessary**._

> We are unable to zip the bundle for you as the zip generated by the native Go lang `archive/zip` package is not compatible
> with native Java zip packages. See [this forum](http://webmail.dev411.com/p/gg/golang-nuts/155g3s6g53/go-nuts-re-zip-files-created-with-archive-zip-arent-recognised-as-zip-files-by-java-util-zip) for an explanation.

**Delete the deployment**
```sh
> shipyardctl delete deployment "org1:env1" "example"
```
This deletes the named deployment.

**Delete the environment**
```sh
> shipyardctl delete environment "org1:env1"
```
This deletes the named environment.

**Delete the image**
```sh
> shipyardctl delete image "example" 1
```
This deletes the built application image, specified by the given app name and reivsion number.
