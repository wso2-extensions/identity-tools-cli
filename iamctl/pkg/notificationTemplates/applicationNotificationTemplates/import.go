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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportTemplateType(rt utils.ResourceType, typeId, localTypePath string, keywordMapping map[string]interface{}, logName string) error {

	appsDir := filepath.Join(localTypePath, ApplicationTemplatesDir)
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		if utils.TOOL_CONFIGS.AllowDelete {
			err := removeDeployedTemplatesOfAllApps(rt, typeId)
			if err != nil {
				return fmt.Errorf("error removing deployed application templates: %w", err)
			}
		}
		return nil
	}
	appDirEntries, err := ioutil.ReadDir(appsDir)
	if err != nil {
		return fmt.Errorf("error reading application templates directory: %w", err)
	}

	appMap := utils.GetResourceIdentifierMap(utils.APPLICATIONS)
	localAppDirs := make(map[string]struct{})

	for _, appDirEntry := range appDirEntries {
		if !appDirEntry.IsDir() {
			continue
		}
		appName := appDirEntry.Name()
		appId, ok := appMap[appName]
		if !ok {
			return fmt.Errorf("referenced application with identifier '%s' has not been exported", appName)
		}
		localAppDirs[appName] = struct{}{}

		if err := importTemplatesOfApp(rt, typeId, appId, appName, appsDir, keywordMapping, logName); err != nil {
			return fmt.Errorf("error importing templates of application %s: %w", appName, err)
		}
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		if err := removeDeployedTemplatesOfDeletedApps(rt, typeId, appMap, localAppDirs); err != nil {
			return fmt.Errorf("error removing templates of deleted applications: %w", err)
		}
	}
	return nil
}

func importTemplatesOfApp(rt utils.ResourceType, typeId, appId, appName, appsDir string, keywordMapping map[string]interface{}, logName string) error {

	deployedTemplates, err := getAppTemplatesList(rt, typeId, appId)
	if err != nil {
		return fmt.Errorf("error getting templates list: %w", err)
	}
	appLocalDir := filepath.Join(appsDir, appName)
	localFiles, err := ioutil.ReadDir(appLocalDir)
	if err != nil {
		return fmt.Errorf("error reading local template files: %w", err)
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		if err := removeDeletedDeployedAppTemplates(rt, typeId, appId, appName, localFiles, deployedTemplates); err != nil {
			return fmt.Errorf("error removing deleted templates: %w", err)
		}
	}

	for _, file := range localFiles {
		filePath := filepath.Join(appLocalDir, file.Name())
		fileInfo := utils.GetFileInfo(filePath)
		locale := fileInfo.ResourceName

		exists := isAppTemplateExists(locale, deployedTemplates)
		if err := importAppTemplate(rt, typeId, appId, locale, filePath, keywordMapping, exists, logName); err != nil {
			return fmt.Errorf("error while importing template %s. %w", locale, err)
		}
	}
	return nil
}

func importAppTemplate(rt utils.ResourceType, typeId, appId, locale, filePath string, keywordMapping map[string]interface{}, exists bool, logName string) error {

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
		return createAppTemplate(rt, typeId, appId, []byte(modifiedFileData), format, logName)
	}
	return updateAppTemplate(rt, typeId, appId, locale, []byte(modifiedFileData), format, logName)
}

func createAppTemplate(rt utils.ResourceType, typeId, appId string, requestBody []byte, format utils.Format, logName string) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, rt)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(rt, jsonBody, utils.WithPathSuffix(typeId+"/app-templates/"+appId))
	if err != nil {
		return fmt.Errorf("error when creating %s: %w", logName, err)
	}
	defer resp.Body.Close()

	return nil
}

func updateAppTemplate(rt utils.ResourceType, typeId, appId, locale string, requestBody []byte, format utils.Format, logName string) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, rt, "locale")
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(rt, typeId+"/app-templates/"+appId+"/"+locale, jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating %s: %w", logName, err)
	}
	defer resp.Body.Close()

	return nil
}

func removeDeployedTemplatesOfAllApps(rt utils.ResourceType, typeId string) error {

	appMap := utils.GetResourceIdentifierMap(utils.APPLICATIONS)
	if len(appMap) == 0 {
		return nil
	}

	for appName, appId := range appMap {
		deployedTemplates, err := getAppTemplatesList(rt, typeId, appId)
		if err != nil {
			return fmt.Errorf("error getting templates for application %s: %w", appName, err)
		}
		if err := removeDeletedDeployedAppTemplates(rt, typeId, appId, appName, []os.FileInfo{}, deployedTemplates); err != nil {
			return fmt.Errorf("error removing templates for application %s: %w", appName, err)
		}
	}
	return nil
}

func removeDeployedTemplatesOfDeletedApps(rt utils.ResourceType, typeId string, appMap map[string]string, localAppDirs map[string]struct{}) error {

	for appName, appId := range appMap {
		if _, hasLocal := localAppDirs[appName]; hasLocal {
			continue
		}
		deployedTemplates, err := getAppTemplatesList(rt, typeId, appId)
		if err != nil {
			return fmt.Errorf("error getting templates for application %s: %w", appName, err)
		}
		if err := removeDeletedDeployedAppTemplates(rt, typeId, appId, appName, []os.FileInfo{}, deployedTemplates); err != nil {
			return fmt.Errorf("error removing templates for application %s: %w", appName, err)
		}
	}
	return nil
}

func removeDeletedDeployedAppTemplates(rt utils.ResourceType, typeId, appId, appName string, localFiles []os.FileInfo, deployedTemplates []appTemplate) error {

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
		log.Printf("Application template for application %s not found locally. Deleting: %s", appName, template.Locale)
		if err := utils.SendDeleteRequest(typeId+"/app-templates/"+appId+"/"+template.Locale, rt); err != nil {
			return fmt.Errorf("error deleting template: %s. %w", template.Locale, err)
		}
	}
	return nil
}
