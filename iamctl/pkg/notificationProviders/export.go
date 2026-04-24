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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func exportAll(resType utils.ResourceType, exportFilePath string, format string) {

	logName := getProviderLogName(resType)
	log.Printf("Exporting %s...", logName)
	exportFilePath = filepath.Join(exportFilePath, resType.String())

	if !utils.IsEntitySupportedInVersion(resType) || utils.IsResourceTypeExcluded(resType) {
		return
	}

	providers, err := getProviderList(resType)
	if err != nil {
		log.Printf("Error while retrieving the %s list: %s", logName, err)
		return
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, getDeployedProviderNames(providers))
		}
	}

	for _, provider := range providers {
		if !utils.IsResourceExcluded(provider.Name, getProviderResourceConfig(resType)) {
			log.Printf("Exporting %s: %s", logName, provider.Name)

			err := exportProvider(resType, logName, provider.Name, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(resType, provider.Name)
				log.Printf("Error while exporting %s: %s. %s", logName, provider.Name, err)
			} else {
				utils.UpdateSuccessSummary(resType, utils.EXPORT)
				log.Printf("%s exported successfully: %s", logName, provider.Name)
			}
		}
	}
}

func exportProvider(resType utils.ResourceType, logName string, name string, outputDirPath string, formatString string) error {

	data, err := utils.GetResourceData(resType, name)
	if err != nil {
		return fmt.Errorf("error while getting %s: %w", logName, err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, name, format)

	keywordMapping := getProviderKeywordMapping(resType, name)
	modifiedData, err := utils.ProcessExportedData(data, exportedFileName, format, keywordMapping, resType)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, resType)
	if err != nil {
		return fmt.Errorf("error while serializing %s: %w", logName, err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
