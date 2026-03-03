# CLI Mode
The CLI mode can be used for bulk resource management. 

Usages:
* Export all/selected resources from a WSO2 IS to a local directory.
* Import all/selected resources from a local directory to a WSO2 IS.
* Promote resources across multiple environments.
* Deploy new resources from resource configuration files to a WSO2 IS.
* Have a backup of resources in a local directory.

Currently, the supported resource types are:
* Applications
* Identity Providers
* Claims
* User Stores
* OIDC Scopes

## Run the tool in CLI mode
To run the tool in CLI mode, follow the steps given below.

### Tool Initialization
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
    │── toolConfig.json
    └── keywordConfig.json
``` 
   It is recommended to place the ```configs``` folder inside the local directory that is created to maintain the resource configuration files. 
   
Example local directory structure if multiple environments (dev, stage, prod) exist:
   ```
   local directory
   │── configs
   │    │── dev
   │    │    │── serverConfig.json
   │    │    │── toolConfig.json
   │    │    └── keywordConfig.json
   │    │── stage
   │    │    │── serverConfig.json
   │    │    │── toolConfig.json
   │    │    └── keywordConfig.json
   │    └──── prod
   │         │── serverConfig.json
   │         │── toolConfig.json
   │         └── keywordConfig.json
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
   "CLIENT_ID" : "********",
   "CLIENT_SECRET" : "********",
   "TENANT_DOMAIN" : "carbon.super"
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
* KEYWORD_CONFIG_PATH

> **Note:** The ```TOOL_CONFIG_PATH``` and ```KEYWORD_CONFIG_PATH``` environment variables should be used to specify the path to the tool configs file and keyword config file respectively. 

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
```
export KEYWORD_CONFIG_PATH="<path to the configs folder>/dev/keywordConfig.json"
```
> **Note:** Before running the CLI commands, be sure to export the environment variables with the correct server details that the action should be performed against.
> 
> It is recommended to use the ```serverConfig.json``` file to provide the server configurations as it is more secure and easier to maintain when dealing with multiple environments.

#### Using environment variables in serverConfig.json
You can also explicitly specify the use of environment variables for certain configurations in the ```serverConfig.json``` file itself. To do this, use the placeholder ```${YOUR_ENV_VAR_NAME}``` in the ```serverConfig.json``` file, as shown in the following example:
```
{
  "CLIENT_ID": "${DEV_CLIENT_ID}",
  "CLIENT_SECRET": "${DEV_CLIENT_SECRET}",
  "SERVER_URL": "https://localhost:9443",
  "TENANT_DOMAIN": "carbon.super"
}
```
The tool will search for the keyword with the name given inside the placeholder in the environment and use its value instead.

### Tool configurations
The ```toolConfig.json``` file contains the configurations needed for overriding the default behaviour of the tool. 

Example configuration file:
```
{
   "ALLOW_DELETE" : true,
   "EXCLUDE" : ["Claims"],
   "APPLICATIONS" : {
       "EXCLUDE" : ["App1", "App2"]
   },
   "IDENTITY_PROVIDERS" : {
       "INCLUDE_ONLY" : ["Idp1", "Idp2"],
       "EXCLUDE_SECRETS" : false
   },
   "USERSTORES" : {
       "EXCLUDE" : ["US1", "US2"]
   },
   "CLAIMS" : {
       "INCLUDE_ONLY" : ["local"]
   },
   "OIDC_SCOPES" : {
       "EXCLUDE" : ["Scope1"]
   }
}
```
The following properties can be configured through the tool configs to manage your resources.
#### Exclude resources
The ```EXCLUDE``` property can be used to exclude a specific resource type during import or export. The resource types that need to be excluded can be added as an array of strings to the ```EXCLUDE``` property in tool configs. 

The ```EXCLUDE``` property can also be used to exclude specific resources based on their name during import or export. These should be specified under the relevant resource type.

Here is the format for adding the ```EXCLUDE``` property to the tool configs:
```
{
   "EXCLUDE" : ["resourceType1", "resourceType2"]
   "<RESOURCE_TYPE_NAME>" : {
      "EXCLUDE" : ["resource1", "resource2"]
   }
}
```

Example:
```
{
   "EXCLUDE" : ["IdentityProviders", "UserStores"]
   "APPLICATIONS" : {
       "EXCLUDE" : ["App1", "App2"]
   }
}
```
#### Include only selected resources
The ```INCLUDE_ONLY``` property can be used to include only specific resource types during import or export. The resource types that need to be included can be added as an array of strings to the ```INCLUDE_ONLY``` property in tool configs.

The ```INCLUDE_ONLY``` property can also be used to include only specific resources based on their name during import or export. These should be specified under the relevant resource type.
```
{
   "INCLUDE_ONLY" : ["resourceType1", "resourceType2"]
   "RESOURCE_TYPE_NAME" : {
       "INCLUDE_ONLY" : ["resource1", "resource2"]
   }
}
```
Example:
```
{
   "INCLUDE_ONLY" : ["Applications", "Claims"]
   "APPLICATIONS" : {
       "INCLUDE_ONLY" : ["App1", "App2"]
   }
}
```
> **Note:** When both EXCLUDE and INCLUDE_ONLY properties are used, INCLUDE_ONLY takes precedence over EXCLUDE.

#### Exclude secrets from exported resources
By default, secrets fields are masked by a string: ```'********'```.
The ```EXCLUDE_SECRETS``` config can be used to override this behaviour and include the secrets in the exported resources. 

> **Note:** This config cannot be used to include secrets for userstores. The secrets of userstores will always be masked by the string: ```'********'```
> 
The ```EXCLUDE_SECRETS``` property can be added to the tool configs globally ```or``` under the relevant resource type as shown below.
```
{
   "EXCLUDE_SECRETS" : false
   "RESOURCE_TYPE_NAME" : {
       "EXCLUDE_SECRETS" : true
   }
}
```
Example:
```
{
   "EXCLUDE_SECRETS" : false
    "IDENTITY_PROVIDERS" : {
        "EXCLUDE_SECRETS" : true
    }   
}
```

#### Allow deleting resources
By default, the tool does not delete any resources during export or import. During export, the deletion of a resource in the target environment will not delete the corresponding resource file in the local directory. The file will have to be deleted manually. Similarly, during import, the deletion of a resource file in the local directory will not delete the corresponding resource in the target environment. 
The ```ALLOW_DELETE``` property can be used to override this behavior and allow the tool to delete resources.

```
{
    "ALLOW_DELETE" : true
}
```
> **Caution:** Use this property cautiously, as it can delete required resources if misconfigured.
> If using this config, make sure to exclude the resources that should not be deleted using the ```EXCLUDE``` property.
>
> Ex: Applications - "Console", "My Account", Management application created for the tool, etc.
> Identity Providers - Resident Identity Provider, etc.

Example:
```
{
   "KEYWORD_MAPPINGS" : {
      "CALLBACK_URL" : "https://demo.dev.io/callback"
   },
    "ALLOW_DELETE" : true,
    
    "APPLICATIONS" : {
        "EXCLUDE" : ["Console", "My Account", "Dev-mgt-app"]
    },
    "IDENTITY_PROVIDERS" : {
        "EXCLUDE" : "LOCAL"
    }  
}
```

> **Note:** Configurations under a particular resource type will take precedence over the global configurations for that resource type.

### Keyword Mapping configurations
The ```keywordConfig.json``` file contains the configurations needed for keyword replacement for environment-specific variables.

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
> **Note:** Keyword mappings can also be incorporated as environment variables.

Find more information on the keyword replacement feature [here](../keyword-replacement.md).

## Commands
### ExportAll command
The ```exportAll``` command can be used to export all resources of all supported resource types from a WSO2 IS to a local directory.
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
The ```--config``` flag can be used to provide the path to the env specific config folder that contains the ```serverConfig.json```,  ```toolConfig.json```, and ```keywordConfig.json``` files with the details of the environment that needs the resources to be exported from. If the flag is not provided, the tool looks for the server configurations in the environment variables.

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
The ```importAll``` command can be used to import all resources of all supported resource types from a local directory to a WSO2 IS.
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
The ```--config``` flag can be used to provide the path to the env specific config folder that contains the ```serverConfig.json```, ```toolConfig.json```, and ```keywordConfig.json``` files with the details of the environment to which the resources should be imported. If the flag is not provided, the tool looks for the server configurations in the environment variables.

The ```--inputDir``` flag can be used to provide the path to the local directory where the resource configuration files are stored. If the flag is not provided, the tool looks for the resource configuration files in the current working directory.

## Supported resource types
The tool supports the following resource types:

### Applications
The tool supports exporting and importing applications. The exported application configuration files can be found under the ```Applications``` folder in the local directory. If it is required to deploy a new application through the `import` command of the tool, the new file should be placed under the ```Applications``` folder in the local directory.

Since the ```Console``` and ```My Account``` are read-only system applications in WSO2 Identity Server, if it is required to update these applications through the `import` command, add the following configurations to the ```deployment.toml``` file and restart the server.
```
[system_applications]
read_only_apps = []
```

> **Caution:** Be cautious when updating the system applications: ```Console``` and ```My Account``` through the tool, since it will result in unexpected errors in these apps if edited incorrectly. It is recommended to exclude the ```Console```, ```My Account``` and the Management application created for the tool during normal usage, unless it is required to update them through the tool.

### Identity providers
The tool supports exporting and importing identity providers. The exported identity provider configuration files can be found under the ```IdentityProviders``` folder in the local directory. If it is required to deploy a new identity provider through the import command of the tool, the new file should be placed under the ```IdentityProviders``` folder in the local directory.

The resident identity provider can also be exported into a file named ```LOCAL``` and can be updated by modifying the ```LOCAL``` file and using the import command.

> **Caution:** Be cautious when updating the resident identity provider through the ```LOCAL``` file since it will result in unexpected errors in the server if edited incorrectly. It is recommended to exclude the ```LOCAL``` file during normal usage unless it is required to update the resident identity provider through the tool.

### User stores
The tool supports exporting and importing secondary user stores. The exported user store configuration files can be found under the ```UserStores``` folder in the local directory. If it is required to deploy a new user store through the import command of the tool, the new file should be placed under the ```UserStores``` folder in the local directory.
By default, the tool masks the secrets of the user stores in the exported files. Make sure to add the correct values for the masked fields (connection password, etc.) during import, to properly deploy the user stores.