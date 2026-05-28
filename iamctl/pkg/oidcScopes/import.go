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
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, "", "Importing OIDC scopes...")
	importFilePath := filepath.Join(inputDirPath, utils.OIDC_SCOPES.String())

	if !utils.IsEntitySupportedInOrg(utils.OIDC_SCOPES) || utils.IsResourceTypeExcluded(utils.OIDC_SCOPES) {
		return
	}
	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, "", "No OIDC scopes to import.")
		return
	}

	existingScopeList, err := getOidcScopeList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.OIDC_SCOPES, "", fmt.Sprintf("Error retrieving the deployed OIDC scope list: %s", err))
		return
	}

	files, err = ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.OIDC_SCOPES, "", fmt.Sprintf("Error importing OIDC scopes: %s", err))
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedScopes(files, existingScopeList)
	}

	for _, file := range files {
		scopeFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(scopeFilePath)
		scopeName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(scopeName, utils.TOOL_CONFIGS.OidcScopeConfigs) {
			scopeExists := isScopeExists(scopeName, existingScopeList)
			err := importOidcScope(scopeName, scopeExists, scopeFilePath)
			if err != nil {
				utils.PrintLog(utils.LogLevelError, utils.OIDC_SCOPES, scopeName, fmt.Sprintf("Error importing OIDC scope: %s", err))
				utils.UpdateFailureSummary(utils.OIDC_SCOPES, scopeName)
			}
		}
	}
}

func importOidcScope(scopeName string, scopeExists bool, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for OIDC scope: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for OIDC scope: %w", err)
	}

	scopeKeywordMapping := getOidcScopeKeywordMapping(scopeName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), scopeKeywordMapping)

	if !scopeExists {
		return importScope([]byte(modifiedFileData), format, scopeName)
	}
	return updateScope(scopeName, []byte(modifiedFileData), format, scopeName)
}

func importScope(requestBody []byte, format utils.Format, scopeName string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, scopeName, "Creating new OIDC scope")

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.OIDC_SCOPES)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.OIDC_SCOPES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when importing OIDC scope: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.OIDC_SCOPES, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, scopeName, "Imported successfully")
	return nil
}

func updateScope(scopeId string, requestBody []byte, format utils.Format, scopeName string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, scopeName, "Updating OIDC scope")

	updateBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.OIDC_SCOPES, "name")
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(utils.OIDC_SCOPES, scopeId, updateBody)
	if err != nil {
		return fmt.Errorf("error when updating OIDC scope: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.OIDC_SCOPES, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, scopeName, "Updated successfully")
	return nil
}

func removeDeletedDeployedScopes(localFiles []os.FileInfo, deployedScopes []oidcScope) {

	if len(deployedScopes) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, scope := range deployedScopes {
		if _, existsLocally := localResourceNames[scope.Name]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(scope.Name, utils.TOOL_CONFIGS.OidcScopeConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, scope.Name, "Excluded from deletion.")
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, utils.OIDC_SCOPES, scope.Name, "Not found locally. Deleting scope.")
		if err := utils.SendDeleteRequest(scope.Name, utils.OIDC_SCOPES); err != nil {
			utils.UpdateFailureSummary(utils.OIDC_SCOPES, scope.Name)
			utils.PrintLog(utils.LogLevelError, utils.OIDC_SCOPES, scope.Name, fmt.Sprintf("Error deleting OIDC scope: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.OIDC_SCOPES, utils.DELETE)
		}
	}
}
