/*
 * Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
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
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package actions

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(outputDirPath, format string) {

	log.Println("Exporting actions...")
	actionsDir := filepath.Join(outputDirPath, utils.ACTIONS.String())

	if !utils.IsEntitySupportedInVersion(utils.ACTIONS) || utils.IsResourceTypeExcluded(utils.ACTIONS) {
		return
	}
	types, err := getActionTypesList()
	if err != nil {
		log.Println("Error retrieving action types list:", err)
		return
	}

	if !utils.AreSecretsExcluded(utils.TOOL_CONFIGS.ActionConfigs) {
		log.Println("Warn: Secrets exclusion cannot be disabled for actions. All secrets will be masked.")
	}

	if _, err := os.Stat(actionsDir); os.IsNotExist(err) {
		os.MkdirAll(actionsDir, 0700)
	}

	var typesWithActions []string
	for _, at := range types {
		if utils.IsResourceExcluded(at.ID, utils.TOOL_CONFIGS.ActionConfigs) {
			continue
		}

		hadActions, err := exportActionType(at, actionsDir, format)
		if err != nil {
			utils.UpdateFailureSummary(utils.ACTIONS, at.ID)
			log.Printf("Error exporting action type %s: %s", at.ID, err)
		} else {
			if hadActions {
				typesWithActions = append(typesWithActions, at.ID)
				utils.UpdateSuccessSummary(utils.ACTIONS, utils.EXPORT)
				log.Println("Action type exported successfully:", at.ID)
			}
		}
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		utils.RemoveDeletedLocalDirectories(actionsDir, typesWithActions)
	}
}

func exportActionType(actionType actionType, parentDir, format string) (bool, error) {

	actions, err := getActionsList(actionType.ID)
	if err != nil {
		return false, fmt.Errorf("error retrieving actions list: %w", err)
	}
	if len(actions) == 0 {
		return false, nil
	}

	log.Println("Exporting action type:", actionType.ID)
	typeDir := filepath.Join(parentDir, actionType.ID)
	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(typeDir, 0700); err != nil {
			return false, fmt.Errorf("error creating action type directory: %w", err)
		}
	} else if utils.TOOL_CONFIGS.AllowDelete {
		utils.RemoveDeletedLocalResources(typeDir, getDeployedActionNames(actions))
	}

	for _, action := range actions {
		if err := exportAction(actionType.ID, action.ID, action.Name, typeDir, format); err != nil {
			return false, fmt.Errorf("error exporting action %s: %w", action.Name, err)
		}
	}
	return true, nil
}

func exportAction(typeId, actionId, actionName, outputDir, formatStr string) error {

	actionData, err := utils.GetResourceData(utils.ACTIONS, typeId+"/"+actionId)
	if err != nil {
		return fmt.Errorf("error getting action data: %w", err)
	}

	actionMap, ok := actionData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for action data")
	}
	if err := processAuthProperties(actionMap); err != nil {
		return fmt.Errorf("error processing auth properties: %w", err)
	}
	if err := replaceRuleReferences(actionMap); err != nil {
		return fmt.Errorf("error replacing rule references: %w", err)
	}

	format := utils.FormatFromString(formatStr)
	exportedFileName := utils.GetExportedFilePath(outputDir, actionName, format)

	keywordMapping := getActionsKeywordMapping(typeId)
	modifiedData, err := utils.ProcessExportedData(actionMap, exportedFileName, format, keywordMapping, utils.ACTIONS)
	if err != nil {
		return fmt.Errorf("error processing exported data: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, utils.ACTIONS)
	if err != nil {
		return fmt.Errorf("error serializing action: %w", err)
	}

	if err := ioutil.WriteFile(exportedFileName, modifiedFile, 0644); err != nil {
		return fmt.Errorf("error writing exported content to file: %w", err)
	}

	return nil
}
