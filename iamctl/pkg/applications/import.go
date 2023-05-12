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

	log.Println("Importing applications...")
	importFilePath := filepath.Join(inputDirPath, utils.APPLICATIONS)

	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No applications to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing applications: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			RemoveDeletedDeployedResources(files, getAppList())
		}
	}

	for _, file := range files {
		appFilePath := filepath.Join(importFilePath, file.Name())
		appName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		appExists, isValidFile := validateFile(appFilePath, appName)

		if isValidFile && !utils.IsResourceExcluded(appName, utils.TOOL_CONFIGS.ApplicationConfigs) {
			importApp(appFilePath, appExists)
		}
	}
}

func validateFile(appFilePath string, appName string) (appExists bool, isValid bool) {

	appExists = false

	fileContent, err := ioutil.ReadFile(appFilePath)
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

func importApp(importFilePath string, isUpdate bool) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for application: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	appKeywordMapping := getAppKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), appKeywordMapping)

	sendImportRequest(isUpdate, importFilePath, modifiedFileData)
	return nil
}

func sendImportRequest(isUpdate bool, importFilePath string, fileData string) error {

	reqUrl := utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications/import"
	fileInfo := utils.GetFileInfo(importFilePath)

	var requestMethod string
	if isUpdate {
		log.Println("Updating app: " + fileInfo.ResourceName)
		requestMethod = "PUT"
	} else {
		log.Println("Creating app: " + fileInfo.ResourceName)
		requestMethod = "POST"
	}

	var buf bytes.Buffer
	var err error
	_, err = io.WriteString(&buf, fileData)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	mime.AddExtensionType(".yml", "application/yaml")
	mime.AddExtensionType(".xml", "application/xml")
	mime.AddExtensionType(".json", "application/json")

	mimeType := mime.TypeByExtension(fileInfo.FileExtension)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", fileInfo.FileName)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	_, err = io.Copy(part, &buf)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	request, err := http.NewRequest(requestMethod, reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
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
		return fmt.Errorf("error when sending the import request: %s", err)
	}

	statusCode := resp.StatusCode
	if statusCode == 201 {
		log.Println("Application created successfully.")
		return nil
	} else if statusCode == 200 {
		log.Println("Application updated successfully.")
		return nil
	} else if statusCode == 409 {
		log.Println("An application with the same name already exists. Please rename the file accordingly.")
		return importApp(importFilePath, true)
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing application: %s", resp.Status)
}

func RemoveDeletedDeployedResources(localFiles []os.FileInfo, deployedResources []utils.Application) {

	// Remove deployed resources that do not exist locally.
deployedResources:
	for _, resource := range deployedResources {
		for _, file := range localFiles {
			if resource.Name == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedResources
			}
		}
		log.Println("Application not found locally. Deleting app: ", resource.Name)
		utils.DeleteResource(resource.Id, "applications")
	}
}
