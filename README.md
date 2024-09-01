# IAM-CTL

IAM-CTL is a tool that can manage WSO2 Identity Server configurations from the command line. It is written in [GO](https://go.dev/) and uses the management REST APIs of WSO2 Identity Server to manage configurations.

### Prerequisites
You need to [setup](https://is.docs.wso2.com/en/latest/get-started/sample-use-cases/set-up/) WSO2 Identity Server 7.0.0.

### Run the tool

1. Download the latest binary file from [Releases](https://github.com/wso2-extensions/identity-tools-cli/releases) based on your Operating System.

2. Extract the `tar` or `zip` file.

    Here onwards, the extracted directory path is referred to as `<IAM-CTL-PATH>`.

3. Open a terminal and create an alias for the `IAM-CTL` executable file using one of the following commands (depending on your platform):
   * linux/mac:
 
       ```
       alias iamctl="<IAM-CTL-PATH>/bin/iamctl" 
       ```

   * windows

       ```
       doskey iamctl=<IAM-CTL-PATH>\bin\iamctl.exe $*
       ```
 
4. Run the tool using the following command to get the basic details.
    ```
    iamctl -h
    ```
5. Start WSO2 IS and [register a M2M application](https://is.docs.wso2.com/en/latest/guides/applications/register-machine-to-machine-app/) with the following API authorization.


API                                              | Scopes
------------------------------------------------ | --------------------------------------------------------------------
Management --> Application Management API        | Create Application, Update Application, Delete Application, View Application
Management --> Claim Management API              | Create Claim, Update Claim, Delete Claim, View Claim
Management --> Identity Provider Management API  | Create Identity Provider, Update Identity Provider, Delete Identity Provider, View Identity Provider
Management --> Userstore Management API          | Create Userstore, Update Userstore, Delete Userstore, View Userstore

6. Take note of the client ID and client secret of this application.

## CLI mode

The CLI mode of the tool can be used to handle bulk configurations in the target environment. This can be used to promote resources across multiple environments, deploy new configurations to target environments, and act as a backup of each environment's configurations.

This mode consists of the `exportAll` and `importAll` commands that can be used to export and import all configurations of the supported resource types from or to a target environment. 

Currently, the supported resource types are: 
* Applications
* Identity Providers
* Claims
* User Stores

### Running the tool in the CLI mode
The following explains the basic steps for running the tool in the simplest way. Find more comprehensive details about the commands used in the CLI mode [here](docs/cli-mode.md).

#### Tool initialization
The tool should be initialized with the server details of the environment it is run against.
1. Create a new folder and navigate to it from your terminal.
2. Run the following command to create the configuration files needed to initialize the tool.
    ```
    iamctl setupCLI
    ```
3. A new folder named ```configs``` will be created with an ```env``` folder inside it. The `env` folder contains three configuration files: ```serverConfig.json```, ```toolConfig.json```, and ```keywordConfig.json```
> **Note:** If you have multiple environments, get a copy of the ```env``` folder and rename it according to the environments you have. For example, if you have two environments: dev and prod, have two separate config folders as ```dev``` and ```prod```. 
4. Open the ```serverConfig.json``` file and provide the WSO2 IS details and client ID/secret of the app you created earlier.

Example configurations:

```
    {
       "SERVER_URL" : "https://localhost:9443",
       "CLIENT-ID" : "********",
       "CLIENT-SECRET" : "********",
       "TENANT-DOMAIN" : "carbon.super"
    }
```

#### Export
Run the following command to export all supported resource configurations from the target environment to the current directory.
```
iamctl exportAll -c ./configs/env
```
A new set of folders are created, which are named after each resource type, with exported yaml files for each available resource in WSO2 IS.

#### Import
Run the following command to import all supported resource configurations from the current directory to the target environment.
```
iamctl importAll -c ./configs/env
```
All resources available inside each resource type folder in the current directory will be imported to WSO2 IS.

## Documentation

* [CLI Mode](docs/cli-mode.md)
* [Environment Specific Variables](docs/env-specific-variables.md)
* [Resource Propagation](docs/resource-propagation.md)
