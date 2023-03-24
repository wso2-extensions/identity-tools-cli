# Handling Environment Specific Variables
When maintaining multiple environments, it is common to have environment specific variables in the resource configuration files. The CLI tool supports dynamic keyword replacement for such environment specific variables during import or export.

###Adding keyword placeholders and keyword mappings
1. Find the environment specific values in the resource configuration files and add a keyword placeholder.
Use the syntax ```{{KEYWORD}}``` to add the keyword placeholder in the resource configuration files.

Example:
If an exported file from the dev environment contains the following value for the ```callbackUrl``` field of an application:
``` 
applicationName: Demo App
callbackUrl: https://demo.dev.io/commonauth
```
A keyword placeholder can be added as follows:
```
applicationName: Demo App
callbackUrl: https://{{CALLBACK_DOMAIN}}/commonauth
```
2. Add the keyword mapping to the tool configs file in each environment.
Example toolConfig.json file in ```config/dev``` directory:
```
{
    "KEYWORD_MAPPINGS" : {
        "CALLBACK_DOMAIN" : "demo.dev.io"
    }
}
```
Example toolConfig.json file in ```config/prod``` directory:
```
{
    "KEYWORD_MAPPINGS" : {
        "CALLBACK_DOMAIN" : "demo.prod.io"
    }
}
```
When importing the resource from the local directory to the prod environment, the keyword ```CALLBACK_DOMAIN``` will be replaced with the value ```demo.prod.io``` and the resource callback url will be updated to ```https://demo.prod.io/commonauth```.

### Recommended workflow
1. Use the CLI tool to export once from the lowest environment and create the local resource configuration directory.
2. Add the keyword placeholders to the exported files and add the relevant keyword mapping to the tool configs of each environment.
3. Use the CLI tool to import the resources from the local directory to higher environments with the replaced keyword values.

> **Note:** If it is required to export again from any environment and update the local resource configurations, there is a chance that the manually added keyword placeholders will get replaced, if the exported value of the keyword value is different. 
> In such cases, a warning will be issued with the details of the removed keyword, and it is recommended to add the keyword placeholders again and update the keyword mappings in the tool configs.

## Advanced Keyword Mapping Configurations

As mentioned above, when dealing with multiple environments we will have to add keyword placeholders and keyword mappings to environment specific variables. 
If some property value is both environment specific and resource specific, you can add a separate keyword mapping for each resource where the default keyword mapping should be overridden.

``` 
"KEYWORD_MAPPINGS" : {
    "KEYWORD1" : "default value",
    <RESOURCE_TYPE_NAME> : {
        "RESOURCE_NAME" : {
            "KEYWORD1" : "resource specific value"
            }
        }
    }
    
   
}
```
The resource specific keyword mapping can be added as a special configuration with the same placeholder name, but during execution the default keyword mapping will be overridden with the resource specific keyword mapping only for that resource.

Example:
If there are 5 applications that needs to imported to a target environment, and the callback url of 4 of them in the prod environment should be ```https://demo.prod.io/callback``` and the callback url of the 5th application (App5) should be ```https://demo.prod.io/callback2```, the keyword mapping can be added as follows:
```
{
    "KEYWORD_MAPPINGS" : {
        "CALLBACK_URL" : "https://demo.prod.io/callback",
    },
    "APPLICATIONS" : {
        "App5" : {
            "CALLBACK_URL" : "https://demo.prod.io/callback2"
        }
    }
}
``` 
Here, during import to the prod environment, the callback url of the 4 applications will be replaced with ```https://demo.prod.io/callback``` and the callback url of the 5th application will be replaced with ```https://demo.prod.io/callback2```.
