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

package userstores

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing user stores...")
	importFilePath := filepath.Join(inputDirPath, utils.USERSTORES.String())

	if utils.IsResourceTypeExcluded(utils.USERSTORES) {
		return
	}
	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No user stores to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing user stores: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			removeDeletedDeployedUserstores(files)
		}
	}

	exportAPIexists := utils.ExportAPIExists(utils.USERSTORES)
	for _, file := range files {
		userStoreFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(userStoreFilePath)
		userStoreName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(userStoreName, utils.TOOL_CONFIGS.UserStoreConfigs) {
			userStoreId, err := getUserStoreId(userStoreName)
			if err != nil {
				log.Printf("Invalid file configurations for user store: %s. %s", userStoreName, err)
			} else {
				err := importUserStore(userStoreId, userStoreName, userStoreFilePath, exportAPIexists)
				if err != nil {
					log.Println("Error importing user store: ", err)
				}
			}
		}
	}
}

func importUserStore(userStoreId, userStoreName, userStoreFilePath string, exportAPIexists bool) error {

	fileBytes, err := ioutil.ReadFile(userStoreFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for user store: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	userStoreKeywordMapping := getUserStoreKeywordMapping(userStoreName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), userStoreKeywordMapping)

	if exportAPIexists {
		if userStoreId == "" {
			return importUserStoreOperation(userStoreName, userStoreFilePath, modifiedFileData)
		}
		return updateUserStoreOperation(userStoreId, userStoreName, userStoreFilePath, modifiedFileData)
	}

	format, err := utils.FormatFromExtension(filepath.Ext(userStoreFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for user store: %w", err)
	}

	if userStoreId == "" {
		return importUserStoreWithCRUD(userStoreName, []byte(modifiedFileData), format)
	}
	return updateUserStoreWithCRUD(userStoreId, userStoreName, []byte(modifiedFileData), format)
}

func importUserStoreOperation(userStoreName, userStoreFilePath, modifiedFileData string) error {

	log.Println("Creating new user store: " + userStoreName)
	err := utils.SendImportRequest(userStoreFilePath, modifiedFileData, utils.USERSTORES)
	if err != nil {
		utils.UpdateFailureSummary(utils.USERSTORES, userStoreName)
		return fmt.Errorf("error when importing user store: %s", err)
	}
	utils.UpdateSuccessSummary(utils.USERSTORES, utils.IMPORT)
	log.Println("User store imported successfully.")
	return nil
}

func updateUserStoreOperation(userStoreId, userStoreName, userStoreFilePath, modifiedFileData string) error {

	log.Println("Updating user store: " + userStoreName)
	err := utils.SendUpdateRequest(userStoreId, userStoreFilePath, modifiedFileData, utils.USERSTORES)
	if err != nil {
		utils.UpdateFailureSummary(utils.USERSTORES, userStoreName)
		return fmt.Errorf("error when updating user store: %s", err)
	}
	utils.UpdateSuccessSummary(utils.USERSTORES, utils.UPDATE)
	log.Println("User store updated successfully.")
	return nil
}

func importUserStoreWithCRUD(userStoreName string, requestBody []byte, format utils.Format) error {

	log.Println("Creating new user store: " + userStoreName)

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.USERSTORES, "typeName", "className")
	if err != nil {
		return err
	}
	resp, err := utils.SendPostRequest(utils.USERSTORES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when importing user store: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.USERSTORES, utils.IMPORT)
	log.Println("User store imported successfully.")
	return nil
}

func updateUserStoreWithCRUD(userStoreId, userStoreName string, requestBody []byte, format utils.Format) error {

	log.Println("Updating user store: " + userStoreName)

	updateBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.USERSTORES, "typeName", "className")
	if err != nil {
		return err
	}
	resp, err := utils.SendPutRequest(utils.USERSTORES, userStoreId, updateBody)
	if err != nil {
		return fmt.Errorf("error when updating user store: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.USERSTORES, utils.UPDATE)
	log.Println("User store updated successfully.")
	return nil
}

func removeDeletedDeployedUserstores(localFiles []os.FileInfo) {

	// Remove deployed user stores that do not exist locally.
	deployedUserstores, err := getUserStoreList()
	if err != nil {
		log.Println("Error retrieving deployed user stores: ", err)
		return
	}
deployedResourcess:
	for _, userstore := range deployedUserstores {
		for _, file := range localFiles {
			if userstore.Name == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedResourcess
			}
		}
		if utils.IsResourceExcluded(userstore.Name, utils.TOOL_CONFIGS.ApplicationConfigs) {
			log.Printf("Userstore: %s is excluded from deletion.\n", userstore.Name)
			continue
		}
		log.Println("User store not found locally. Deleting user store: ", userstore.Name)
		err := utils.SendDeleteRequest(userstore.Id, utils.USERSTORES)
		if err != nil {
			utils.UpdateFailureSummary(utils.USERSTORES, userstore.Name)
			log.Println("Error deleting user store: ", err)
		}
		utils.UpdateSuccessSummary(utils.USERSTORES, utils.DELETE)
	}
}
