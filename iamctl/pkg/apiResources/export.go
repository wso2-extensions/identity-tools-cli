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

package apiResources

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	utils.PrintLog(utils.LogLevelInfo, utils.API_RESOURCES, "", "Exporting API resources...")
	exportFilePath = filepath.Join(exportFilePath, utils.API_RESOURCES.String())

	if utils.ShouldSkip(utils.API_RESOURCES) {
		return
	}
	resources, err := GetApiResourceList(true)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.API_RESOURCES, "", fmt.Sprintf("Error when retrieving API resource list: %s", err))
		utils.MarkResTypeFailure(utils.API_RESOURCES)
		return
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, utils.API_RESOURCES, "", fmt.Sprintf("Error creating API resources directory: %s", err))
			utils.MarkResTypeFailure(utils.API_RESOURCES)
			return
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedIdentifiers := getDeployedApiResourceIdentifiers(resources)
			utils.RemoveDeletedLocalResources(exportFilePath, deployedIdentifiers)
		}
	}

	exportedScopesMap = map[string]string{}
	successCount := 0

	for _, resource := range resources {
		if !utils.IsResourceExcluded(resource.Identifier, utils.TOOL_CONFIGS.ApiResourceConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.API_RESOURCES, resource.Identifier, "Exporting")
			err := exportApiResource(resource.ID, resource.Identifier, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.API_RESOURCES, resource.Identifier)
				utils.PrintLog(utils.LogLevelError, utils.API_RESOURCES, resource.Identifier, fmt.Sprintf("Error while exporting: %s", err))
			} else {
				successCount++
				utils.AddToIdentifierMap(utils.API_RESOURCES, resource.ID, resource.Identifier, utils.EXPORT)
				utils.PrintLog(utils.LogLevelInfo, utils.API_RESOURCES, resource.Identifier, "Exported successfully")
			}
		}
	}

	err = writeScopesMap(exportFilePath, exportedScopesMap, format)
	updateApiResourceExportSummary(err == nil, successCount)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.API_RESOURCES, "", fmt.Sprintf("Error writing scope name map: %s", err))
	}
}

func exportApiResource(resourceId string, resourceIdentifier string, outputDirPath string, formatString string) error {

	resourceData, err := utils.GetResourceData(utils.API_RESOURCES, resourceId)
	if err != nil {
		return fmt.Errorf("error while getting API resource: %w", err)
	}
	resourceMap, ok := resourceData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for API resource data")
	}

	scopeNames, err := processScopes(resourceMap)
	if err != nil {
		return fmt.Errorf("error while processing scopes: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, resourceIdentifier, format)

	keywordMapping := getApiResourceKeywordMapping(resourceIdentifier)
	modifiedResource, err := utils.ProcessExportedData(resourceMap, exportedFileName, format, keywordMapping, utils.API_RESOURCES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedResource, format, utils.API_RESOURCES)
	if err != nil {
		return fmt.Errorf("error while serializing API resource: %w", err)
	}

	if err := ioutil.WriteFile(exportedFileName, modifiedFile, 0644); err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	for _, scopeName := range scopeNames {
		exportedScopesMap[scopeName] = resourceIdentifier
	}

	return nil
}
