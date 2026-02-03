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
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing user stores...")
	importFilePath := filepath.Join(inputDirPath, configs.USERSTORES)
	if !utils.IsEntitySupportedInVersion(configs.USERSTORES) {
		return
	}

	if utils.IsResourceTypeExcluded(configs.USERSTORES) {
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

	for _, file := range files {
		userStoreFilePath := filepath.Join(importFilePath, file.Name())
		userStoreName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(userStoreName, utils.TOOL_CONFIGS.UserStoreConfigs) {
			userStoreId, err := getUserStoreId(userStoreFilePath)
			if err != nil {
				log.Printf("Invalid file configurations for user store: %s. %s", userStoreName, err)
			} else {
				err := importUserStore(userStoreId, userStoreFilePath)
				if err != nil {
					log.Println("Error importing user store: ", err)
				}
			}
		}
	}
}

func importUserStore(userStoreId string, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for user store: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	userStoreKeywordMapping := getUserStoreKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), userStoreKeywordMapping)

	if userStoreId == "" {
		return importUserStoreOperation(importFilePath, modifiedFileData, fileInfo)
	}
	return updateUserStoreOperation(userStoreId, importFilePath, modifiedFileData, fileInfo)
}

func importUserStoreOperation(importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Creating new user store: " + fileInfo.ResourceName)
	err := utils.SendImportRequest(importFilePath, modifiedFileData, configs.USERSTORES)
	if err != nil {
		utils.UpdateFailureSummary(configs.USERSTORES, fileInfo.ResourceName)
		return fmt.Errorf("error when importing user store: %s", err)
	}
	utils.UpdateSuccessSummary(configs.USERSTORES, utils.IMPORT)
	log.Println("User store imported successfully.")
	return nil
}

func updateUserStoreOperation(userStoreId string, importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Updating user store: " + fileInfo.ResourceName)
	err := utils.SendUpdateRequest(userStoreId, importFilePath, modifiedFileData, configs.USERSTORES)
	if err != nil {
		utils.UpdateFailureSummary(configs.USERSTORES, fileInfo.ResourceName)
		return fmt.Errorf("error when updating user store: %s", err)
	}
	utils.UpdateSuccessSummary(configs.USERSTORES, utils.UPDATE)
	log.Println("User store updated successfully.")
	return nil
}

func removeDeletedDeployedUserstores(localFiles []os.FileInfo) {

	// Remove deployed user stores that do not exist locally.
	deployedUserstores, err := getUserStoreList()
	if err != nil {
		log.Println("Error retrieving deployed userstores: ", err)
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
		log.Println("User store not found locally. Deleting userstore: ", userstore.Name)
		err := utils.SendDeleteRequest(userstore.Id, configs.USERSTORES)
		if err != nil {
			utils.UpdateFailureSummary(configs.USERSTORES, userstore.Name)
			log.Println("Error deleting user store: ", err)
		}
		utils.UpdateSuccessSummary(configs.USERSTORES, utils.DELETE)
	}
}
