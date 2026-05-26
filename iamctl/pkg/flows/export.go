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

package flows

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting flows...")
	exportFilePath = filepath.Join(exportFilePath, utils.FLOWS.String())

	if !utils.IsEntitySupportedInVersion(utils.FLOWS) || !utils.IsEntitySupportedInOrg(utils.FLOWS) || utils.IsResourceTypeExcluded(utils.FLOWS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			log.Println("Error creating flows directory:", err)
			return
		}
	}

	var exportedFlowNames []string
	for name, id := range flowTypes {
		if !utils.IsResourceExcluded(name, utils.TOOL_CONFIGS.FlowConfigs) {
			log.Println("Exporting flow:", name)

			exists, err := exportFlow(name, id, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.FLOWS, name)
				log.Printf("Error while exporting flow: %s. %s", name, err)
			} else {
				if exists {
					exportedFlowNames = append(exportedFlowNames, name)
					utils.UpdateSuccessSummary(utils.FLOWS, utils.EXPORT)
					log.Println("Flow exported successfully:", name)
				} else {
					log.Printf("Flow: %s is not configured", name)
				}
			}
		}
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		utils.RemoveDeletedLocalResources(exportFilePath, exportedFlowNames)
	}
}

func exportFlow(name, id string, outputDirPath string, formatString string) (exists bool, err error) {

	if name == invitedUserRegistrationFlowName {
		if _, exists := utils.GetResourceIdentifierMap(utils.GOVERNANCE_CONNECTORS)[utils.USER_ONBOARDING_GOVERNANCE_CATEGORY_ID]; !exists {
			return false, fmt.Errorf("required resource %s governance connector category has not been exported", utils.USER_ONBOARDING_GOVERNANCE_CATEGORY_NAME)
		}
	}

	flowData, exists, err := getFlowData(id)
	if err != nil {
		return false, fmt.Errorf("error while getting flow data: %w", err)
	}
	if !exists {
		return false, nil
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, name, format)
	keywordMapping := getFlowKeywordMapping(name)

	modifiedData, err := utils.ProcessExportedData(flowData, exportedFileName, format, keywordMapping, utils.FLOWS)
	if err != nil {
		return false, fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, utils.FLOWS)
	if err != nil {
		return false, fmt.Errorf("error while serializing flow: %w", err)
	}

	if err := ioutil.WriteFile(exportedFileName, modifiedFile, 0644); err != nil {
		return false, fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return true, nil
}

func getFlowData(id string) (flow map[string]interface{}, exists bool, err error) {

	flowData, err := utils.GetResourceData(utils.FLOWS, "", utils.WithQueryParams(map[string]string{"flowType": id}))
	if err != nil {
		return nil, false, fmt.Errorf("error while retrieving flow: %w", err)
	}
	flowMap, ok := flowData.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("unexpected format for flow data")
	}
	steps, exists := flowMap["steps"]
	if !exists {
		return nil, false, nil
	}

	configData, err := utils.GetResourceData(utils.FLOWS, "config", utils.WithQueryParams(map[string]string{"flowType": id}))
	if err != nil {
		return nil, false, fmt.Errorf("error while retrieving flow config: %w", err)
	}
	configMap, ok := configData.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("unexpected format for flow config data")
	}
	configMap["steps"] = steps

	return configMap, true, nil
}
