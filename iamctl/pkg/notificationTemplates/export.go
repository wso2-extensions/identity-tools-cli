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

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/notificationTemplates/applicationNotificationTemplates"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(rt utils.ResourceType, exportFilePath string, format string) {

	utils.PrintLog(utils.LogLevelInfo, rt, "", fmt.Sprintf("Exporting %s...", getTemplateLogName(rt)))
	exportFilePath = filepath.Join(exportFilePath, rt.String())

	if utils.ShouldSkip(rt) {
		return
	}

	types, err := getTemplateTypeList(rt)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, rt, "", fmt.Sprintf("Error while retrieving the list: %s", err))
		utils.MarkResTypeFailure(rt)
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, rt, "", fmt.Sprintf("Error creating %s directory: %s", getTemplateLogName(rt), err))
			utils.MarkResTypeFailure(rt)
			return
		}
	}

	var allTypeNames []string
	var typesWithTemplates []string
	for _, templateType := range types {
		allTypeNames = append(allTypeNames, templateType.DisplayName)

		if !utils.IsResourceExcluded(templateType.DisplayName, getTemplateResourceConfig(rt)) {
			utils.PrintLog(utils.LogLevelInfo, rt, templateType.DisplayName, "Exporting")
			hadTemplates, err := exportTemplateType(rt, templateType.ID, templateType.DisplayName, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(rt, templateType.DisplayName)
				utils.PrintLog(utils.LogLevelError, rt, templateType.DisplayName, fmt.Sprintf("Error while exporting: %s", err))
			} else {
				if hadTemplates {
					typesWithTemplates = append(typesWithTemplates, templateType.DisplayName)
				}
				utils.UpdateSuccessSummary(rt, utils.EXPORT)
				utils.PrintLog(utils.LogLevelInfo, rt, templateType.DisplayName, "Exported successfully")
			}
		}
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		utils.RemoveDeletedLocalDirectories(exportFilePath, typesWithTemplates)
	}

	if err := writeTemplateTypesList(exportFilePath, allTypeNames, rt, utils.FormatFromString(format)); err != nil {
		utils.PrintLog(utils.LogLevelError, rt, "", fmt.Sprintf("Error writing type list: %s", err))
		utils.UpdateFailureSummary(rt, "TemplateTypes")
	}
}

func exportTemplateType(rt utils.ResourceType, typeId, displayName, parentDir, formatString string) (bool, error) {

	format := utils.FormatFromString(formatString)
	typeDir := filepath.Join(parentDir, displayName)
	orgDir := filepath.Join(typeDir, orgTemplatesDir)

	deployedTemplates, err := getTemplatesList(rt, typeId)
	if err != nil {
		return false, fmt.Errorf("error retrieving deployed templates: %w", err)
	}

	keywordMapping := getTemplateKeywordMapping(rt, displayName)
	hadOrgTemplates := len(deployedTemplates) > 0

	if hadOrgTemplates {
		if _, err := os.Stat(orgDir); os.IsNotExist(err) {
			if err := os.MkdirAll(orgDir, 0700); err != nil {
				return false, fmt.Errorf("error creating template type directory: %w", err)
			}
		} else {
			if utils.TOOL_CONFIGS.AllowDelete {
				utils.RemoveDeletedLocalResources(orgDir, getDeployedTemplateLocales(deployedTemplates))
			}
		}

		for _, template := range deployedTemplates {
			err := exportTemplate(rt, typeId, template.Locale, orgDir, format, keywordMapping)
			if err != nil {
				return false, fmt.Errorf("error while exporting template: %s. %w", template.Locale, err)
			}
		}
	} else if utils.TOOL_CONFIGS.AllowDelete {
		if _, err := os.Stat(orgDir); err == nil {
			if err := os.RemoveAll(orgDir); err != nil {
				utils.PrintLog(utils.LogLevelError, rt, displayName, fmt.Sprintf("Error removing organization templates directory: %s", err))
			} else {
				utils.PrintLog(utils.LogLevelInfo, rt, displayName, fmt.Sprintf("Removed the directory: %s", orgTemplatesDir))
			}
		}
	}

	hadAppTemplates, err := applicationNotificationTemplates.ExportTemplateType(rt, typeId, displayName, typeDir, format, keywordMapping)
	if err != nil {
		return hadOrgTemplates, fmt.Errorf("error while exporting application templates: %w", err)
	}

	return hadOrgTemplates || hadAppTemplates, nil
}

func exportTemplate(rt utils.ResourceType, typeId, locale, typeDir string, format utils.Format, keywordMapping map[string]interface{}) error {

	templateData, err := utils.GetResourceData(rt, typeId+"/org-templates/"+locale)
	if err != nil {
		return fmt.Errorf("error while getting template: %w", err)
	}

	exportedFileName := utils.GetExportedFilePath(typeDir, locale, format)

	modifiedData, err := utils.ProcessExportedData(templateData, exportedFileName, format, keywordMapping, rt)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, rt)
	if err != nil {
		return fmt.Errorf("error while serializing template: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
