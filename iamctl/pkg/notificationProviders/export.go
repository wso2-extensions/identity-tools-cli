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
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func exportAll(resType utils.ResourceType, exportFilePath string, format string) {

	logName := getProviderLogName(resType)
	utils.PrintLog(utils.LogLevelInfo, resType, "", fmt.Sprintf("Exporting %s...", logName))
	exportFilePath = filepath.Join(exportFilePath, resType.String())

	if resType == utils.EMAIL_PROVIDERS && utils.SERVER_CONFIGS.TenantDomain == utils.DEFAULT_TENANT_DOMAIN {
		utils.PrintLog(utils.LogLevelInfo, resType, "", "Exporting email providers for super tenant not supported.")
		return
	}
	if !utils.IsEntitySupportedInVersion(resType) || !utils.IsEntitySupportedInOrg(resType) || utils.IsResourceTypeExcluded(resType) {
		return
	}

	providers, err := getProviderList(resType)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, resType, "", fmt.Sprintf("Error while retrieving the %s list: %s", logName, err))
		return
	}

	if resType == utils.EMAIL_PROVIDERS && !utils.AreSecretsExcluded(utils.TOOL_CONFIGS.EmailProviderConfigs) {
		utils.PrintLog(utils.LogLevelWarn, resType, "", "Secrets exclusion cannot be disabled for email providers. All secrets will be masked.")
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, resType, "", fmt.Sprintf("Error creating %s directory: %s", logName, err))
			return
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, getDeployedProviderNames(providers))
		}
	}

	for _, provider := range providers {
		if !utils.IsResourceExcluded(provider.Name, getProviderResourceConfig(resType)) {
			utils.PrintLog(utils.LogLevelInfo, resType, provider.Name, fmt.Sprintf("Exporting %s", logName))

			err := exportProvider(resType, logName, provider.Name, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(resType, provider.Name)
				utils.PrintLog(utils.LogLevelError, resType, provider.Name, fmt.Sprintf("Error while exporting %s: %s", logName, err))
			} else {
				utils.UpdateSuccessSummary(resType, utils.EXPORT)
				utils.PrintLog(utils.LogLevelInfo, resType, provider.Name, fmt.Sprintf("%s exported successfully", logName))
			}
		}
	}
}

func exportProvider(resType utils.ResourceType, logName string, name string, outputDirPath string, formatString string) error {

	data, err := utils.GetResourceData(resType, name)
	if err != nil {
		return fmt.Errorf("error while getting %s: %w", logName, err)
	}

	data, err = processAuthSecrets(resType, data)
	if err != nil {
		return fmt.Errorf("error while processing auth secrets: %w", err)
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
