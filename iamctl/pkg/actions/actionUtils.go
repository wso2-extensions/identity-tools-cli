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

const (
	actionTypePreUpdatePassword   = "preUpdatePassword"
	actionTypePreUpdateProfile    = "preUpdateProfile"
	actionTypePreIssueAccessToken = "preIssueAccessToken"
	actionTypePreIssueIdToken     = "preIssueIdToken"
)

type actionType struct {
	ID   string
	Self string `json:"self"`
}

type action struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	PasswordSharing struct {
		Certificate string `json:"certificate"`
	} `json:"passwordSharing"`
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

func processAuthProperties(actionMap map[string]interface{}) error {

	endpoint, ok := actionMap["endpoint"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for endpoint")
	}
	auth, ok := endpoint["authentication"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for authentication")
	}

	authType, ok := auth["type"].(string)
	if !ok {
		return fmt.Errorf("unexpected format for authentication type")
	}

	var props map[string]interface{}
	if rawProps, exists := auth["properties"]; !exists {
		props = map[string]interface{}{}
	} else {
		props, ok = rawProps.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for authentication properties")
		}
	}

	switch authType {
	case "NONE":
	case "BASIC":
		if _, exists := props["username"]; !exists {
			props["username"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		}
		props["password"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "BEARER":
		props["accessToken"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "API_KEY":
		if _, exists := props["header"]; !exists {
			props["header"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		}
		props["value"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "CLIENT_CREDENTIAL":
		props["clientSecret"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "PASSWORD_CREDENTIAL":
		props["clientSecret"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		props["password"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	default:
		return fmt.Errorf("unknown authentication type %s", authType)
	}

	auth["properties"] = props
	return nil
}

func replaceRuleReferences(actionMap map[string]interface{}) error {

	ruleRaw, exists := actionMap["rule"]
	if !exists {
		return nil
	}
	ruleMap, ok := ruleRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for rule field")
	}

	rulesRaw, exists := ruleMap["rules"]
	if !exists {
		return nil
	}
	rules, ok := rulesRaw.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for rules array")
	}
	appMap := utils.GetResourceIdentifierMap(utils.APPLICATIONS)

	for _, rule := range rules {
		ruleMap, ok := rule.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for rule element")
		}
		exprs, ok := ruleMap["expressions"].([]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for rule expressions")
		}

		for _, expr := range exprs {
			exprMap, ok := expr.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected format for expression element")
			}
			if exprMap["field"] != "application" {
				continue
			}

			oldVal, ok := exprMap["value"].(string)
			if !ok {
				return fmt.Errorf("unexpected format for value in expression")
			}
			newVal, found := appMap[oldVal]
			if !found {
				return fmt.Errorf("referenced application '%s' has not been exported", oldVal)
			}
			exprMap["value"] = newVal
		}
	}
	return nil
}

func addMissingFields(localMap map[string]interface{}, typeName, actionId string) error {

	if _, inLocal := localMap["rule"]; !inLocal {
		localMap["rule"] = map[string]interface{}{}
	}

	switch typeName {
	case actionTypePreUpdatePassword, actionTypePreUpdateProfile:
		if _, inLocal := localMap["attributes"]; !inLocal {
			localMap["attributes"] = []interface{}{}
		}
		if typeName == actionTypePreUpdatePassword {
			if err := addCertificateIfDeployed(localMap, typeName, actionId); err != nil {
				return err
			}
		}
	case actionTypePreIssueAccessToken, actionTypePreIssueIdToken:
		localEndpoint, ok := localMap["endpoint"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for endpoint")
		}
		if _, inLocal := localEndpoint["allowedHeaders"]; !inLocal {
			localEndpoint["allowedHeaders"] = []interface{}{}
		}
		if _, inLocal := localEndpoint["allowedParameters"]; !inLocal {
			localEndpoint["allowedParameters"] = []interface{}{}
		}
	}
	return nil
}

func addCertificateIfDeployed(localMap map[string]interface{}, typeName, actionId string) error {

	psRaw, exists := localMap["passwordSharing"]
	if !exists {
		return nil
	}
	localPS, ok := psRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for passwordSharing")
	}
	if _, inLocal := localPS["certificate"]; inLocal {
		return nil
	}

	body, err := utils.SendGetRequest(utils.ACTIONS, typeName+"/"+actionId)
	if err != nil {
		return fmt.Errorf("error getting deployed action data: %w", err)
	}
	var deployed action
	if _, err := utils.Deserialize(body, utils.FormatJSON, utils.ACTIONS, &deployed); err != nil {
		return fmt.Errorf("error deserializing deployed action: %w", err)
	}

	if deployed.PasswordSharing.Certificate != "" {
		localPS["certificate"] = ""
	}
	return nil
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
