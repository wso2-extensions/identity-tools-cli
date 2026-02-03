# IAM-CTL

IAM-CTL is a tool that can manage WSO2 Identity Server configurations from the command line. It is written in [GO](https://go.dev/) and uses the management REST APIs of WSO2 Identity Server to manage configurations.

Refer IS Official documentation for further information [Promote configurations between environments](https://is.docs.wso2.com/en/latest/deploy/promote-configurations/)

### Prerequisites
You need to [setup](https://is.docs.wso2.com/en/latest/get-started/sample-use-cases/set-up/) WSO2 Identity Server.

### Run the tool

1. Download the latest binary file from [Releases](https://github.com/wso2-extensions/identity-tools-cli/releases) based on your Operating System.

2. Extract the `tar` or `zip` file.

    Here onwards, the extracted directory path is referred to as `<IAM-CTL-PATH>`.

3. Open a terminal and create an alias for the `IAM-CTL` executable file using one of the following commands (depending on your platform):
   * linux/mac:
 
       ```
       alias iamctl="<IAM-CTL-PATH>/bin/iamctl" 
       ```

   * windows (On Command Prompt):

       ```
       doskey iamctl="<IAM-CTL-PATH>\bin\iamctl.exe" $*
       ```
 
4. Run the tool using the following command to get the basic details.
    ```
    iamctl -h
    ```
   
### Registering an Application in WSO2 Identity Server

#### For Export and Import in the Root Organization
1. Start WSO2 Identity Server.
2. Open the Console application.
3. Login as the admin user (admin/admin).
4. [Register a M2M application](https://is.docs.wso2.com/en/latest/guides/applications/register-machine-to-machine-app/).
5. Grant the following API authorizations under Management APIs.

API                                              | Scopes
------------------------------------------------ | --------------------------------------------------------------------
Management --> Application Management API        | Create Application, Update Application, Delete Application, View Application, Update authorised business APIs of an Application, Update authorised internal APIs of an Application, View application client secret, Regenerate application client secret
Management --> Application Authentication Script Management API | Update Application Authentication Script
Management --> Claim Management API              | Create Claim, Update Claim, Delete Claim, View Claim
Management --> Identity Provider Management API  | Create Identity Provider, Update Identity Provider, Delete Identity Provider, View Identity Provider
Management --> Userstore Management API          | Create Userstore, Update Userstore, Delete Userstore, View Userstore

6. Take note of the client ID and client secret of this application.

#### For Export and Import in a Sub-Organization
1. WSO2 Identity Server.
2. Open the Console application.
3. Login as the admin user (admin/admin) of the root organization.
4. [Register a Standard-Based Application](https://is.docs.wso2.com/en/latest/guides/applications/register-standard-based-app/) in the root organization.
5. Share the application with the relevant sub-organization (e.g., `wso2.com`).
6. Allow following grant types in the newly created Standard-Based Application:
   * Client Credentials
   * Organization Switch
7. Grant the following API authorizations under Organization APIs.

API                                              | Scopes
------------------------------------------------ | --------------------------------------------------------------------
Organization --> Application Management API        | Create Application, Update Application, Delete Application, View Application, View application client secret, Regenerate application client secret
Organization --> Application Authentication Script Management API | Update Application Authentication Script
Organization --> Identity Provider Management API  | Create Identity Provider, Update Identity Provider, Delete Identity Provider, View Identity Provider
Organization --> Userstore Management API          | Create Userstore, Update Userstore, Delete Userstore, View Userstore

8. Take note of the client ID and client secret of this application from the root organization.


## CLI mode

The CLI mode of the tool can be used to handle bulk configurations in the target environment. This can be used to promote resources across multiple environments, deploy new configurations to target environments, and act as a backup of each environment's configurations.

This mode consists of the `exportAll` and `importAll` commands that can be used to export and import all configurations of the supported resource types from or to a target environment. This can be used to transfer resources across sub organisations. 

The supported resource types to transfer resources between root organizations are: 
* Applications
* Identity Providers
* Claims
* User Stores

The supported resource types to transfer resources between sub organizations are:
* Applications
* Identity Providers
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

*Root Organization*

```
    {
       "SERVER_URL" : "https://localhost:9443",
       "CLIENT_ID" : "********",
       "CLIENT_SECRET" : "********",
       "TENANT_DOMAIN" : "carbon.super",
       "SERVER_VERSION" : "7.2.0"
    }
```

*Sub Organization*

```
    {
       "SERVER_URL" : "https://localhost:9443",
       "CLIENT_ID" : "********",
       "CLIENT_SECRET" : "********",
       "TENANT_DOMAIN" : "carbon.super",
       "SERVER_VERSION" : "7.2.0",
       "ORGANIZATION": "b833d7de-264c-4c4e-8d52-61f9c57e84ca"
    }
```

#### Export
Run the following command to export all supported resource configurations from the target environment to the current directory.

* linux/mac:

    ```
    iamctl exportAll -c ./configs/env
    ```

 * windows (On Command Prompt):

    ```
    iamctl exportAll -c .\configs\env
    ```

A new set of folders are created, which are named after each resource type, with exported yaml files for each available resource in WSO2 IS.


#### Import
Run the following command to import all supported resource configurations from the current directory to the target environment.

* linux/mac:

    ```
    iamctl importAll -c ./configs/env
    ```

* windows (On Command Prompt):

   ```
   iamctl importAll -c .\configs\env
   ```
  
All resources available inside each resource type folder in the current directory will be imported to WSO2 IS.

## Documentation

* [CLI Mode](docs/cli-mode.md)
* [Environment Specific Variables](docs/env-specific-variables.md)
* [Resource Propagation](docs/resource-propagation.md)
