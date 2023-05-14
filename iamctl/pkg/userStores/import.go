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

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing user stores...")
	importFilePath := filepath.Join(inputDirPath, utils.USERSTORES)

	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No user stores to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing user stores: ", err)
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

	fmt.Println(userStoreId)
	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for user store: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	userStoreKeywordMapping := getUserStoreKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), userStoreKeywordMapping)

	if userStoreId == "" {
		log.Println("Creating new user store: " + fileInfo.ResourceName)
		err = utils.SendImportRequest(importFilePath, modifiedFileData, utils.USERSTORES)
	} else {
		log.Println("Updating user store: " + fileInfo.ResourceName)
		err = utils.SendUpdateRequest(userStoreId, importFilePath, modifiedFileData, utils.USERSTORES)
	}
	if err != nil {
		return fmt.Errorf("error when importing user store: %s", err)
	}
	log.Println("User store imported successfully.")
	return nil
}
