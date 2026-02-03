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

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing applications...")
	importFilePath := filepath.Join(inputDirPath, configs.APPLICATIONS)
	if !utils.IsEntitySupportedInVersion(configs.APPLICATIONS) {
		return
	}

	if utils.IsResourceTypeExcluded(configs.APPLICATIONS) {
		return
	}
	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No applications to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing applications: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			removeDeletedDeployedApps(files, importFilePath)
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
	fileDataWithReplacedKeywords := utils.ReplaceKeywords(string(fileBytes), appKeywordMapping)
	modifiedFileData := utils.RemoveSecretMasks(fileDataWithReplacedKeywords)

	if isUpdate {
		return updateApplication(importFilePath, modifiedFileData, fileInfo)
	}
	return importApplication(importFilePath, modifiedFileData, fileInfo)
}

func updateApplication(importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Updating application: " + fileInfo.ResourceName)
	err := utils.SendUpdateRequest("", importFilePath, modifiedFileData, configs.APPLICATIONS)
	if err != nil {
		utils.UpdateFailureSummary(configs.APPLICATIONS, fileInfo.ResourceName)
		return fmt.Errorf("error when updating application: %s", err)
	}
	utils.UpdateSuccessSummary(configs.APPLICATIONS, utils.UPDATE)
	log.Println("Application updated successfully.")
	return nil
}

func importApplication(importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Creating new application: " + fileInfo.ResourceName)
	err := utils.SendImportRequest(importFilePath, modifiedFileData, configs.APPLICATIONS)
	if err != nil {
		utils.UpdateFailureSummary(configs.APPLICATIONS, fileInfo.ResourceName)
		return fmt.Errorf("error when importing application: %s", err)
	}

	if oauthApp, err := isOauthApp(modifiedFileData); err != nil {
		fmt.Println("Failed to check if the applications is an OAuth app:", err.Error())
	} else if oauthSecretGiven, err := isOauthSecretGiven(modifiedFileData); err != nil {
		fmt.Println("Failed to check if oauthConsumerSecret is given:", err.Error())
	} else if oauthApp && !oauthSecretGiven {
		// Check if oauthConsumerSecret is given or else add an indicator to the summary informing a new secret is generated.
		utils.AddNewSecretIndicatorToSummary(fileInfo.ResourceName)
	}
	utils.UpdateSuccessSummary(configs.APPLICATIONS, utils.IMPORT)
	log.Println("Application imported successfully.")
	return nil
}

func removeDeletedDeployedApps(localFiles []os.FileInfo, importFilePath string) {

	// Remove deployed applications that do not exist locally.
	deployedApps := getAppList()
deployedResources:
	for _, app := range deployedApps {
		for _, file := range localFiles {
			isToolManagementApp, err := isToolMgtApp(file, importFilePath)
			if err != nil {
				log.Printf("Error checking if application is a tool management app: %s\n", err.Error())
				log.Printf("Application: %s is excluded from deletion.\n", app.Name)
				continue deployedResources
			}
			if app.Name == utils.GetFileInfo(file.Name()).ResourceName || isToolManagementApp {
				continue deployedResources
			}
		}
		if utils.IsResourceExcluded(app.Name, utils.TOOL_CONFIGS.ApplicationConfigs) || app.Name == utils.CONSOLE || app.Name == utils.MY_ACCOUNT {
			log.Printf("Application: %s is excluded from deletion.\n", app.Name)
			continue
		}
		log.Println("Application not found locally. Deleting app: ", app.Name)
		err := utils.SendDeleteRequest(app.Id, configs.APPLICATIONS)
		if err != nil {
			utils.UpdateFailureSummary(configs.APPLICATIONS, app.Name)
			log.Println("Error deleting application: ", app.Name, err)
		}
		utils.UpdateSuccessSummary(configs.APPLICATIONS, utils.DELETE)
	}
}
