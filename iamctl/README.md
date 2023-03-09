## IAM-CTL

### How to build the executable file 
To build the IAM-ctl in your computer, you should have go installed. If you do not have go in your computer, you can install go using this [link](https://golang.org/doc/install).

Now you can build the IAM-ctl.
Here onwards, the location of the identity-tools-cli repository will be referred as ```<identity-cli_HOME>```.
1. Open a terminal and set directory to ```<identity-cli_HOME>/iamctl```

2. Then build the IAM-ctl.
```
go build
```
 As a result, an executable file named ```iamctl``` will be created.
 
 ### If you want to build for supports to cross platform you can do as follows.
 
  Open a terminal and set directory to ```<identity-cli_HOME>/iamctl```
  Then build the IAMCTL.
  
  To build for mac:
  ```
GOOS=darwin GOARCH=amd64 go build
   ```
To build for windows:
```
GOOS=windows GOARCH=amd64 go build
```
To build for linux:
```
GOOS=linux GOARCH=amd64 go build
```

 
         
    