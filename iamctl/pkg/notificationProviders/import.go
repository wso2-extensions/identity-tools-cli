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

package notificationProviders

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func importAll(resType utils.ResourceType, inputDirPath string) {

	logName := getProviderLogName(resType)
	utils.PrintLog(utils.LogLevelInfo, resType, "", fmt.Sprintf("Importing %s...", logName))
	importFilePath := filepath.Join(inputDirPath, resType.String())

	if resType == utils.EMAIL_PROVIDERS && utils.SERVER_CONFIGS.TenantDomain == utils.DEFAULT_TENANT_DOMAIN {
		utils.PrintLog(utils.LogLevelInfo, resType, "", "Importing email providers for super tenant not supported.")
		return
	}
	if utils.ShouldSkip(resType) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, resType, "", fmt.Sprintf("No %s to import.", logName))
		return
	}

	existingProviderList, err := getProviderList(resType)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, resType, "", fmt.Sprintf("Error retrieving the deployed %s list: %s", logName, err))
		utils.MarkResTypeFailure(resType)
		return
	}
	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, resType, "", fmt.Sprintf("Error reading %s directory: %s", logName, err))
		utils.MarkResTypeFailure(resType)
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedProviders(resType, files, existingProviderList, logName)
	}

	for _, file := range files {
		providerFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(providerFilePath)
		providerName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(providerName, getProviderResourceConfig(resType)) {
			providerExists := isProviderExists(providerName, existingProviderList)
			err := importProvider(resType, logName, providerName, providerExists, providerFilePath)
			if err != nil {
				utils.PrintLog(utils.LogLevelError, resType, providerName, fmt.Sprintf("Error importing %s: %s", logName, err))
				utils.UpdateFailureSummary(resType, providerName)
			}
		}
	}
}

func importProvider(resType utils.ResourceType, logName string, name string, exists bool, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for %s: %w", logName, err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for %s: %w", logName, err)
	}

	keywordMapping := getProviderKeywordMapping(resType, name)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	if !exists {
		return createProvider(resType, []byte(modifiedFileData), format, name, logName)
	}
	return updateProvider(resType, name, []byte(modifiedFileData), format, logName)
}

func createProvider(resType utils.ResourceType, requestBody []byte, format utils.Format, name string, logName string) error {

	utils.PrintLog(utils.LogLevelInfo, resType, name, fmt.Sprintf("Creating new %s", logName))

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, resType)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(resType, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating %s: %w", logName, err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(resType, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, resType, name, fmt.Sprintf("%s created successfully", logName))
	return nil
}

func updateProvider(resType utils.ResourceType, name string, requestBody []byte, format utils.Format, logName string) error {

	utils.PrintLog(utils.LogLevelInfo, resType, name, fmt.Sprintf("Updating %s", logName))

	updateBody, err := utils.PrepareJSONRequestBody(requestBody, format, resType, "name")
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(resType, name, updateBody)
	if err != nil {
		return fmt.Errorf("error when updating %s: %w", logName, err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(resType, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, resType, name, fmt.Sprintf("%s updated successfully", logName))
	return nil
}

func removeDeletedDeployedProviders(resType utils.ResourceType, localFiles []os.FileInfo, deployedProviders []notificationProvider, logName string) {

	if len(deployedProviders) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, provider := range deployedProviders {
		if _, existsLocally := localResourceNames[provider.Name]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(provider.Name, getProviderResourceConfig(resType)) {
			utils.PrintLog(utils.LogLevelInfo, resType, provider.Name, fmt.Sprintf("%s is excluded from deletion", logName))
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, resType, provider.Name, fmt.Sprintf("%s not found locally. Deleting.", logName))
		if err := utils.SendDeleteRequest(provider.Name, resType); err != nil {
			utils.UpdateFailureSummary(resType, provider.Name)
			utils.PrintLog(utils.LogLevelError, resType, provider.Name, fmt.Sprintf("Error deleting %s: %s", logName, err))
		} else {
			utils.UpdateSuccessSummary(resType, utils.DELETE)
		}
	}
}
