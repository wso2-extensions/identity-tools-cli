/*
Copyright (c) 2020, WSO2 Inc. (http://www.wso2.com). All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var technology string
var application string
var packageName string

var ArtifactServiceUrl string

var clientapp = []*survey.Question{
	{
		Name:     "technology",
		Prompt:   &survey.Input{Message: "Enter your web app technology (Eg: spring-boot) :"},
		Validate: survey.Required,
	},
	{
		Name:     "package",
		Prompt:   &survey.Input{Message: "Enter the package name of the project (Eg: com.example.demo) :"},
		Validate: survey.Required,
	},
	{
		Name:     "application",
		Prompt:   &survey.Input{Message: "Enter your OAuth application Name (Eg:TestApp) :"},
		Validate: survey.Required,
	},
}

//Response from artifact-service
type HttpResponse struct {
	Error        bool   `json:"error"`
	ErrorMessage string `json:"errorMessage"`
	ResponseData ResponseData
}
type ResponseData struct {
	ArtifactInfo ArtifactInfo
	Data         []Data
}

type ArtifactInfo struct {
	Name string `json:"name"`
}

type Data struct {
	Data     string   `json:"data"`
	Metadata Metadata `json:"metadata"`
}
type Metadata struct {
	Path      string `json:"path"`
	Operation string `json:"operation"`
}

func setProjectPath() {

	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render(appName)
	fmt.Print(renderStr)

	ClientAppProject := struct {
		ProjectSrcPath string `survey:"technology"`
		PackageName    string `survey:"package"`
		Application    string `survey:"application"`
	}{}

	err := survey.Ask(clientapp, &ClientAppProject)
	if err != nil {
		fmt.Println(err)
		return
	}
	technology = ClientAppProject.ProjectSrcPath
	application = ClientAppProject.Application
	packageName = ClientAppProject.PackageName
}

func init() {

	rootCmd.AddCommand(createClientApp)
	createClientApp.Flags().StringP("technology", "t", "", "enter web app technology")
	createClientApp.Flags().StringP("package", "k", "", "enter your project package")
	createClientApp.Flags().StringP("application", "a", "", "enter name of the OAuth application")
}

var createClientApp = &cobra.Command{

	Use:   "createclientapp",
	Short: "Secure your webapp with IS",
	Long:  `This will help you to integrate your webapp with IS`,
	Run: func(cmd *cobra.Command, args []string) {

		technology, _ = cmd.Flags().GetString("technology")
		packageName, _ = cmd.Flags().GetString("package")
		application, _ = cmd.Flags().GetString("application")

		if technology == "" {
			setProjectPath()
		}
		fmt.Println("Generating client side artifacts for: ")
		fmt.Println("Technology: " + technology)
		fmt.Println("Package: " + packageName)
		fmt.Println("Application Name: " + application + "\n")

		installArtifacts()
	},
}

func installArtifacts() {

	resp := callArtifactServiceApi(technology, application)
	respBody, err := ioutil.ReadAll(resp.Body)
	status := resp.StatusCode
	if err != nil {
		fmt.Println("Error while reading the response body from artifact service")
		return
	}
	if status == 401 {
		fmt.Println("Resource Server " + ArtifactServiceUrl + "returned 401-Unauthorized access.")
		fmt.Println("Please enter your UserName and password for server.")
		SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()
		setServerWithInit(SERVER)
		installArtifacts()
	} else if status == 400 {
		fmt.Println("Server " + ArtifactServiceUrl + "returned 400 response. Please retry.. ")
		fmt.Println("Provided parameters are not in correct format..")
	} else if status == 403 {
		fmt.Println("Forbidden")
		fmt.Println("Server " + ArtifactServiceUrl + "returned 403 response. Please retry.. ")
	} else if status == 201 || status == 200 {
		fmt.Println("Obtained the response from the artifact-service..")
		if technology == "spring-boot" {
			installSpringBootArtifacts(respBody)
		}
	}
	defer resp.Body.Close()
}
