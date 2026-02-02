/**
* Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package oidcScopes

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

	log.Println("Importing OIDC scopes...")
	importFilePath := filepath.Join(inputDirPath, utils.OIDC_SCOPES.String())

	if utils.IsResourceTypeExcluded(utils.OIDC_SCOPES) {
		return
	}
	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No OIDC scopes to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing OIDC scopes: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			removeDeletedDeployedScopes(files)
		}

	}

	for _, file := range files {
		scopeFilePath := filepath.Join(importFilePath, file.Name())
		scopeName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(scopeName, utils.TOOL_CONFIGS.OidcScopeConfigs) {
			scopeExists, err := isScopeExists(scopeFilePath, scopeName)
			if err != nil {
				log.Printf("Invalid file configurations for OIDC scope: %s. %s", scopeName, err)
			} else {
				err := importOidcScope(scopeName, scopeExists, scopeFilePath)
				if err != nil {
					log.Println("Error importing OIDC scope: ", err)
				}
			}
		}
	}
}

func importOidcScope(scopeName string, scopeExists bool, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for OIDC scope: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	scopeKeywordMapping := getOidcScopeKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), scopeKeywordMapping)

	if !scopeExists {
		return importScope(importFilePath, modifiedFileData, fileInfo)
	}
	return updateScope(scopeName, importFilePath, modifiedFileData, fileInfo)
}

func importScope(importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Creating new OIDC scope: " + fileInfo.ResourceName)
	err := utils.SendImportRequest(importFilePath, modifiedFileData, utils.OIDC_SCOPES)
	if err != nil {
		utils.UpdateFailureSummary(utils.OIDC_SCOPES, fileInfo.ResourceName)
		return fmt.Errorf("error when importing OIDC scope: %s", err)
	}
	utils.UpdateSuccessSummary(utils.OIDC_SCOPES, utils.IMPORT)
	log.Println("OIDC scope imported successfully.")
	return nil
}

func updateScope(scopeId string, importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Updating OIDC scope: " + fileInfo.ResourceName)
	err := utils.SendUpdateRequest(scopeId, importFilePath, modifiedFileData, utils.OIDC_SCOPES)
	if err != nil {
		utils.UpdateFailureSummary(utils.OIDC_SCOPES, fileInfo.ResourceName)
		return fmt.Errorf("error when updating OIDC scope: %s", err)
	}
	utils.UpdateSuccessSummary(utils.OIDC_SCOPES, utils.UPDATE)
	log.Println("OIDC scope updated successfully.")
	return nil
}

func removeDeletedDeployedScopes(localFiles []os.FileInfo) {

	// Remove deployed OIDC scopes that do not exist locally.
	deployedScopes, err := getOidcScopeList()
	if err != nil {
		log.Println("Error retrieving deployed OIDC scopes: ", err)
		return
	}
deployedScopes:
	for _, scope := range deployedScopes {
		for _, file := range localFiles {
			if scope.Name == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedScopes
			}
		}
		if utils.IsResourceExcluded(scope.Name, utils.TOOL_CONFIGS.OidcScopeConfigs) {
			log.Println("OIDC scope is excluded from deletion: ", scope.Name)
			continue
		}
		log.Printf("OIDC scope: %s not found locally. Deleting scope.\n", scope.Name)
		err := utils.SendDeleteRequest(scope.Name, utils.OIDC_SCOPES)
		if err != nil {
			utils.UpdateFailureSummary(utils.OIDC_SCOPES, scope.Name)
			log.Println("Error deleting OIDC scope: ", scope.Name, err)
		}
		utils.UpdateSuccessSummary(utils.OIDC_SCOPES, utils.DELETE)
	}
}
