# IAM-CTL

The IAM-CTL is a tool that can be used to manage WSO2 Identity Server configurations from the command line. It is written in Go and uses the WSO2 Identity Server management REST APIs to manage configurations.

### Pre-Requisites
* WSO2 Identity Server 6.1.0 

### How to run the tool 

1. Download the latest release binary file from [Releases](https://github.com/wso2-extensions/identity-tools-cli/releases)
 based on your Operating System.
2. Extract the tar or zip file.
Here onwards the extracted directory path is referred as ```<IAM-CTL-PATH>```.
3. Open a terminal and create an alias for the IAM-CTL executable file using following command according to your platform.
* linux/mac:
 
    ```
    alias iamctl="<IAM-CTL-PATH>/bin/iamctl" 
    ```

* windows

    ```
    doskey iamctl=<IAM-CTL-PATH>/bin/iamctl.exe $*
    ```
 
5. Run the tool using the following command to get the basic details.
    ```
    iamctl -h
    ```
6. Start the IS server and [create an application](https://is.docs.wso2.com/en/6.1.0/guides/applications/register-sp) with **Management Application** enabled.
7. Configure Oauth inbound authentication configuration with a dummy callback URL and make note of the client ID and client secret.


The IAM CTL can be used in two basic modes.
## CLI Mode

The CLI mode can be used to handle bulk configurations in the target environment. This can be used to propagate resources across multiple environments, deploy new configurations to target environments and will act as a backup of each environment's configurations.

This mode consists of the exportAll and importAll commands which can be used to export and import all configurations of the supported resources types from or to a target environment. 

Currently, supported resource types are: 
* Applications

### Running the tool in the CLI mode
The following steps will provide the basic steps to run the tool in the simplest way. Find more comprehensive details about the commands used in the CLI mode [here](docs/cli-mode.md).

#### Tool Initialization
The tool should be initialized against the environment it is run against.
1. Create a new folder and open a terminal inside it.
2. Run the following command to create the configuration files needed to initialize the tool.
    ```
    iamctl setupCLI
    ```
3. A new folder named ```configs``` will be created with a ```env``` folder inside it, having the 2 configuration files ```serverConfig.json``` and ```toolConfig.json```.
> **Note:** If you have multiple environments, copy the ```env``` folder and rename it according to the environments you have. For example, if you have 2 environments: dev and prod, have 2 separate config folders ```dev``` and ```prod```. 
4. Open the ```serverConfig.json``` file and provide the IS server details and client ID/secret of the app you created earlier.

Example configurations:

    ```
    "SERVER_URL" : "https://localhost:9443",
    "CLIENT-ID" : "********",
    "CLIENT-SECRET" : "********",
    "USERNAME" : "admin",
    "PASSWORD" : "admin",
    "TENANT-DOMAIN" : "carbon.super"
    ```

#### Export
Run the following command to export all supported configurations from the target environment to the current directory.
```
iamctl exportAll -c ./configs/env
```
A new folder named ```Applications``` will be created with exported yaml files for each application available in the IS server.

#### Import
Run the following command to import all supported configurations from the current directory to the target environment.
```
iamctl importAll -c ./configs/env
```
All applications available inside the ```Applications``` folder in the current directory will be imported to the IS server.

## Interactive Mode
The interactive mode can be used to handle application configurations in an interactive manner. This can be used to add, list, export and import applications in the target environment.
> Note: This mode does not provide support for bulk resource export or import.

### Running the tool in the interactive mode
#### Tool Initialization
1. Run the following command to initialize the tool by providing the Identity server details and client ID/secret of the app you created earlier.
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
* To add application, select ```Add application``` option and proceed by providing the application details.

* To view list of applications, select ```Get List``` option.

### Application related commands
After initializing the tool, the following commands can be used to add, list, export and import applications in the target environment.
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