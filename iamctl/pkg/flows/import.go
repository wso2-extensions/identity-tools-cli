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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.FLOWS, "", "Importing flows...")
	importFilePath := filepath.Join(inputDirPath, utils.FLOWS.String())

	if !utils.IsEntitySupportedInVersion(utils.FLOWS) || !utils.IsEntitySupportedInOrg(utils.FLOWS) || utils.IsResourceTypeExcluded(utils.FLOWS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.FLOWS, "", "No flows to import.")
		return
	}

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.FLOWS, "", fmt.Sprintf("Error reading flows directory: %s", err))
		return
	}

	for _, file := range files {
		flowFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(flowFilePath)
		name := fileInfo.ResourceName

		if !utils.IsResourceExcluded(name, utils.TOOL_CONFIGS.FlowConfigs) {
			id, ok := flowTypes[name]
			if !ok {
				utils.PrintLog(utils.LogLevelError, utils.FLOWS, name, "Error importing flow: unknown flow type")
				utils.UpdateFailureSummary(utils.FLOWS, name)
				continue
			}

			err := importFlow(name, id, flowFilePath)
			if err != nil {
				utils.UpdateFailureSummary(utils.FLOWS, name)
				utils.PrintLog(utils.LogLevelError, utils.FLOWS, name, fmt.Sprintf("Error importing flow: %s", err))
			}
		}
	}
}

func importFlow(name, id, importFilePath string) error {

	if name == invitedUserRegistrationFlowName {
		if _, exists := utils.GetResourceIdentifierMap(utils.GOVERNANCE_CONNECTORS)[utils.USER_ONBOARDING_GOVERNANCE_CATEGORY_NAME]; !exists {
			return fmt.Errorf("required resource %s governance connector category has not been imported", utils.USER_ONBOARDING_GOVERNANCE_CATEGORY_NAME)
		}
	}

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for flow: %w", err)
	}
	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file: %w", err)
	}

	keywordMapping := getFlowKeywordMapping(name)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	return updateFlow(name, id, []byte(modifiedFileData), format)
}

func updateFlow(name, id string, data []byte, format utils.Format) error {

	utils.PrintLog(utils.LogLevelInfo, utils.FLOWS, name, "Updating flow")

	dataMap, err := utils.DeserializeToMap(data, format, utils.FLOWS)
	if err != nil {
		return fmt.Errorf("error when deserializing file: %w", err)
	}
	steps, ok := dataMap["steps"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for flow steps")
	}

	flowJSON, err := json.Marshal(map[string]interface{}{
		"flowType": id,
		"steps":    steps,
	})
	if err != nil {
		return fmt.Errorf("error when marshalling flow request body: %w", err)
	}

	resp, err := utils.SendPutRequest(utils.FLOWS, "", flowJSON)
	if err != nil {
		return fmt.Errorf("error when updating flow: %w", err)
	}
	resp.Body.Close()

	delete(dataMap, "steps")
	if err := updateFlowConfig(dataMap); err != nil {
		return err
	}

	utils.UpdateSuccessSummary(utils.FLOWS, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.FLOWS, name, "Updated successfully")
	return nil
}

func updateFlowConfig(dataMap map[string]interface{}) error {

	configJSON, err := json.Marshal(dataMap)
	if err != nil {
		return fmt.Errorf("error when marshalling flow config request body: %w", err)
	}

	resp, err := utils.SendPatchRequest(utils.FLOWS, "config", configJSON)
	if err != nil {
		return fmt.Errorf("error when updating flow config: %w", err)
	}
	resp.Body.Close()
	return nil
}
