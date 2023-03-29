/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

func ImportAll(inputDirPath string) {

	importFilePath := inputDirPath + "/Applications/"

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		appFilePath := importFilePath + file.Name()
		appName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		appExists, isValidFile := validateFile(appFilePath, appName)

		if isValidFile && !utils.IsResourceExcluded(appName, utils.TOOL_CONFIGS.ApplicationConfigs) {
			importApp(appFilePath, appExists)
		}
	}
}

func validateFile(appFilePath string, appName string) (appExists bool, isValid bool) {

	appExists = false

	fileContent, err := os.ReadFile(appFilePath)
	if err != nil {
		log.Println("Error when reading the file for app: ", appName, err)
		return appExists, false
	}

	// Validate the YAML format.
	var appConfig AppConfig
	err = yaml.Unmarshal(fileContent, &appConfig)
	if err != nil {
		log.Println("Invalid file content for app: ", appName, err)
		return appExists, false
	}

	existingAppList := getDeployedAppNames()
	for _, app := range existingAppList {
		if app == appConfig.ApplicationName {
			appExists = true
			break
		}
	}
	if appConfig.ApplicationName != appName {
		log.Println("Warning: Application name in the file " + appFilePath + " is not matching with the file name.")
	}
	return appExists, true
}

func importApp(importFilePath string, isUpdate bool) {

	fileBytes, err := os.ReadFile(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	appName, _, _ := getAppFileInfo(importFilePath)
	appKeywordMapping := getAppKeywordMapping(appName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), appKeywordMapping)

	sendImportRequest(isUpdate, importFilePath, modifiedFileData)

}

func sendImportRequest(isUpdate bool, importFilePath string, fileData string) {

	reqUrl := utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications/import"
	appName, filename, fileExtension := getAppFileInfo(importFilePath)

	var requestMethod string
	if isUpdate {
		log.Println("Updating app: " + appName)
		requestMethod = "PUT"
	} else {
		log.Println("Creating app: " + appName)
		requestMethod = "POST"
	}

	var buf bytes.Buffer
	var err error
	_, err = io.WriteString(&buf, fileData)
	if err != nil {
		log.Fatal(err)
	}

	mime.AddExtensionType(".yml", "text/yaml")
	mime.AddExtensionType(".xml", "application/xml")

	mimeType := mime.TypeByExtension(fileExtension)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(part, &buf)
	if err != nil {
		log.Fatal(err)
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
	if statusCode == 201 {
		log.Println("Application created successfully.")
	} else if statusCode == 200 {
		log.Println("Application updated successfully.")
	} else if statusCode == 409 {
		log.Println("An application with the same name already exists. Please rename the file accordingly.")
		importApp(importFilePath, true)
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		log.Println(error)
	} else {
		log.Println("Error while updating the application.")
	}
}
