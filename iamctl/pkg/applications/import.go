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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing applications...")
	importFilePath := filepath.Join(inputDirPath, utils.APPLICATIONS.String())
	exportAPIExists := utils.ExportAPIExists(utils.APPLICATIONS)

	if utils.IsResourceTypeExcluded(utils.APPLICATIONS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No applications to import.")
		return
	}

	deployedApps := getAppList()
	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing applications: ", err)
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedApps(files, deployedApps)
	}

	for _, file := range files {
		appFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(appFilePath)
		appName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(appName, utils.TOOL_CONFIGS.ApplicationConfigs) {
			appId := getAppId(appName, deployedApps)
			err := importApp(appId, appName, appFilePath, exportAPIExists)
			if err != nil {
				log.Println("Error importing application: ", err)
				utils.UpdateFailureSummary(utils.APPLICATIONS, appName)
			}
		}
	}
}

func importApp(appId, appName, importFilePath string, exportAPIExists bool) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for application: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	appKeywordMapping := getAppKeywordMapping(appName)
	fileDataWithReplacedKeywords := utils.ReplaceKeywords(string(fileBytes), appKeywordMapping)
	modifiedFileData := utils.RemoveSecretMasks(fileDataWithReplacedKeywords)

	if exportAPIExists {
		if appId == "" {
			return importApplication(appName, importFilePath, modifiedFileData)
		}
		return updateApplication(appName, importFilePath, modifiedFileData)
	}

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for application: %w", err)
	}

	appMap, err := utils.DeserializeToMap([]byte(modifiedFileData), format, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error deserializing application: %w", err)
	}
	delete(appMap, "id")

	if appId == "" {
		return importAppWithCRUD(appName, appMap)
	}
	return updateAppWithCRUD(appId, appName, appMap)
}

func importApplication(appName, importFilePath, modifiedFileData string) error {

	log.Println("Creating new application: " + appName)
	err := utils.SendImportRequest(importFilePath, modifiedFileData, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error when importing application: %s", err)
	}

	if oauthApp, err := isOauthApp(modifiedFileData); err != nil {
		fmt.Println("Failed to check if the applications is an OAuth app:", err.Error())
	} else if oauthSecretGiven, err := isOauthSecretGiven(modifiedFileData); err != nil {
		fmt.Println("Failed to check if oauthConsumerSecret is given:", err.Error())
	} else if oauthApp && !oauthSecretGiven {
		// Check if oauthConsumerSecret is given or else add an indicator to the summary informing a new secret is generated.
		utils.AddNewSecretIndicatorToSummary(appName)
	}
	utils.UpdateSuccessSummary(utils.APPLICATIONS, utils.IMPORT)
	log.Println("Application imported successfully.")
	return nil
}

func updateApplication(appName, importFilePath, modifiedFileData string) error {

	log.Println("Updating application: " + appName)
	err := utils.SendUpdateRequest("", importFilePath, modifiedFileData, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error when updating application: %s", err)
	}
	utils.UpdateSuccessSummary(utils.APPLICATIONS, utils.UPDATE)
	log.Println("Application updated successfully.")
	return nil
}

func importAppWithCRUD(appName string, appMap map[string]interface{}) error {

	log.Println("Creating new application: " + appName)

	newSecretCreated, err := processInboundProtocolsForPost(appMap)
	if err != nil {
		return fmt.Errorf("error processing inbound protocols: %w", err)
	}

	body, err := json.Marshal(appMap)
	if err != nil {
		return fmt.Errorf("error marshalling application: %w", err)
	}

	resp, err := utils.SendPostRequest(utils.APPLICATIONS, body)
	if err != nil {
		return fmt.Errorf("error creating application: %w", err)
	}
	defer resp.Body.Close()

	if newSecretCreated {
		utils.AddNewSecretIndicatorToSummary(appName)
	}
	utils.UpdateSuccessSummary(utils.APPLICATIONS, utils.IMPORT)
	log.Println("Application imported successfully.")
	return nil
}

func updateAppWithCRUD(appId, appName string, appMap map[string]interface{}) error {

	// Implemented in commit 4.
	return fmt.Errorf("updateAppWithCRUD not yet implemented")
}

func removeDeletedDeployedApps(localFiles []os.FileInfo, deployedApps []Application) {

	localAppNames := make(map[string]struct{})
	for _, file := range localFiles {
		localAppNames[utils.GetFileInfo(file.Name()).ResourceName] = struct{}{}
	}

	for _, app := range deployedApps {
		if _, existsLocally := localAppNames[app.Name]; existsLocally {
			continue
		}

		if utils.IsResourceExcluded(app.Name, utils.TOOL_CONFIGS.ApplicationConfigs) ||
			app.Name == utils.CONSOLE || app.Name == utils.MY_ACCOUNT || app.Name == utils.CARBON_SP {
			log.Printf("Application: %s is excluded from deletion.\n", app.Name)
			continue
		}
		if isToolMgt, err := isToolMgtApp(app.Id); err != nil {
			log.Printf("Error checking if application is the tool management app: %s\n", err.Error())
			log.Printf("Application: %s is excluded from deletion.\n", app.Name)
			continue
		} else if isToolMgt {
			log.Printf("Info: Tool Management App: %s is excluded from deletion.\n", app.Name)
			continue
		}

		log.Println("Application not found locally. Deleting app: ", app.Name)
		err := utils.SendDeleteRequest(app.Id, utils.APPLICATIONS)
		if err != nil {
			utils.UpdateFailureSummary(utils.APPLICATIONS, app.Name)
			log.Println("Error deleting application: ", app.Name, err)
		}
		utils.UpdateSuccessSummary(utils.APPLICATIONS, utils.DELETE)
	}
}
