# IAM-CTL

This will use to get support for create service providers inside the Identity Server.
Here you can create service providers by entering inputs as a command  and flags or entering inputs in an interactive way.


### Pre-Requisites
* WSO2IS 5.10.0 alpha2 and start the server.

### How to run the executable file 

1. First clone or download  ```identity-cli``` module.
2. Copy the path of executable file according to your platform.
* linux:

```<identity-cli_PATH>/command-line-extension/bin/linux/iamctl```

* mac:

```<identity-cli_PATH>/command-line-extension/bin/mac/iamctl```

* windows

```<identity-cli_PATH>/command-line-extension/bin/windows/iamctl.exe```


3. Then you select your directory to work and open a terminal.
4. You can choose any name as keyword of IAM-CTL and set that using following command according to your platform.
Here I have chosen 'iamctl' as keyword.

* linux:
 
    ```
    alias iamctl="<identity-cli_PATH>/command-line-extension/bin/linux/iamctl" 
    ```
  
* mac:
 
  ```
   alias iamctl="<identity-cli_PATH>/command-line-extension/bin/mac/iamctl" 
  ```

* windows

    ```
    doskey iamctl=<identity-cli_PATH>/command-line-extension/bin/windows/iamctl.exe $*
    ```
 
 
5. Now you can run IAM-ctl using your keyword.
```
iamctl -h
```
It gives following details.
```
:~$ iamctl -h
Service Provider configuration

Usage:
  IAM-CTL [flags]
  IAM-CTL [command]

Available Commands:
  application         Create a service provider
  help                Help about any command
  init                you can set your sample SP
  serverConfiguration you can set your server domain

Flags:
  -h, --help   help for IAM-CTL

Use "IAM-CTL [command] --help" for more information about a command.
```
6. First you need to configure service provider to get access from identity server. For that this [link](https://docs.wso2.com/display/IS570/Configuring+OAuth2-OpenID+Connect+Single-Sign-On) will help to you..
7. Then use  client_key, client_secret of created service provider to do the authorization relevant to server domain. It should be completed as follows.

```
iamctl init
```
Now you should gives answers for questions asked by CTL.
```
:~$ iamctl init
  ___      _      __  __            ____   _____   _     
 |_ _|    / \    |  \/  |          / ___| |_   _| | |    
  | |    / _ \   | |\/| |  _____  | |       | |   | |    
  | |   / ___ \  | |  | | |_____| | |___    | |   | |___ 
 |___| /_/   \_\ |_|  |_|          \____|   |_|   |_____|
      
? Enter IAM URL [<schema>://<host>]: https://localhost:9443                                                   
? Enter clientID: M4fucPehFTuFkHHLNGfxIEf6ydka
? Enter clientSecret: DWXnw7UgvsRUKWXrftvnU_vclzAa
? Enter Tenant domain: carbon.super
```
8. Then you need to configure server domain as follows.
### Set server domain by entering inputs as a command and flags.
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
### Set server domain by entering inputs in an interactive way.
```
iamctl serverConfiguration
```
example:-
```
~$ iamctl serverConfiguration
  ___      _      __  __            ____   _____   _     
 |_ _|    / \    |  \/  |          / ___| |_   _| | |    
  | |    / _ \   | |\/| |  _____  | |       | |   | |    
  | |   / ___ \  | |  | | |_____| | |___    | |   | |___ 
 |___| /_/   \_\ |_|  |_|          \____|   |_|   |_____|
                                                         
? Enter IAM URL [<schema>://<host>]: https://localhost:9443
? Enter Username: admin
? Enter Password: *****
```
9.Now authorization part is completed. You can add application and get list of applications using IAM-CTL. Structure as follows.

### create service providers by entering inputs as a command and flags.
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
You cat set server domain and create application at the same time.

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

### create service providers by entering inputs in an interactive way.
**Add application and get list of applications**

```
iamctl application
```
It gives following output after entering the server domain.
```
$ iamctl application
  ___      _      __  __            ____   _____   _     
 |_ _|    / \    |  \/  |          / ___| |_   _| | |    
  | |    / _ \   | |\/| |  _____  | |       | |   | |    
  | |   / ___ \  | |  | | |_____| | |___    | |   | |___ 
 |___| /_/   \_\ |_|  |_|          \____|   |_|   |_____|
                                                         
? Select the option to move on:  [Use arrows to move, type to filter]
> Add application
  Get List
  Exit
```
To add application you should select ```add application``` from selections.
example:-
```
~$ iamctl application
  ___      _      __  __            ____   _____   _     
 |_ _|    / \    |  \/  |          / ___| |_   _| | |    
  | |    / _ \   | |\/| |  _____  | |       | |   | |    
  | |   / ___ \  | |  | | |_____| | |___    | |   | |___ 
 |___| /_/   \_\ |_|  |_|          \____|   |_|   |_____|
                                                         
? Select the option to move on: Add application
? Select the configuration type:  [Use arrows to move, type to filter]
> Basic application
  oauth
```
To view list of applications you should select ```Get List``` from selections.
 
example:-
```
$ iamctl application

  ___      _      __  __            ____   _____   _     
 |_ _|    / \    |  \/  |          / ___| |_   _| | |    
  | |    / _ \   | |\/| |  _____  | |       | |   | |    
  | |   / ___ \  | |  | | |_____| | |___    | |   | |___ 
 |___| /_/   \_\ |_|  |_|          \____|   |_|   |_____|
                                                         
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
  ___      _      __  __            ____   _____   _     
 |_ _|    / \    |  \/  |          / ___| |_   _| | |    
  | |    / _ \   | |\/| |  _____  | |       | |   | |    
  | |   / ___ \  | |  | | |_____| | |___    | |   | |___ 
 |___| /_/   \_\ |_|  |_|          \____|   |_|   |_____|
                                                         
? Enter your web app technology (Eg: spring-boot) : spring-boot
? Enter the package name of the project (Eg: com.example.demo) : com.example.demo
? Enter your OAuth application Name (Eg:TestApp) : Testapp
```
Now you have successfully installed the artifacts and secured with OIDC using WSO2 IS..