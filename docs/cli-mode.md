# CLI Mode
The CLI mode can be used for bulk resource management. 

Usages:
* Export all/selected resources from an Identity Server to a local directory.
* Import all/selected resources from a local directory to an Identity Server.
* Propagate resources across multiple environments.
* Deploy resources from resource configuration files to an Identity Server.
* Have a backup of resources in a local directory.

Currently, supported resource types are:
* Applications

## Running the Tool in the CLI Mode
### Initialization
Before running the tool to export or import any resource, the tool should be configured against the target environment.

First, create the config files using the ```setupCLI``` command:

    ```
   iamctl setupCLI
   ```
   Use the ```--help``` flag to get more information on the command.
   ```
   Flags:
    -d, --baseDir string   Path to the base directory
    -h, --help             help for setupCLI
   ```
   The above command will create a new folder named ```configs``` which will contain all the config files needed to setup the tool. 
    The folder structure of the ```configs``` directory will be as follows:
    ```
    configs
    └── env
        │── serverConfig.json
        └── toolConfig.json
    ``` 
   It is recommended to place the ```configs``` folder inside local directory created to maintain the resource configuration files. 
   Example local directory structure if multiple environments: dev, stage, prod exist:
    ```
   local directory
   │── configs
   │    │── dev
   │    │    │── serverConfig.json
   │    │    └── toolConfig.json
   │    │── stage
   │    │    │── serverConfig.json
   │    │    └── toolConfig.json
   │    └──── prod
   │         │── serverConfig.json
   │         └── toolConfig.jsondev
   │── Applications
   │    │── app1.yml
   │    │── ... other exported app files
   │── ... other resource types
   ```
   Use the ```--baseDir``` flag to specify the path to the local directory when creating the ```configs``` folder.

### Server Configurations
The ```serverConfig.json``` file contains the server details of the target environment. It is mandatory to provide the details of the relevant IS server, as requested in the file to run the CLI commands.

Example configurations:
```
"SERVER_URL" : "https://localhost:9443",
"CLIENT-ID" : "********",
"CLIENT-SECRET" : "********",
"USERNAME" : "admin",
"PASSWORD" : "admin",
"TENANT-DOMAIN" : "carbon.super"
```
> **Note:** The CLI tool uses management rest apis of the IS, to export and import resources. In order to perform these API requests, the client ID and client secret of a management application is required.
> 1. [Create an application](https://is.docs.wso2.com/en/6.1.0/guides/applications/register-sp) with **Management Application** enabled.
> 2. Configure Oauth inbound authentication configuration with a dummy callback URL and use the client ID and client secret for the above configurations.

### Tool Configurations
The ```toolConfig.json``` file contains the configurations needed to override the default behaviour of the tool. 

The following properties can be configured though the tool configs to manage your resources.

#### Keyword Replacement for Environment Specific Variables
The ```KEYWORD_MAPPINGS``` property can be used to replace environment specific variables in the exported resource configuration files, with the actual values needed in the target environment. The keyword mapping should be added as a JSON object to the ```KEYWORD_MAPPINGS``` property in tool configs in the following format.
```
"KEYWORD_MAPPINGS" : {
    "<KEYWORD>" : "<VALUE>"
}
```
Example:
```
"KEYWORD_MAPPINGS" : {
    "CALLBACK_URL" : "https://demo.dev.io/callback"
}
```
Find more information on the keyword replacement feature [here](../keyword-replacement.md).

#### Exclude Resources
The ```EXCLUDE``` property can be used to exclude specific resources based on their name during import or export. The resources that needs to be excluded can be added as an array of strings to the ```EXCLUDE``` property in tool configs under the relevant resource type.

Here is the format of adding the ```EXCLUDE``` property to the tool configs:
```
"<RESOURCE_TYPE_NAME>" : {
    "EXCLUDE" : ["resource1", "resource2"]
}
```

Example:
```
"APPLICATIONS" : {
    "EXCLUDE" : ["App1", "App2"]
}
```
#### Include Only Selected Resources
The ```INCLUDE_ONLY``` property can be used to include only specific resources based on their name during import or export. The resources that needs to be included can be added as an array of strings to the ```INCLUDE_ONLY``` property in tool configs under the relevant resource type.
```
"RESOURCE_TYPE_NAME" : {
    "INCLUDE_ONLY" : ["resource1", "resource2"]
}
```
Example:
```
"APPLICATIONS" : {
    "INCLUDE_ONLY" : ["App1", "App2"]
}
```

## Commands
### ExportAll Command
The ```exportAll``` command can be used to export all resources of all supported resource types from an Identity Server to a local directory.
```
iamctl exportAll -c <path to the env specific config folder> -d <path to the local directory>
```
Use the ```--help``` flag to get more information on the command.
``` 
Flags:
  -c, --config string      Path to the env specific config folder
  -f, --format string      Format of the exported files (default "yaml")
  -h, --help               help for exportAll
  -o, --outputDir string   Path to the output directory
```
The ```--config``` flag is mandatory, and it should provide the path to the env specific config folder which contains the serverConfig.json and toolConfig.json files with the details of the environment that needs the resources to be exported from.

The ```--outputDir``` flag is optional, and it should provide the path to the local directory where the exported resource configuration files should be stored. If the flag is not provided, the exported resource configuration files will be created at the current working directory.

The ```--format``` flag defines the format of the exported resource configuration files. Currently, the tool supports YAML, XML and JSON formats and will default to YAML if the flag is not provided.

Running this command will create separate folders for each resource type at the provided output directory path. A new file will be created with the resource name and in the given file format for each individual resource, under the relevant resource type folder.
For example:
Example local directory structure if multiple environments: dev, stage, prod exist:
```
output directory
│── Applications
│    │── My app.yml
│    │── Pickup Manager.yml
│── ... other resource types
   ```

### ImportAll Command
The ```importAll``` command can be used to import all resources of all supported resource types from a local directory to an Identity Server.
```
iamctl importAll -c <path to the env specific config folder> -i <path to the local directory>
```
Use the ```--help``` flag to get more information on the command.
```
Flags:
  -c, --config string     Path to the env specific config folder
  -h, --help              help for importAll
  -i, --inputDir string   Path to the input directory
```
The ```--config``` flag is mandatory, and it should provide the path to the env specific config folder which contains the serverConfig.json and toolConfig.json files with the details of the environment that needs the resources to be imported to.

The ```--inputDir``` flag is optional, and it should provide the path to the local directory where the resource configuration files are stored. If the flag is not provided, the tool will look for the resource configuration files at the current working directory.