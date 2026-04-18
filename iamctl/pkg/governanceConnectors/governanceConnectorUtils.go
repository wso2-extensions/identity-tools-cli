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

package governanceConnectors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

const passwordExpiryConnectorId = "cGFzc3dvcmRFeHBpcnk"

type connectorCategory struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type connector struct {
	Id           string `json:"id"`
	FriendlyName string `json:"friendlyName"`
}

func getCategoryList() ([]connectorCategory, error) {

	var categories []connectorCategory
	resp, err := utils.SendGetListRequest(utils.GOVERNANCE_CONNECTORS, -1)
	if err != nil {
		return nil, fmt.Errorf("error retrieving governance connector category list: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading governance connector category list response: %w", err)
		}
		if err := json.Unmarshal(body, &categories); err != nil {
			return nil, fmt.Errorf("error unmarshalling governance connector categories: %w", err)
		}
		return categories, nil
	} else if errMsg, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error retrieving governance connector category list. Status code: %d, Error: %s", statusCode, errMsg)
	}
	return nil, fmt.Errorf("error retrieving governance connector category list")
}

func getConnectorListForCategory(categoryId string) ([]connector, error) {

	body, err := utils.SendGetRequest(utils.GOVERNANCE_CONNECTORS, categoryId+"/connectors")
	if err != nil {
		return nil, fmt.Errorf("error retrieving connectors list: %w", err)
	}

	var connectors []connector
	if err := json.Unmarshal(body, &connectors); err != nil {
		return nil, fmt.Errorf("error unmarshalling connectors list: %w", err)
	}
	return connectors, nil
}

func isCategoryExists(catName string, categories []connectorCategory) *connectorCategory {

	for i := range categories {
		if categories[i].Name == catName {
			return &categories[i]
		}
	}
	return nil
}

func getConnectorId(connectorName string, connectors []connector) string {

	for i := range connectors {
		if connectors[i].FriendlyName == connectorName {
			return connectors[i].Id
		}
	}
	return ""
}

func getDeployedConnectorNames(connectors []connector) []string {

	var names []string
	for _, c := range connectors {
		names = append(names, c.FriendlyName)
	}
	return names
}

func getDeployedCategoryNames() []string {

	categories, err := getCategoryList()
	if err != nil {
		return []string{}
	}

	var catNames []string
	for _, cat := range categories {
		catNames = append(catNames, cat.Name)
	}
	return catNames
}

func getGovernanceCategoryKeywordMapping(categoryName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.GovernanceConnectorConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(categoryName, utils.KEYWORD_CONFIGS.GovernanceConnectorConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func processPasswordExpiryConnector(data interface{}) error {

	connectorMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for connector data")
	}
	props, ok := connectorMap["properties"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for properties")
	}

	roleMap := utils.GetResourceIdentifierMap(utils.ROLES)
	var filtered []interface{}

	for _, item := range props {
		propMap, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for property")
		}

		name, ok := propMap["name"].(string)
		if !ok {
			return fmt.Errorf("unexpected format for property name")
		}
		if !strings.HasPrefix(name, "passwordExpiry.rule") {
			filtered = append(filtered, item)
			continue
		}

		value, ok := propMap["value"].(string)
		if !ok {
			return fmt.Errorf("unexpected format for property value in property: %s", name)
		}
		parts := strings.Split(value, ",")
		if len(parts) < 5 {
			return fmt.Errorf("unexpected format for rule in property %s: expected 5 fields", name)
		}

		switch parts[2] {
		case "groups":
			continue
		case "roles":
			identifier := parts[4]
			replacement, exists := roleMap[identifier]
			if !exists {
				return fmt.Errorf("referenced Role with identifier '%s' has not been exported", identifier)
			}
			parts[4] = replacement
			propMap["value"] = strings.Join(parts, ",")
			filtered = append(filtered, item)
		default:
			return fmt.Errorf("unexpected rule type %s in property: %s", parts[2], name)
		}
	}

	connectorMap["properties"] = filtered
	return nil
}

func buildPatchRequestBody(requestBody []byte, format utils.Format, connectorId string) ([]byte, error) {

	connectorMap, err := utils.DeserializeToMap(requestBody, format, utils.GOVERNANCE_CONNECTORS)
	if err != nil {
		return nil, fmt.Errorf("error deserializing connector file: %w", err)
	}

	if connectorId == passwordExpiryConnectorId {
		if err := processPasswordExpiryConnector(connectorMap); err != nil {
			return nil, fmt.Errorf("error processing password expiry connector: %w", err)
		}
	}

	properties, ok := connectorMap["properties"]
	if !ok {
		return nil, fmt.Errorf("properties field not found in connector data")
	}
	formattedProperties, err := formatConnectorProperties(properties)
	if err != nil {
		return nil, fmt.Errorf("error extracting properties for connector: %w", err)
	}

	patchBody := map[string]interface{}{
		"operation":  "UPDATE",
		"properties": formattedProperties,
	}
	patchBodyBytes, err := utils.Serialize(patchBody, utils.FormatJSON, utils.GOVERNANCE_CONNECTORS)
	if err != nil {
		return nil, fmt.Errorf("error serializing to JSON: %w", err)
	}

	return patchBodyBytes, nil
}

func formatConnectorProperties(raw interface{}) ([]map[string]interface{}, error) {

	props, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("properties field is not an array")
	}

	var result []map[string]interface{}
	for _, item := range props {
		propMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("properties field is not in expected format")
		}
		result = append(result, map[string]interface{}{
			"name":  propMap["name"],
			"value": propMap["value"],
		})
	}
	return result, nil
}
