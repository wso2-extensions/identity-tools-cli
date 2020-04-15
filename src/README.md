## IAM-CTL

### How to build the executable file 
To build the IAM-ctl in your computer, you should have the go in your computer. If you  have not go in your computer you can install go using this [link](https://golang.org/doc/install).

Now you can build the IAM-ctl in your computer.
1. Open a terminal and set directory to ```<identity-cli_PATH>/command-line-extension/src```

2. Then build the IAM-ctl.
```
go build
```
 As the result created the executable file named ```iamctl```
 
 ### If you want to build for supports to cross platform you can do as follows.
 
  Open a terminal and set directory to ```<identity-cli_PATH>/command-line-extension/src```
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

 
         
    