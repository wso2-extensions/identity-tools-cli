# CLI Mode
The CLI mode can be used for bulk resource management. 

Usages:
* Export all/selected resources from an Identity Server to a local directory.
* Import all/selected resources from a local directory to an Identity Server.
* Propagate resources across multiple environments.
* Deploy new resources from resource configuration files to an Identity Server.
* Have a backup of resources in a local directory.

Currently, the supported resource types are:
* Applications
* Identity Providers

## Run the tool in CLI mode
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
The above command creates a new folder named ```configs```, which contains all the config files needed to setup the tool. 
The folder structure of the ```configs``` directory is as follows:
```
configs
└── env
    │── serverConfig.json
    └── toolConfig.json
``` 
   It is recommended to place the ```configs``` folder inside the local directory that is created to maintain the resource configuration files. 
   
Example local directory structure if multiple environments (dev, stage, prod) exist:
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
   │    │── app2.yml
   │    │── ... other exported app files
   │
   │── IdentityProviders
   │    │── idp1.yml
   │    │── idp2.yml
   │    │── ... other exported idp files
   │
   │── ... other resource types
   ```
   Use the ```--baseDir``` flag to specify the path to the local directory when creating the ```configs``` folder. If not specified, the tool creates the ```configs``` folder in the current directory.

### Server configurations
Server configurations are the configurations needed for connecting to the target environment. Server configurations can be provided through the ```serverConfig.json``` file or through environment variables. It is mandatory to provide the following parameters relevant to the target identity server to run the CLI commands.
* Server URL of the target identity server
* Client ID of a management application in the target IS
* Client Secret of a management application in the target IS
* Tenant Domain (optional)

These configurations differ from each environment and therefore should be maintained separately.  
#### Load server configurations from a file
Server configurations can be provided through the ```serverConfig.json``` file as a json object in the following format.

Example configurations:
```
{
   "SERVER_URL" : "https://localhost:9443",
   "CLIENT-ID" : "********",
   "CLIENT-SECRET" : "********",
   "TENANT-DOMAIN" : "carbon.super"
}
```
> **Note:** The CLI tool uses management rest apis of the IS to export and import resources. In order to perform these API requests, the client ID and client secret of a management application is required.
> 1. [Create an application](https://is.docs.wso2.com/en/6.1.0/guides/applications/register-sp) with **Management Application** enabled in the target IS.
> 2. Update Oauth inbound authentication configuration with a dummy callback URL and use the client ID and client secret for the above configurations.

> **Note:** Provide the required tenant domain from which the resources should be exported or imported. If the tenant domain is not provided, the tool uses the super tenant domain (carbon.super) by default.

In order to load these configurations from the ```serverConfig.json``` file, the ```--config``` flag should be used when running the exportAll/importAll commands specifying the path to the environment-specific config folder that contains the ```serverConfig.json``` file.

Example:
```
iamctl exportAll -c <path to the configs folder>/dev 
```
The tool performs the required action by selecting the target environment based on the path provided in the ```--config``` flag.
#### Load server configurations from environment variables
The server configurations can be provided through environment variables as well. 
If the ```--config``` flag is not used when running the exportAll/importAll commands, the tool looks for the server configurations in the following environment variables. 
* SERVER_URL
* CLIENT_ID
* CLIENT_SECRET
* TENANT_DOMAIN
* TOOL_CONFIG_PATH

> **Note:** The ```TOOL_CONFIG_PATH``` environment variable should be used to specify the path to the tool configs file. 

Example:
```
export SERVER_URL="https://localhost:9443"
```
```
export CLIENT_ID="********"
```
```
export CLIENT_SECRET="********"
```
```   
export TENANT_DOMAIN="carbon.super"
```
```
export TOOL_CONFIG_PATH="<path to the configs folder>/dev/toolConfig.json"
```
> **Note:** Before running the CLI commands, be sure to export the environment variables with the correct server details that the action should be performed against.
> 
> It is recommended to use the ```serverConfig.json``` file to provide the server configurations as it is more secure and easier to maintain when dealing with multiple environments.

### Tool configurations
The ```toolConfig.json``` file contains the configurations needed for overriding the default behaviour of the tool. 

Example configuration file:
```
{
   "KEYWORD_MAPPINGS" : {
      "CALLBACK_URL" : "https://demo.dev.io/callback",
      "ENV" : "dev"
   },
   
   "APPLICATIONS" : {
       "EXCLUDE" : ["App1", "App2"]
   },
   "IDENTITY_PROVIDERS" : {
       "INCLUDE_ONLY" : ["Idp1", "Idp2"],
       "EXCLUDE_SECRETS" : false
    }
}

```
The following properties can be configured through the tool configs to manage your resources.

#### Keyword replacement for environment-specific variables
The ```KEYWORD_MAPPINGS``` property can be used to replace environment specific variables in the exported resource configuration files with the actual values needed in the target environment. The keyword mapping should be added as a JSON object to the ```KEYWORD_MAPPINGS``` property in the tool configs in the following format.
```
{
   "KEYWORD_MAPPINGS" : {
      "<KEYWORD>" : "<VALUE>"
   }
}
```
Example:
```
{
   "KEYWORD_MAPPINGS" : {
      "CALLBACK_URL" : "https://demo.dev.io/callback"
   }
}
```
Find more information on the keyword replacement feature [here](../keyword-replacement.md).

#### Exclude resources
The ```EXCLUDE``` property can be used to exclude specific resources based on their name during import or export. The resources that need to be excluded can be added as an array of strings to the ```EXCLUDE``` property in tool configs under the relevant resource type.

Here is the format for adding the ```EXCLUDE``` property to the tool configs:
```
{
   "<RESOURCE_TYPE_NAME>" : {
      "EXCLUDE" : ["resource1", "resource2"]
   }
}
```

Example:
```
{
   "APPLICATIONS" : {
       "EXCLUDE" : ["App1", "App2"]
   }
}
```
#### Include only selected resources
The ```INCLUDE_ONLY``` property can be used to include only specific resources based on their name during import or export. The resources that need to be included can be added as an array of strings to the ```INCLUDE_ONLY``` property in tool configs under the relevant resource type.
```
{
   "RESOURCE_TYPE_NAME" : {
       "INCLUDE_ONLY" : ["resource1", "resource2"]
   }
}
```
Example:
```
{
   "APPLICATIONS" : {
       "INCLUDE_ONLY" : ["App1", "App2"]
   }
}
```
#### Exclude secrets from exported resources
By default, secrets are removed from the exported resources. For applications, the secret fields are not included in the exported file, and for identity providers the value of secrets will be masked by a string: ```'********'```.
The ```EXCLUDE_SECRETS``` config can be used to override this behaviour and include the secrets in the exported resources. 

The ```EXCLUDE_SECRETS``` property can be added to the tool configs under the relevant resource type as shown below.
```
{
   "RESOURCE_TYPE_NAME" : {
       "EXCLUDE_SECRETS" : false
   }
}
```
Example:
```
{
    "IDENTITY_PROVIDERS" : {
        "EXCLUDE_SECRETS" : false
    }   
}
```


## Commands
### ExportAll command
The ```exportAll``` command can be used to export all resources of all supported resource types from an Identity Server to a local directory.
```
iamctl exportAll -c <path to the env specific config folder> -o <path to the local output directory>
```
Use the ```--help``` flag to get more information on the command.
``` 
Flags:
  -c, --config string      Path to the env specific config folder
  -f, --format string      Format of the exported files (default "yaml")
  -h, --help               help for exportAll
  -o, --outputDir string   Path to the output directory
```
The ```--config``` flag can be used to provide the path to the env specific config folder that contains the ```serverConfig.json``` and ```toolConfig.json``` files with the details of the environment that needs the resources to be exported from. If the flag is not provided, the tool looks for the server configurations in the environment variables.

The ```--outputDir``` flag can be used to provide the path to the local directory where the exported resource configuration files should be stored. If the flag is not provided, the exported resource configuration files are created at the current working directory.

The ```--format``` flag defines the format of the exported resource configuration files. Currently, the tool supports only YAML format but will soon provide support for JSON and XML formats as well.

Running this command creates separate folders for each resource type at the provided output directory path. A new file is created with the resource name, in the given file format for each individual resource, under the relevant resource type folder.

Example local directory structure if multiple environments (dev, stage, prod) exist:
```
output directory
│── Applications
│    │── My app.yml
│    │── Pickup Manager.yml
│
│── IdentityProviders
│    │── Google.yml
│    │── Facebook.yml
│
│── ... other resource types
   ```

### ImportAll command
The ```importAll``` command can be used to import all resources of all supported resource types from a local directory to an Identity Server.
```
iamctl importAll -c <path to the env specific config folder> -i <path to the local input directory>
```
Use the ```--help``` flag to get more information on the command.
```
Flags:
  -c, --config string     Path to the env specific config folder
  -h, --help              help for importAll
  -i, --inputDir string   Path to the input directory
```
The ```--config``` flag can be used to provide the path to the env specific config folder that contains the ```serverConfig.json``` and ```toolConfig.json``` files with the details of the environment to which the resources should be imported. If the flag is not provided, the tool looks for the server configurations in the environment variables.

The ```--inputDir``` flag can be used to provide the path to the local directory where the resource configuration files are stored. If the flag is not provided, the tool looks for the resource configuration files at the current working directory.