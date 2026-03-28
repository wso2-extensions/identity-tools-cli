/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

package applications

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all applications to the Applications folder.
	log.Println("Exporting applications...")
	exportFilePath = filepath.Join(exportFilePath, utils.APPLICATIONS.String())
	exportAPIExists := utils.ExportAPIExists(utils.APPLICATIONS)

	if utils.IsResourceTypeExcluded(utils.APPLICATIONS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, append(getDeployedAppNames(), utils.RESIDENT_APP))
		}
	}

	apps := getAppList()
	excludeSecrets := utils.AreSecretsExcluded(utils.TOOL_CONFIGS.ApplicationConfigs)
	for _, app := range apps {
		if !utils.IsResourceExcluded(app.Name, utils.TOOL_CONFIGS.ApplicationConfigs) {
			log.Println("Exporting application: ", app.Name)
			var err error
			if exportAPIExists {
				err = exportApp(app.Id, exportFilePath, format, excludeSecrets)
			} else {
				err = exportAppWithCRUD(app.Id, app.Name, exportFilePath, format, excludeSecrets)
			}
			if err != nil {
				utils.UpdateFailureSummary(utils.APPLICATIONS, app.Name)
				log.Printf("Error while exporting application: %s. %s", app.Name, err)
			} else {
				utils.UpdateSuccessSummary(utils.APPLICATIONS, utils.EXPORT)
				log.Println("Application exported successfully: ", app.Name)
			}
		}
	}

	if !utils.IsResourceExcluded(utils.RESIDENT_APP, utils.TOOL_CONFIGS.ApplicationConfigs) {
		if err := exportResidentApp(exportFilePath, format); err != nil {
			utils.UpdateFailureSummary(utils.APPLICATIONS, utils.RESIDENT_APP)
			log.Printf("Error while exporting resident application: %s", err)
		} else {
			utils.UpdateSuccessSummary(utils.APPLICATIONS, utils.EXPORT)
			log.Println("Resident application exported successfully.")
		}
	}
}

func exportApp(appId string, outputDirPath string, format string, excludeSecrets bool) error {

	var fileType string
	// TODO: Extend support for json and xml formats.
	switch format {
	case "json":
		fileType = utils.MEDIA_TYPE_JSON
	case "xml":
		fileType = utils.MEDIA_TYPE_XML
	default:
		fileType = utils.MEDIA_TYPE_YAML
	}

	resp, err := utils.SendExportRequest(appId, fileType, utils.APPLICATIONS, excludeSecrets)
	if err != nil {
		return fmt.Errorf("error while exporting the application: %s", err)
	}
	var attachmentDetail = resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(attachmentDetail)
	if err != nil {
		return fmt.Errorf("error while parsing the content disposition header: %s", err)
	}

	fileName := params["filename"]
	exportedFileName := filepath.Join(outputDirPath, fileName)
	fileInfo := utils.GetFileInfo(exportedFileName)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response body when exporting app: %s. %s", fileName, err)
	}

	if excludeSecrets {
		body = maskOAuthConsumerSecret(body)
	}
	appKeywordMapping := getAppKeywordMapping(fileInfo.ResourceName)
	modifiedFile, err := utils.ProcessExportedContent(exportedFileName, body, appKeywordMapping, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error while processing exported data: %s", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing the exported content to file: %w", err)
	}
	return nil
}

func exportAppWithCRUD(appId, appName, outputDirPath, formatString string, excludeSecrets bool) error {

	appMap, err := getApp(appId, excludeSecrets)
	if err != nil {
		return fmt.Errorf("error while getting application: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, appName, format)

	appKeywordMapping := getAppKeywordMapping(appName)
	modifiedApp, err := utils.ProcessExportedData(appMap, exportedFileName, format, appKeywordMapping, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedApp, format, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error while serializing application: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}

func exportResidentApp(outputDirPath, formatString string) error {

	log.Println("Exporting Resident application...")

	appData, err := utils.GetResourceData(utils.APPLICATIONS, "resident")
	if err != nil {
		return fmt.Errorf("error retrieving application: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, utils.RESIDENT_APP, format)

	appKeywordMapping := getAppKeywordMapping(utils.RESIDENT_APP)
	modifiedApp, err := utils.ProcessExportedData(appData, exportedFileName, format, appKeywordMapping, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedApp, format, utils.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error while serializing application: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}
	return nil
}

func getApp(appId string, excludeSecrets bool) (map[string]interface{}, error) {

	body, err := utils.SendGetRequest(utils.APPLICATIONS, appId)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving application: %w", err)
	}

	var appStruct Application
	if err := json.Unmarshal(body, &appStruct); err != nil {
		return nil, fmt.Errorf("error unmarshalling application response: %w", err)
	}
	var appMap map[string]interface{}
	if err := json.Unmarshal(body, &appMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling application response to map: %w", err)
	}

	if err := processInboundProtocolConfigs(appId, appStruct.InboundProtocols, appMap, excludeSecrets); err != nil {
		return nil, fmt.Errorf("error while processing inbound protocol configs: %w", err)
	}

	delete(appMap, "access")
	return appMap, nil
}
