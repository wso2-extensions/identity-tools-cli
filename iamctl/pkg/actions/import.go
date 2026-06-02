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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, "", "Importing actions...")
	importFilePath := filepath.Join(inputDirPath, utils.ACTIONS.String())

	if utils.ShouldSkip(utils.ACTIONS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, "", "No actions to import.")
		return
	}

	deployedTypes, err := getActionTypesList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ACTIONS, "", fmt.Sprintf("Error retrieving action types list: %s", err))
		utils.MarkResTypeFailure(utils.ACTIONS)
		return
	}
	typeFolders, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ACTIONS, "", fmt.Sprintf("Error reading action type directories: %s", err))
		utils.MarkResTypeFailure(utils.ACTIONS)
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedActionTypes(typeFolders, deployedTypes)
	}

	for _, typeFolder := range typeFolders {
		if !typeFolder.IsDir() {
			continue
		}
		typeName := typeFolder.Name()

		if !utils.IsResourceExcluded(typeName, utils.TOOL_CONFIGS.ActionConfigs) {
			err := importActionType(importFilePath, typeName)
			if err != nil {
				utils.UpdateFailureSummary(utils.ACTIONS, typeName)
				utils.PrintLog(utils.LogLevelError, utils.ACTIONS, typeName, fmt.Sprintf("Error importing action type: %s", err))
			}
		}
	}
}

func importActionType(importFilePath, typeName string) error {

	typeDir := filepath.Join(importFilePath, typeName)

	deployed, err := getActionsList(typeName)
	if err != nil {
		return fmt.Errorf("error retrieving deployed action list: %w", err)
	}
	localFiles, err := ioutil.ReadDir(typeDir)
	if err != nil {
		return fmt.Errorf("error reading action type directory: %w", err)
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		if err := removeDeletedDeployedActions(typeName, localFiles, deployed); err != nil {
			return fmt.Errorf("error removing deleted deployed actions: %w", err)
		}
	}

	for _, file := range localFiles {
		actionFilePath := filepath.Join(typeDir, file.Name())
		fileInfo := utils.GetFileInfo(actionFilePath)
		actionName := fileInfo.ResourceName

		actionId := getActionId(actionName, deployed)
		err := importAction(typeName, actionId, actionName, actionFilePath)
		if err != nil {
			return fmt.Errorf("error importing action %s: %w", actionName, err)
		}
	}
	return nil
}

func importAction(typeName, actionId, actionName, filePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported file format: %w", err)
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error when reading the file: %w", err)
	}

	keywordMapping := getActionsKeywordMapping(typeName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	actionMap, err := utils.DeserializeToMap([]byte(modifiedFileData), format, utils.ACTIONS, "id", "type", "createdAt", "updatedAt")
	if err != nil {
		return fmt.Errorf("error when deserializing action data: %w", err)
	}
	if err := replaceRuleReferences(actionMap); err != nil {
		return fmt.Errorf("error replacing rule references: %w", err)
	}

	status, ok := actionMap["status"].(string)
	if !ok {
		return fmt.Errorf("unexpected format for status field")
	}
	delete(actionMap, "status")

	if actionId == "" {
		return createAction(typeName, actionName, status, actionMap)
	}
	return updateAction(typeName, actionId, actionName, status, actionMap)
}

func createAction(typeName, actionName, status string, actionMap map[string]interface{}) error {

	utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, actionName, fmt.Sprintf("Creating new action of type %s", typeName))

	delete(actionMap, "version")
	jsonBody, err := utils.Serialize(actionMap, utils.FormatJSON, utils.ACTIONS)
	if err != nil {
		return fmt.Errorf("error when serializing action data: %w", err)
	}

	resp, err := utils.SendPostRequest(utils.ACTIONS, jsonBody, utils.WithPathSuffix(typeName))
	if err != nil {
		return fmt.Errorf("error when importing action: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading create response: %w", err)
	}
	var created action
	if err := json.Unmarshal(respBody, &created); err != nil {
		return fmt.Errorf("error parsing create response: %w", err)
	}

	if err := setActionStatus(typeName, created.ID, status); err != nil {
		return fmt.Errorf("error setting action status: %w", err)
	}

	utils.UpdateSuccessSummary(utils.ACTIONS, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, actionName, "Imported successfully")
	return nil
}

func updateAction(typeName, actionId, actionName, status string, actionMap map[string]interface{}) error {

	utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, actionName, fmt.Sprintf("Updating action of type %s", typeName))

	if err := addMissingFields(actionMap, typeName, actionId); err != nil {
		return fmt.Errorf("error adding missing fields: %w", err)
	}
	jsonBody, err := utils.Serialize(actionMap, utils.FormatJSON, utils.ACTIONS)
	if err != nil {
		return fmt.Errorf("error when serializing action data: %w", err)
	}

	resp, err := utils.SendPatchRequest(utils.ACTIONS, typeName+"/"+actionId, jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating action: %w", err)
	}
	defer resp.Body.Close()

	if err := setActionStatus(typeName, actionId, status); err != nil {
		return fmt.Errorf("error setting action status: %w", err)
	}

	utils.UpdateSuccessSummary(utils.ACTIONS, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, actionName, "Updated successfully")
	return nil
}

func removeDeletedDeployedActionTypes(localDirs []os.FileInfo, deployedTypes []actionType) {

	localDirNames := make(map[string]struct{})
	for _, dir := range localDirs {
		if dir.IsDir() {
			localDirNames[dir.Name()] = struct{}{}
		}
	}

	for _, deployedType := range deployedTypes {
		if _, existsLocally := localDirNames[deployedType.ID]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(deployedType.ID, utils.TOOL_CONFIGS.ActionConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, deployedType.ID, "Excluded from deletion.")
			continue
		}
		actions, err := getActionsList(deployedType.ID)
		if err != nil {
			utils.PrintLog(utils.LogLevelError, utils.ACTIONS, deployedType.ID, fmt.Sprintf("Error retrieving deployed actions: %s", err))
			continue
		}
		if err := removeDeletedDeployedActions(deployedType.ID, nil, actions); err != nil {
			utils.UpdateFailureSummary(utils.ACTIONS, deployedType.ID)
			utils.PrintLog(utils.LogLevelError, utils.ACTIONS, deployedType.ID, fmt.Sprintf("Error deleting actions: %s", err))
		}
	}
}

func removeDeletedDeployedActions(typeName string, localFiles []os.FileInfo, deployed []action) error {

	if len(deployed) == 0 {
		return nil
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, action := range deployed {
		if _, existsLocally := localResourceNames[action.Name]; existsLocally {
			continue
		}
		utils.PrintLog(utils.LogLevelInfo, utils.ACTIONS, action.Name, fmt.Sprintf("Not found locally. Deleting action of type %s.", typeName))
		if err := utils.SendDeleteRequest(typeName+"/"+action.ID, utils.ACTIONS); err != nil {
			return fmt.Errorf("error deleting action: %s. %w", action.Name, err)
		} else {
			utils.UpdateSuccessSummary(utils.ACTIONS, utils.DELETE)
		}
	}
	return nil
}
