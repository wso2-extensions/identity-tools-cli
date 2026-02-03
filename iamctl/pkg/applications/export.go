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
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all applications to the Applications folder.
	log.Println("Exporting applications...")
	exportFilePath = filepath.Join(exportFilePath, configs.APPLICATIONS)
	if !utils.IsEntitySupportedInVersion(configs.APPLICATIONS) {
		return
	}
	if utils.IsResourceTypeExcluded(configs.APPLICATIONS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, getDeployedAppNames())
		}
	}

	apps := getAppList()
	for _, app := range apps {
		excludeSecrets := utils.AreSecretsExcluded(utils.TOOL_CONFIGS.ApplicationConfigs)
		if !utils.IsResourceExcluded(app.Name, utils.TOOL_CONFIGS.ApplicationConfigs) {
			log.Println("Exporting application: ", app.Name)
			err := exportApp(app.Id, exportFilePath, format, excludeSecrets)
			if err != nil {
				utils.UpdateFailureSummary(configs.APPLICATIONS, app.Name)
				log.Printf("Error while exporting application: %s. %s", app.Name, err)
			} else {
				utils.UpdateSuccessSummary(configs.APPLICATIONS, utils.EXPORT)
				log.Println("Application exported successfully: ", app.Name)
			}
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

	resp, err := utils.SendExportRequest(appId, fileType, configs.APPLICATIONS, excludeSecrets)
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
	modifiedFile, err := utils.ProcessExportedContent(exportedFileName, body, appKeywordMapping, configs.APPLICATIONS)
	if err != nil {
		return fmt.Errorf("error while processing exported data: %s", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing the exported content to file: %w", err)
	}
	return nil
}
