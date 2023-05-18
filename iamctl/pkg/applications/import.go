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
	"fmt"
	"io/ioutil"
	"log"
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

	if isUpdate {
		log.Println("Updating application: " + fileInfo.ResourceName)
		err = utils.SendUpdateRequest("", importFilePath, modifiedFileData, utils.APPLICATIONS)
	} else {
		log.Println("Creating new application: " + fileInfo.ResourceName)
		err = utils.SendImportRequest(importFilePath, modifiedFileData, utils.APPLICATIONS)
	}
	if err != nil {
		return fmt.Errorf("error when importing application: %s", err)
	}
	log.Println("Application imported successfully.")
	return nil
}
