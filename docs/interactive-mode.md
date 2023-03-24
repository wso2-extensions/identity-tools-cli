# Interactive Mode
The interactive mode can be used to handle application configurations in an interactive manner. This can be used to add, list, export and import applications in the target environment.
> Note: This mode does not provide support for bulk resource export or import.

### Running the tool in the interactive mode
#### Tool Initialization
1. Setup the tool and the IS, following the steps in the [How to run the tool ](../README.md##How to run the tool ) section.
2. Run the following command to initialize the tool by providing the Identity server details and client ID/secret of the app you created.
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
Run the following command to provide admin user credentials.
```
iamctl serverConfiguration [flags]
```

Flags:
```
  -h, --help              help for serverConfiguration
  -p, --password string   enter your password
  -s, --server string     set server domain
  -u, --username string   enter your username
```
example:-
```
iamctl serverConfiguration -h                                           //help for serverConfiguration
iamctl serverConfiguration -s=https://localhost:9443 -u=admin -p=*****  //to complete the authorization
```
Set admin user credentials by entering inputs in an interactive way.
```
iamctl serverConfiguration
```
example:-
```
~$ iamctl serverConfiguration 
? Enter IAM URL [<schema>://<host>]: https://localhost:9443
? Enter Username: admin
? Enter Password: *****
```

### Application Related Commands
**Add application**
```
iamctl application [commands]
iamctl application     add      [flags]
```

Flags:
 ```
   -c, --callbackURl string    callbackURL  of SP - **for oauth application
   -d, --description string    description of SP - **for basic application
   -h, --help                  help for add
   -n, --name string           name of service provider - **compulsory
   -p, --password string       Password for Identity Server
   -s, --serverDomain string   server Domain
   -t, --type string           Enter application type (default "oauth")
   -u, --userName string       Username for Identity Server
 ```
Users have freedom to set flags and values according to their choices.

This ```-t, --type string           Enter application type (default "oauth")``` flag  is not mandatory. If user wants to create basic application, then should declare ```-t=basic```. Otherwise will create the oauth application as default type.

example:-
```
//create an oauth application
iamctl application add  -n=TestApplication 
iamctl application add -t=oauth -n=TestApplication
iamctl application add -t=oauth -n=TestApplication -d=description
iamctl application add -t=oauth -n=TestApplication -c=https://localhost:8010/oauth
iamctl application add -t=oauth -n=TestApplication -c=https://localhost:8010/oauth -d=description

//create an basic application
iamctl application add -t=basic -n=TestApplication
iamctl application add -t=basic -n=TestApplication -d=description
```
You can set server domain and create application at the same time.

example:-
```
//create an oauth application
iamctl application add -s=https://localhost:9443 -u=admin -p=***** -n=TestApplication 
iamctl application add -s=https://localhost:9443 -u=admin -p=***** -t=oauth -n=TestApplication
iamctl application add -s=https://localhost:9443 -u=admin -p=***** -t=oauth -n=TestApplication -d=description
iamctl application add -s=https://localhost:9443 -u=admin -p=***** -t=oauth -n=TestApplication -c=https://localhost:8010/oauth
iamctl application add -s=https://localhost:9443 -u=admin -p=***** -t=oauth -n=TestApplication -c=https://localhost:8010/oauth -d=description

//create an basic application
iamctl application add -s=https://localhost:9443 -u=admin -p=***** -t=basic -n=TestApplication
iamctl application add -s=https://localhost:9443 -u=admin -p=***** -t=basic -n=TestApplication -d=description
```

**Get list of applications**
```
iamctl application     list     [flags]
```
Flags:
```
  -h, --help              help for list
  -p, --password string   Password for Identity Server
  -s, --server string     server
  -u, --userName string   User name for Identity Server
```
example:-
```
//get list of applications
iamctl application list 
```
You cat set server domain and get the list of applications at the same time.
example:-
```
//get list of applications
iamctl application list -s=https://localhost:9443 -u=admin -p=*****
```

#### Create service providers by entering inputs in an interactive way.
**Add application and get list of applications**

```
iamctl application
```
It gives following output after entering the server domain.
```
$ iamctl application                                                      
? Select the option to move on:  [Use arrows to move, type to filter]
> Add application
  Get List
  Exit
```
To add application you should select ```add application``` from selections.
example:-
```
~$ iamctl application                                                       
? Select the option to move on: Add application
? Select the configuration type:  [Use arrows to move, type to filter]
> Basic application
  oauth
```
To view list of applications you should select ```Get List``` from selections.

example:-
```
$ iamctl application                                                        
? Select the option to move on: Get List
```
**Create a client application by getting framework specific artifacts**
```
iamctl createclientapp
```
It gives the following output
example:-
```
~$ iamctl createclientapp                                                       
? Enter your web app technology (Eg: spring-boot) : spring-boot
? Enter the package name of the project (Eg: com.example.demo) : com.example.demo
? Enter your OAuth application Name (Eg:TestApp) : Testapp
```

Flags:
```
  -h, --help                  Help for list
  -t, --technology   string   Technology or Framework of the web app. Eg: spring-boot
  -k, --package      string   Package name of the project where artifacts are going to be placed. Eg: com.example.demo
  -a, --application  string   Name of the application Eg:TestApp
```
Now you have successfully installed the artifacts and secured with OIDC using WSO2 IS..