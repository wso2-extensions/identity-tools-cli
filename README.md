# IAM-CTL

IAM-CTL is a tool that can manage WSO2 Identity Server configurations from the command line. It is written in [GO](https://go.dev/) and uses the management REST APIs of WSO2 Identity Server to manage configurations.

### Prerequisites
You need to [setup](https://is.docs.wso2.com/en/6.0.0/get-started/sample-use-cases/set-up/) WSO2 Identity Server 6.1.0.

### Run the tool

1. Download the latest binary file from [Releases](https://github.com/wso2-extensions/identity-tools-cli/releases).
 based on your Operating System.

2. Extract the `tar` or `zip` file.

    Here onwards, the extracted directory path is referred to as `<IAM-CTL-PATH>`.

3. Open a terminal and create an alias for the `IAM-CTL` executable file using one of the following commands (depending on your platform):
   * linux/mac:
 
       ```
       alias iamctl="<IAM-CTL-PATH>/bin/iamctl" 
       ```

   * windows

       ```
       doskey iamctl=<IAM-CTL-PATH>/bin/iamctl.exe $*
       ```
 
4. Run the tool using the following command to get the basic details.
    ```
    iamctl -h
    ```
5. Start WSO2 IS and [create an application](https://is.docs.wso2.com/en/6.1.0/guides/applications/register-sp) with **Management Application** enabled.
6. Update the Oauth inbound authentication configuration with a dummy callback URL and take note of the client ID and client secret.


The IAM CTL can be used in two basic modes.
## CLI mode

The CLI mode can be used to handle bulk configurations in the target environment. This can be used to propagate resources across multiple environments, deploy new configurations to target environments, and act as a backup of each environment's configurations.

This mode consists of the `exportAll` and `importAll` commands that can be used to export and import all configurations of the supported resource types from or to a target environment. 

Currently, the supported resource types are: 
* Applications
* Identity Providers
* Claims

### Running the tool in the CLI mode
The following explains the basic steps for running the tool in the simplest way. Find more comprehensive details about the commands used in the CLI mode [here](docs/cli-mode.md).

#### Tool initialization
The tool should be initialized with the server details of the environment it is run against.
1. Create a new folder and navigate to it from your terminal.
2. Run the following command to create the configuration files needed to initialize the tool.
    ```
    iamctl setupCLI
    ```
3. A new folder named ```configs``` will be created with an ```env``` folder inside it. The `env` folder contains two configuration files: ```serverConfig.json``` and ```toolConfig.json```.
> **Note:** If you have multiple environments, get a copy of the ```env``` folder and rename it according to the environments you have. For example, if you have two environments: dev and prod, have two separate config folders as ```dev``` and ```prod```. 
4. Open the ```serverConfig.json``` file and provide the WSO2 IS details and client ID/secret of the app you created earlier.

Example configurations:

    ```
    "SERVER_URL" : "https://localhost:9443",
    "CLIENT-ID" : "********",
    "CLIENT-SECRET" : "********",
    "TENANT-DOMAIN" : "carbon.super"
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

## Interactive Mode
The interactive mode can be used to handle application configurations in an interactive manner. This can be used to add, list, export, and import applications in the target environment.
> Note: This mode does not provide support for bulk resource export or import.

### Run the tool in interactive mode

See the topics given below to run the tool in interactive mode.
#### Tool Initialization
1. Run the following command to initialize the tool by providing details of WSO2 IS and the client ID/secret of the app you created earlier.
```
iamctl init
```
Provide the details as prompted by the tool.
```
:~$ iamctl init
  ___      _      __  __            ____   _____   _     
 |_ _|    / \    |  \/  |          / ___| |_   _| | |    
  | |    / _ \   | |\/| |  _____  | |       | |   | |    
  | |   / ___ \  | |  | | |_____| | |___    | |   | |___ 
 |___| /_/   \_\ |_|  |_|          \____|   |_|   |_____|
      
? Enter IAM URL [<schema>://<host>]: https://localhost:9443                                                   
? Enter clientID: *******
? Enter clientSecret: *******
? Enter Tenant domain: carbon.super
```
2. Run the following command to provide admin user credentials.
```
iamctl serverConfiguration
```
```
~$ iamctl serverConfiguration
? Enter IAM URL [<schema>://<host>]: https://localhost:9443
? Enter Username: admin
? Enter Password: admin
```

### Create new applications and list existing applications interactively
Run the following command to get the options available for applications.
```
iamctl application
```
```
$ iamctl application                                                      
? Select the option to move on:  [Use arrows to move, type to filter]
> Add application
  Get List
  Exit
```
* To add an application, select ```Add application``` and proceed by providing the application details.

* To view the list of applications, select ```Get List```.

### Application-related commands
After initializing the tool, the following commands can be used to add, list, export, and import applications in the target environment.
#### Add application
```
iamctl application add -n=TestApplication 
```
#### List applications
```
iamctl application list
```
#### Export application
```
iamctl application export -s <applicationID> -p <path-to-export-location>
```
#### Import application
```
iamctl application import -i <path-to-import-file>
```
Find more comprehensive details about the commands used in the interactive mode [here](docs/interactive-mode.md).

## Documentation

* [CLI Mode](docs/cli-mode.md)
* [Interactive Mode](docs/interactive-mode.md)
* [Environment Specific Variables](docs/env-specific-variables.md)
* [Resource Propagation](docs/resource-propagation.md)