# Handling environment-specific variables
When maintaining multiple environments, it is common to have environment specific variables in the resource configuration files. The CLI tool supports dynamic keyword replacement for such environment-specific variables during import or export.

### Adding keyword placeholders and keyword mappings
1. Find the environment specific values in the resource configuration files and add a keyword placeholder.
Use the syntax ```{{KEYWORD}}``` to add the keyword placeholder in the resource configuration files.

Example:
Consider an example where an exported file from the dev environment contains the following value for the ```callbackUrl``` field of an application:
``` 
applicationName: Demo App
callbackUrl: https://demo.dev.io/commonauth
```
In the above example, a keyword placeholder can be added as follows:
```
applicationName: Demo App
callbackUrl: https://{{CALLBACK_DOMAIN}}/commonauth
```
2. Add the keyword mapping to the keyword configs file in each environment.
Example `keywordConfig.json` file in the ```config/dev``` directory:
```
{
    "KEYWORD_MAPPINGS" : {
        "CALLBACK_DOMAIN" : "demo.dev.io"
    }
}
```
Example `keywordConfig.json` file in the ```config/prod``` directory:
```
{
    "KEYWORD_MAPPINGS" : {
        "CALLBACK_DOMAIN" : "demo.prod.io"
    }
}
```
When importing the resource from the local directory to the prod environment, the ```CALLBACK_DOMAIN``` keyword is replaced with the value ```demo.prod.io``` and the resource callback url is updated to ```https://demo.prod.io/commonauth```.

### Incorporate keyword mappings as environment variables
You can also incorporate keyword mappings as environment variables using a similar approach. In the ```keywordConfig.json``` file, you can add ```${}``` placeholders for the keyword values. The tool will search for the keyword with the name given inside the placeholder in the environment and use its value instead.

Example:
```
{
    "KEYWORD_MAPPINGS" : {
        "CALLBACK_DOMAIN" : "${DEV_CALLBACK_DOMAIN}"
    }
}
```
In the above example, the tool will look for the value of the ```CALLBACK_DOMAIN``` keyword in the corresponding environment variable named ```DEV_CALLBACK_DOMAIN``` and use that value instead.

Make sure to set the environment variable ```DEV_CALLBACK_DOMAIN``` with the appropriate value before running the CLI commands.

### Recommended workflow
1. Use the CLI tool to export once from the lowest environment and create the local resource configuration directory.
2. Add the keyword placeholders to the exported files and add the relevant keyword mapping to the keyword configs of each environment.
3. Use the CLI tool to import the resources from the local directory to higher environments with the replaced keyword values.

> **Note:** If it is required to export again from any environment and update the local resource configurations, there is a chance that the manually added keyword placeholders will get replaced if the exported keyword value is different. 
> In such cases, a warning is issued with details of the removed keyword. It is recommended to add the keyword placeholders again and update the keyword mappings in the keyword configs.

## Advanced keyword mapping configurations

As mentioned above, when dealing with multiple environments, we have to add keyword placeholders and keyword mappings to environment-specific variables. 
If some property value is both environment-specific and resource-specific, you can add a separate keyword mapping for each resource where the default keyword mapping should be overridden.

``` 
{
    "KEYWORD_MAPPINGS" : {
        "KEYWORD1" : "default value",
    },
    
    <RESOURCE_TYPE_NAME> : {
        "RESOURCE_NAME" : {
            "KEYWORD_MAPPINGS" : {
                "KEYWORD1" : "resource specific value",
            }    
        }
    }
}
```
The resource-specific keyword mapping can be added as a special configuration with the same placeholder name, but during execution the default keyword mapping is overridden with the resource-specific keyword mapping only for that resource.

Example:
If there are five applications that need to be imported to a target environment, and the callback URL of four of them in the prod environment should be ```https://demo.prod.io/callback``` and the callback URL of the fifth application (App5) should be ```https://demo.prod.io/callback2```, the keyword mapping can be added as follows:
```
{
    "KEYWORD_MAPPINGS" : {
        "CALLBACK_URL" : "https://demo.prod.io/callback",
    },
    
    "APPLICATIONS" : {
        "App5" : {
            "KEYWORD_MAPPINGS" : {
                "CALLBACK_URL" : "https://demo.prod.io/callback2"
            }
        }
    }
}
``` 
Here, during import to the prod environment, the callback URL of the four applications are replaced with ```https://demo.prod.io/callback``` and the callback URL of the fifth application is replaced with ```https://demo.prod.io/callback2```.
