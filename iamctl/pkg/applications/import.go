/**
* Copyright (c) 2022, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
*
* WSO2 LLC. licenses this file to you under the Apache License,
* Version 2.0 (the "License"); you may not use this file except
* in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package applications

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	ApplicationName string `yaml:"applicationName"`
}

func ImportAll(inputDirPath string) {

	var importFilePath = "."
	if inputDirPath != "" {
		importFilePath = inputDirPath
	}
	importFilePath = importFilePath + "/Applications/"

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var appFilePath string
	appList := getAppNames()
	for _, file := range files {
		appFilePath = importFilePath + file.Name()
		fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		// Read the content of the file.
		fileContent, err := ioutil.ReadFile(appFilePath)
		if err != nil {
			log.Fatal(err)
		}

		// Parse the YAML content.
		var appConfig AppConfig
		err = yaml.Unmarshal(fileContent, &appConfig)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(appConfig.ApplicationName)

		// Check if app exists.
		var appExists bool
		for _, app := range appList {
			if app == appConfig.ApplicationName {
				appExists = true
				break
			}
		}

		if appConfig.ApplicationName != fileName {
			log.Println("Application name in the file " + appFilePath + " is not matching with the file name.")
		}

		importApp(appFilePath, appExists)
	}
}

func importApp(importFilePath string, update bool) {

	var reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications/import"
	var err error

	fmt.Println(reqUrl)
	file, err := os.Open(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	filename := filepath.Base(importFilePath)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Get file extension
	fileExtension := filepath.Ext(filename)

	mime.AddExtensionType(".yml", "text/yaml")
	mime.AddExtensionType(".xml", "application/xml")

	mimeType := mime.TypeByExtension(fileExtension)

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal(err)
	}

	defer writer.Close()

	var requestMethod string
	if update {
		log.Println("Updating app: " + filename)
		requestMethod = "PUT"
	} else {
		log.Println("Creating app: " + filename)
		requestMethod = "POST"
	}

	request, err := http.NewRequest(requestMethod, reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	statusCode := resp.StatusCode
	fmt.Println(statusCode)
	switch statusCode {
	case 401:
		log.Println("Unauthorized access.\nPlease check your Username and password.")
	case 400:
		log.Println("Provided parameters are not in correct format.")
	case 403:
		log.Println("Forbidden request.")
	case 409:
		log.Println("An application with the same name already exists. Please rename the file accordingly.")
		importApp(importFilePath, true)
	case 500:
		log.Println("Internal server error.")
	case 201:
		log.Println("Application imported successfully.")
	}
}

func getAppNames() []string {

	// Get the list of applications.
	apps := getAppList()

	// Extract application names from the list.
	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}

	return appNames
}
