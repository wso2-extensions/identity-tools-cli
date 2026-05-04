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
	"path"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type actionType struct {
	ID   string
	Self string `json:"self"`
}

type action struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getActionTypesList() ([]actionType, error) {

	body, err := utils.SendGetRequest(utils.ACTIONS, "types")
	if err != nil {
		return nil, fmt.Errorf("error when getting the list: %w", err)
	}
	var types []actionType
	if err := json.Unmarshal(body, &types); err != nil {
		return nil, fmt.Errorf("error when unmarshalling the list: %w", err)
	}
	for i := range types {
		types[i].ID = typeIdFromSelf(types[i].Self)
	}
	return types, nil
}

func getActionsList(actionType string) ([]action, error) {

	body, err := utils.SendGetRequest(utils.ACTIONS, actionType)
	if err != nil {
		return nil, err
	}
	var summaries []action
	if err := json.Unmarshal(body, &summaries); err != nil {
		return nil, fmt.Errorf("error unmarshalling actions list: %w", err)
	}
	return summaries, nil
}

func getDeployedActionTypeIds(types []actionType) []string {

	var ids []string
	for _, at := range types {
		ids = append(ids, at.ID)
	}
	return ids
}

func getDeployedActionNames(summaries []action) []string {

	var names []string
	for _, s := range summaries {
		names = append(names, s.Name)
	}
	return names
}

func getActionsKeywordMapping(typeName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ActionConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(typeName, utils.KEYWORD_CONFIGS.ActionConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getActionId(name string, existingActionList []action) string {

	for i := range existingActionList {
		if existingActionList[i].Name == name {
			return existingActionList[i].ID
		}
	}
	return ""
}

func typeIdFromSelf(self string) string {

	return path.Base(self)
}

func setActionStatus(typePath, actionId, status string) error {

	endpoint := typePath + "/" + actionId + "/"
	switch status {
	case "ACTIVE":
		endpoint += "activate"
	case "INACTIVE":
		endpoint += "deactivate"
	default:
		return fmt.Errorf("unexpected value for status: %s", status)
	}

	resp, err := utils.SendPostRequest(utils.ACTIONS, nil, utils.WithPathSuffix(endpoint))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
