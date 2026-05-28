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

package applicationNotificationTemplates

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportTemplateType(rt utils.ResourceType, typeId, displayName, typeDir string, format utils.Format, keywordMapping map[string]interface{}) (bool, error) {

	appMap := utils.GetResourceIdentifierMap(utils.APPLICATIONS)
	if len(appMap) == 0 {
		return false, nil
	}

	appsDir := filepath.Join(typeDir, ApplicationTemplatesDir)
	appsDirExistedBefore := false
	if _, err := os.Stat(appsDir); err == nil {
		appsDirExistedBefore = true
	}

	var appsWithTemplates []string
	for appId, appName := range appMap {
		hadAppTemplates, err := exportTemplatesOfApp(rt, typeId, appId, appName, appsDir, format, keywordMapping)
		if err != nil {
			return false, fmt.Errorf("error exporting templates of application %s: %w", appName, err)
		}
		if hadAppTemplates {
			appsWithTemplates = append(appsWithTemplates, appName)
		}
	}

	if utils.TOOL_CONFIGS.AllowDelete && appsDirExistedBefore {
		utils.RemoveDeletedLocalDirectories(appsDir, appsWithTemplates)
		if len(appsWithTemplates) == 0 {
			if err := os.Remove(appsDir); err != nil {
				utils.PrintLog(utils.LogLevelError, rt, displayName, fmt.Sprintf("Error removing application templates directory: %s", err))
			} else {
				utils.PrintLog(utils.LogLevelInfo, rt, displayName, fmt.Sprintf("Removed the directory: %s", ApplicationTemplatesDir))
			}
		}
	}
	return len(appsWithTemplates) > 0, nil
}

func exportTemplatesOfApp(rt utils.ResourceType, typeId, appId, appName, appsDir string, format utils.Format, keywordMapping map[string]interface{}) (bool, error) {

	templates, err := getAppTemplatesList(rt, typeId, appId)
	if err != nil {
		return false, fmt.Errorf("error retrieving templates list: %w", err)
	}
	if len(templates) == 0 {
		return false, nil
	}

	appDir := filepath.Join(appsDir, appName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appDir, 0700); err != nil {
			return false, fmt.Errorf("error creating template directory: %w", err)
		}
	} else if utils.TOOL_CONFIGS.AllowDelete {
		utils.RemoveDeletedLocalResources(appDir, getDeployedAppTemplateLocales(templates))
	}

	for _, template := range templates {
		if err := exportAppTemplate(rt, typeId, appId, template.Locale, appDir, format, keywordMapping); err != nil {
			return false, fmt.Errorf("error while exporting template %s. %w", template.Locale, err)
		}
	}
	return true, nil
}

func exportAppTemplate(rt utils.ResourceType, typeId, appId, locale, appDir string, format utils.Format, keywordMapping map[string]interface{}) error {

	templateData, err := utils.GetResourceData(rt, typeId+"/app-templates/"+appId+"/"+locale)
	if err != nil {
		return fmt.Errorf("error while getting template: %w", err)
	}

	exportedFileName := utils.GetExportedFilePath(appDir, locale, format)

	modifiedData, err := utils.ProcessExportedData(templateData, exportedFileName, format, keywordMapping, rt)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, rt)
	if err != nil {
		return fmt.Errorf("error while serializing template: %w", err)
	}
	if err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644); err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
