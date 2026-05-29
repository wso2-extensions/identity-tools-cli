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

package emailTemplates

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	notificationTemplates "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/notificationTemplates"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	setNotificationTemplatesApiExists()
	if utils.NotificationTemplatesApiExists {
		notificationTemplates.ImportAll(utils.EMAIL_TEMPLATES, inputDirPath)
		return
	}
	ImportAllLegacyApi(inputDirPath)
}

func ImportAllLegacyApi(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, "", "Importing email templates...")
	importFilePath := filepath.Join(inputDirPath, utils.EMAIL_TEMPLATES.String())

	if utils.ShouldSkip(utils.EMAIL_TEMPLATES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, "", "No email templates to import.")
		return
	}

	deployedTypes, err := getEmailTemplateTypeList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.EMAIL_TEMPLATES, "", fmt.Sprintf("Error retrieving deployed email template types: %s", err))
		utils.MarkResTypeFailure(utils.EMAIL_TEMPLATES)
		return
	}

	localTypeDirs, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.EMAIL_TEMPLATES, "", fmt.Sprintf("Error reading email templates directory: %s", err))
		utils.MarkResTypeFailure(utils.EMAIL_TEMPLATES)
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedTypes(localTypeDirs, deployedTypes)
	}

	for _, entry := range localTypeDirs {
		if !entry.IsDir() {
			continue
		}
		displayName := entry.Name()
		localTypePath := filepath.Join(importFilePath, displayName)

		if !utils.IsResourceExcluded(displayName, utils.TOOL_CONFIGS.EmailTemplateConfigs) {
			err := importEmailTemplateType(localTypePath, displayName, deployedTypes)
			if err != nil {
				utils.UpdateFailureSummary(utils.EMAIL_TEMPLATES, displayName)
				utils.PrintLog(utils.LogLevelError, utils.EMAIL_TEMPLATES, displayName, fmt.Sprintf("Error importing: %s", err))
			}
		}
	}
}

func importEmailTemplateType(localTypePath, displayName string, deployedTypes []emailTemplateType) error {

	var typeId string
	existingType := isEmailTemplateTypeExists(displayName, deployedTypes)
	if existingType == nil {
		utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, displayName, "Creating new email template type")
		created, err := createEmailTemplateType(displayName)
		if err != nil {
			return fmt.Errorf("error creating email template type: %w", err)
		}
		typeId = created.ID
	} else {
		utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, displayName, "Updating email template type")
		typeId = existingType.ID
	}

	typeDetails, err := getEmailTemplateTypeDetails(typeId)
	if err != nil {
		return fmt.Errorf("error getting deployed templates: %w", err)
	}

	localFiles, err := ioutil.ReadDir(localTypePath)
	if err != nil {
		return fmt.Errorf("error reading local template files: %w", err)
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		err := removeDeletedDeployedTemplates(typeId, localFiles, typeDetails.Templates)
		if err != nil {
			return fmt.Errorf("error removing deleted deployed templates: %w", err)
		}
	}

	keywordMapping := getEmailTemplateKeywordMapping(displayName)

	for _, file := range localFiles {
		filePath := filepath.Join(localTypePath, file.Name())
		fileInfo := utils.GetFileInfo(filePath)
		templateId := fileInfo.ResourceName

		templateExists := isTemplateExists(templateId, typeDetails.Templates)

		err := importEmailTemplate(typeId, templateId, filePath, keywordMapping, templateExists)
		if err != nil {
			return fmt.Errorf("error importing template: %s. %w", templateId, err)
		}
	}

	if existingType != nil {
		utils.UpdateSuccessSummary(utils.EMAIL_TEMPLATES, utils.UPDATE)
		utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, displayName, "Updated successfully")
	} else {
		utils.UpdateSuccessSummary(utils.EMAIL_TEMPLATES, utils.IMPORT)
		utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, displayName, "Imported successfully")
	}

	return nil
}

func importEmailTemplate(typeId, templateId, filePath string, keywordMapping map[string]interface{}, templateExists bool) error {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for email template: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for email template: %w", err)
	}

	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	if !templateExists {
		return createTemplate(typeId, []byte(modifiedFileData), format)
	}
	return updateTemplate(typeId, templateId, []byte(modifiedFileData), format)
}

func createTemplate(typeId string, requestBody []byte, format utils.Format) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.EMAIL_TEMPLATES)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.EMAIL_TEMPLATES, jsonBody, utils.WithPathSuffix(typeId))
	if err != nil {
		return fmt.Errorf("error when creating email template: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func updateTemplate(typeId, templateId string, requestBody []byte, format utils.Format) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.EMAIL_TEMPLATES)
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(utils.EMAIL_TEMPLATES, typeId+"/templates/"+templateId, jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating email template: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func removeDeletedDeployedTypes(localDirs []os.FileInfo, deployedTypes []emailTemplateType) {

	if len(deployedTypes) == 0 {
		return
	}

	localNames := make(map[string]struct{})
	for _, dir := range localDirs {
		if dir.IsDir() {
			localNames[dir.Name()] = struct{}{}
		}
	}

	for _, deployedType := range deployedTypes {
		if _, existsLocally := localNames[deployedType.DisplayName]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(deployedType.DisplayName, utils.TOOL_CONFIGS.EmailTemplateConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, deployedType.DisplayName, "Excluded from deletion.")
			continue
		}
		utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, deployedType.DisplayName, "Not found locally. Deleting template type.")
		if err := utils.SendDeleteRequest(deployedType.ID, utils.EMAIL_TEMPLATES); err != nil {
			utils.UpdateFailureSummary(utils.EMAIL_TEMPLATES, deployedType.DisplayName)
			utils.PrintLog(utils.LogLevelError, utils.EMAIL_TEMPLATES, deployedType.DisplayName, fmt.Sprintf("Error deleting email template type: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.EMAIL_TEMPLATES, utils.DELETE)
		}
	}
}

func removeDeletedDeployedTemplates(typeId string, localFiles []os.FileInfo, deployedTemplates []emailTemplate) error {

	if len(deployedTemplates) == 0 {
		return nil
	}

	localIds := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localIds[resourceName] = struct{}{}
	}

	for _, template := range deployedTemplates {
		if _, existsLocally := localIds[template.ID]; !existsLocally {
			utils.PrintLog(utils.LogLevelInfo, utils.EMAIL_TEMPLATES, template.ID, "Not found locally. Deleting template.")
			if err := utils.SendDeleteRequest(typeId+"/templates/"+template.ID, utils.EMAIL_TEMPLATES); err != nil {
				return fmt.Errorf("error deleting email template: %w", err)
			}
		}
	}
	return nil
}
