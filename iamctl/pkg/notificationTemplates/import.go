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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(rt utils.ResourceType, inputDirPath string) {

	logName := getTemplateLogName(rt)
	log.Printf("Importing %s...", logName)
	importFilePath := filepath.Join(inputDirPath, rt.String())

	if !utils.IsEntitySupportedInVersion(rt) || utils.IsResourceTypeExcluded(rt) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Printf("No %s to import.", logName)
		return
	}

	deployedTypes, err := getTemplateTypeList(rt)
	if err != nil {
		log.Printf("Error while retrieving the %s list: %s", logName, err)
		return
	}
	localTypeDirs, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Printf("Error reading %s directory: %s", logName, err)
		return
	}
	exportedTypeNames, err := readLocalTemplateTypeNames(importFilePath, rt)
	if err != nil {
		log.Printf("Error reading %s type list: %s", logName, err)
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
				log.Printf("Error: when importing %s type: %s. %s", logName, displayName, err)
			}
		}
	}
}

func importTemplateType(rt utils.ResourceType, localTypePath, displayName string, deployedTypes []notificationTemplateType, logName string) error {

	var typeId string
	existingType := getTemplateTypeId(displayName, deployedTypes)
	if existingType == "" {
		log.Printf("Creating new %s type: %s", logName, displayName)
		createdId, err := createTemplateType(rt, displayName)
		if err != nil {
			return fmt.Errorf("error creating template type: %w", err)
		}
		typeId = createdId
	} else {
		log.Printf("Updating %s type: %s", logName, displayName)
		typeId = existingType
	}

	deployedTemplates, err := getTemplatesList(rt, typeId)
	if err != nil {
		return fmt.Errorf("error getting deployed templates: %w", err)
	}
	localFiles, err := ioutil.ReadDir(localTypePath)
	if err != nil {
		return fmt.Errorf("error reading local template files: %w", err)
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		err := removeDeletedDeployedTemplates(rt, typeId, localFiles, deployedTemplates)
		if err != nil {
			return fmt.Errorf("error removing deleted deployed templates: %w", err)
		}
	}

	keywordMapping := getTemplateKeywordMapping(rt, displayName)

	for _, file := range localFiles {
		filePath := filepath.Join(localTypePath, file.Name())
		fileInfo := utils.GetFileInfo(filePath)
		locale := fileInfo.ResourceName

		exists := isTemplateExists(locale, deployedTemplates)
		err := importTemplate(rt, typeId, locale, filePath, keywordMapping, exists, logName)
		if err != nil {
			return fmt.Errorf("error importing template: %s. %w", locale, err)
		}
	}

	if existingType != "" {
		utils.UpdateSuccessSummary(rt, utils.UPDATE)
		log.Printf("Template type updated successfully: %s", displayName)
	} else {
		utils.UpdateSuccessSummary(rt, utils.IMPORT)
		log.Printf("Template type imported successfully: %s", displayName)
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
			log.Printf("%s type: %s is excluded from deletion.", logName, deployedType.DisplayName)
			continue
		}

		if _, isExported := exportedNames[deployedType.DisplayName]; isExported {
			templates, err := getTemplatesList(rt, deployedType.ID)
			if err != nil {
				log.Printf("Error checking whether templates exist for type: %s", err.Error())
				log.Printf("%s type: %s is excluded from deletion.", logName, deployedType.DisplayName)
				continue
			}
			if len(templates) == 0 {
				continue
			}
			log.Printf("%s type not found locally. Resetting: %s", logName, deployedType.DisplayName)
			if err := resetTemplateType(rt, deployedType.ID); err != nil {
				utils.UpdateFailureSummary(rt, deployedType.DisplayName)
				log.Printf("Error resetting %s type: %s. %s", logName, deployedType.DisplayName, err)
				continue
			}
		} else {
			log.Printf("%s type not found locally. Deleting: %s", logName, deployedType.DisplayName)
			if err := utils.SendDeleteRequest(deployedType.ID, rt); err != nil {
				utils.UpdateFailureSummary(rt, deployedType.DisplayName)
				log.Printf("Error deleting %s type: %s. %s", logName, deployedType.DisplayName, err)
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
		log.Printf("Template not found locally. Deleting: %s", template.Locale)
		if err := utils.SendDeleteRequest(typeId+"/org-templates/"+template.Locale, rt); err != nil {
			return fmt.Errorf("error deleting template: %s. %s", template.Locale, err)
		}
	}
	return nil
}
