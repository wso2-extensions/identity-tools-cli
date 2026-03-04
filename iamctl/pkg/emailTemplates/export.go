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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting email templates...")
	exportFilePath = filepath.Join(exportFilePath, utils.EMAIL_TEMPLATES.String())

	if utils.IsResourceTypeExcluded(utils.EMAIL_TEMPLATES) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedTypeNames := getDeployedEmailTemplateTypeNames()
			utils.RemoveDeletedLocalDirectories(exportFilePath, deployedTypeNames)
		}
	}

	types, err := getEmailTemplateTypeList()
	if err != nil {
		log.Println("Error: when exporting email templates.", err)
	} else {
		for _, emailType := range types {
			if !utils.IsResourceExcluded(emailType.DisplayName, utils.TOOL_CONFIGS.EmailTemplateConfigs) {
				log.Println("Exporting email template type:", emailType.DisplayName)
				err := exportEmailTemplateType(emailType.ID, emailType.DisplayName, exportFilePath, format)
				if err != nil {
					utils.UpdateFailureSummary(utils.EMAIL_TEMPLATES, emailType.DisplayName)
					log.Printf("Error while exporting email template type: %s. %s", emailType.DisplayName, err)
				} else {
					utils.UpdateSuccessSummary(utils.EMAIL_TEMPLATES, utils.EXPORT)
					log.Println("Email template type exported successfully:", emailType.DisplayName)
				}
			}
		}
	}

}

func exportEmailTemplateType(typeId, displayName, parentDir, formatString string) error {

	typeDetails, err := getEmailTemplateTypeDetails(typeId)
	if err != nil {
		return fmt.Errorf("error getting template type details: %w", err)
	}

	format := utils.FormatFromString(formatString)
	typeDir := filepath.Join(parentDir, displayName)

	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(typeDir, 0700); err != nil {
			return fmt.Errorf("error creating template type directory: %w", err)
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedTemplateIds := getDeployedEmailTemplatesList(*typeDetails)
			utils.RemoveDeletedLocalResources(typeDir, deployedTemplateIds)
		}
	}

	keywordMapping := getEmailTemplateKeywordMapping(displayName)
	for _, template := range typeDetails.Templates {
		err := exportEmailTemplate(typeId, template.ID, typeDir, format, keywordMapping)
		if err != nil {
			return fmt.Errorf("error while exporting email template: %s. %w", template.ID, err)
		}
	}

	return nil
}

func exportEmailTemplate(typeId, templateId, typeDir string, format utils.Format, keywordMapping map[string]interface{}) error {

	templateData, err := utils.GetResourceData(utils.EMAIL_TEMPLATES, typeId+"/templates/"+templateId)
	if err != nil {
		return fmt.Errorf("error while getting email template: %w", err)
	}

	exportedFileName := utils.GetExportedFilePath(typeDir, templateId, format)

	modifiedData, err := utils.ProcessExportedData(templateData, exportedFileName, format, keywordMapping, utils.EMAIL_TEMPLATES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, utils.EMAIL_TEMPLATES)
	if err != nil {
		return fmt.Errorf("error while serializing email template: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
