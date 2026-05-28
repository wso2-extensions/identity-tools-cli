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

package notificationTemplates

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	applicationNotificationTemplates "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/notificationTemplates/applicationNotificationTemplates"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(rt utils.ResourceType, inputDirPath string) {

	logName := getTemplateLogName(rt)
	utils.PrintLog(utils.LogLevelInfo, rt, "", fmt.Sprintf("Importing %s...", logName))
	importFilePath := filepath.Join(inputDirPath, rt.String())

	if !utils.IsEntitySupportedInVersion(rt) || !utils.IsEntitySupportedInOrg(rt) || utils.IsResourceTypeExcluded(rt) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, rt, "", fmt.Sprintf("No %s to import.", logName))
		return
	}

	deployedTypes, err := getTemplateTypeList(rt)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, rt, "", fmt.Sprintf("Error while retrieving the list: %s", err))
		return
	}
	localTypeDirs, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, rt, "", fmt.Sprintf("Error reading directory: %s", err))
		return
	}
	exportedTypeNames, err := readLocalTemplateTypeNames(importFilePath, rt)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, rt, "", fmt.Sprintf("Error reading type list: %s", err))
		utils.UpdateFailureSummary(rt, "TemplateTypes")
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedTypes(rt, localTypeDirs, deployedTypes, exportedTypeNames, logName)
	}

	for _, entry := range localTypeDirs {
		if !entry.IsDir() {
			continue
		}
		displayName := entry.Name()
		localTypePath := filepath.Join(importFilePath, displayName)

		if !utils.IsResourceExcluded(displayName, getTemplateResourceConfig(rt)) {
			err := importTemplateType(rt, localTypePath, displayName, deployedTypes, logName)
			if err != nil {
				utils.UpdateFailureSummary(rt, displayName)
				utils.PrintLog(utils.LogLevelError, rt, displayName, fmt.Sprintf("Error when importing: %s", err))
			}
		}
	}
}

func importTemplateType(rt utils.ResourceType, localTypePath, displayName string, deployedTypes []notificationTemplateType, logName string) error {

	var typeId string
	existingType := getTemplateTypeId(displayName, deployedTypes)
	if existingType == "" {
		utils.PrintLog(utils.LogLevelInfo, rt, displayName, fmt.Sprintf("Creating new %s type", logName))
		createdId, err := createTemplateType(rt, displayName)
		if err != nil {
			return fmt.Errorf("error creating template type: %w", err)
		}
		typeId = createdId
	} else {
		utils.PrintLog(utils.LogLevelInfo, rt, displayName, fmt.Sprintf("Updating %s type", logName))
		typeId = existingType
	}

	deployedTemplates, err := getTemplatesList(rt, typeId)
	if err != nil {
		return fmt.Errorf("error getting deployed templates: %w", err)
	}

	orgDir := filepath.Join(localTypePath, orgTemplatesDir)
	var localFiles []os.FileInfo
	if _, err := os.Stat(orgDir); os.IsNotExist(err) {
		localFiles = []os.FileInfo{}
	} else {
		localFiles, err = ioutil.ReadDir(orgDir)
		if err != nil {
			return fmt.Errorf("error reading local template files: %w", err)
		}
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		err := removeDeletedDeployedTemplates(rt, typeId, localFiles, deployedTemplates)
		if err != nil {
			return fmt.Errorf("error removing deleted deployed templates: %w", err)
		}
	}

	keywordMapping := getTemplateKeywordMapping(rt, displayName)

	for _, file := range localFiles {
		filePath := filepath.Join(orgDir, file.Name())
		fileInfo := utils.GetFileInfo(filePath)
		locale := fileInfo.ResourceName

		exists := isTemplateExists(locale, deployedTemplates)
		err := importTemplate(rt, typeId, locale, filePath, keywordMapping, exists, logName)
		if err != nil {
			return fmt.Errorf("error importing template: %s. %w", locale, err)
		}
	}

	if err := applicationNotificationTemplates.ImportTemplateType(rt, typeId, localTypePath, keywordMapping, logName); err != nil {
		return fmt.Errorf("error while importing application templates: %w", err)
	}

	if existingType != "" {
		utils.UpdateSuccessSummary(rt, utils.UPDATE)
		utils.PrintLog(utils.LogLevelInfo, rt, displayName, "Updated successfully")
	} else {
		utils.UpdateSuccessSummary(rt, utils.IMPORT)
		utils.PrintLog(utils.LogLevelInfo, rt, displayName, "Imported successfully")
	}

	return nil
}

func importTemplate(rt utils.ResourceType, typeId, locale, filePath string, keywordMapping map[string]interface{}, exists bool, logName string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported file format: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error when reading the file: %w", err)
	}

	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	if !exists {
		return createTemplate(rt, typeId, []byte(modifiedFileData), format, logName)
	}
	return updateTemplate(rt, typeId, locale, []byte(modifiedFileData), format, logName)
}

func createTemplate(rt utils.ResourceType, typeId string, requestBody []byte, format utils.Format, logName string) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, rt)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(rt, jsonBody, utils.WithPathSuffix(typeId+"/org-templates"))
	if err != nil {
		return fmt.Errorf("error when creating %s: %w", logName, err)
	}
	defer resp.Body.Close()

	return nil
}

func updateTemplate(rt utils.ResourceType, typeId, locale string, requestBody []byte, format utils.Format, logName string) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, rt, "locale")
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(rt, typeId+"/org-templates/"+locale, jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating %s: %w", logName, err)
	}
	defer resp.Body.Close()

	return nil
}

func removeDeletedDeployedTypes(rt utils.ResourceType, localDirs []os.FileInfo, deployedTypes []notificationTemplateType, exportedTypeNames []string, logName string) {

	if len(deployedTypes) == 0 {
		return
	}

	localDirNames := make(map[string]struct{})
	for _, dir := range localDirs {
		if dir.IsDir() {
			localDirNames[dir.Name()] = struct{}{}
		}
	}
	exportedNames := make(map[string]struct{})
	for _, name := range exportedTypeNames {
		exportedNames[name] = struct{}{}
	}

	for _, deployedType := range deployedTypes {
		if _, existsLocally := localDirNames[deployedType.DisplayName]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(deployedType.DisplayName, getTemplateResourceConfig(rt)) {
			utils.PrintLog(utils.LogLevelInfo, rt, deployedType.DisplayName, fmt.Sprintf("%s type excluded from deletion.", logName))
			continue
		}

		if _, isExported := exportedNames[deployedType.DisplayName]; isExported {
			utils.PrintLog(utils.LogLevelInfo, rt, deployedType.DisplayName, fmt.Sprintf("%s type not found locally. Resetting.", logName))
			if err := resetTemplateType(rt, deployedType.ID); err != nil {
				utils.UpdateFailureSummary(rt, deployedType.DisplayName)
				utils.PrintLog(utils.LogLevelError, rt, deployedType.DisplayName, fmt.Sprintf("Error resetting %s type: %s", logName, err))
				continue
			}
		} else {
			utils.PrintLog(utils.LogLevelInfo, rt, deployedType.DisplayName, fmt.Sprintf("%s type not found locally. Deleting.", logName))
			if err := utils.SendDeleteRequest(deployedType.ID, rt); err != nil {
				utils.UpdateFailureSummary(rt, deployedType.DisplayName)
				utils.PrintLog(utils.LogLevelError, rt, deployedType.DisplayName, fmt.Sprintf("Error deleting %s type: %s", logName, err))
				continue
			}
		}
		utils.UpdateSuccessSummary(rt, utils.DELETE)
	}
}

func removeDeletedDeployedTemplates(rt utils.ResourceType, typeId string, localFiles []os.FileInfo, deployedTemplates []notificationTemplate) error {

	if len(deployedTemplates) == 0 {
		return nil
	}

	localLocales := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localLocales[resourceName] = struct{}{}
	}

	for _, template := range deployedTemplates {
		if _, existsLocally := localLocales[template.Locale]; existsLocally {
			continue
		}
		utils.PrintLog(utils.LogLevelInfo, rt, template.Locale, "Template not found locally. Deleting.")
		if err := utils.SendDeleteRequest(typeId+"/org-templates/"+template.Locale, rt); err != nil {
			return fmt.Errorf("error deleting template: %s. %w", template.Locale, err)
		}
	}
	return nil
}
