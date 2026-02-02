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

package oidcScopes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v3"
)

type oidcScope struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Claims      []string `json:"claims"`
}

type oidcScopeConfig struct {
	Name string `yaml:"name"`
}

func getOidcScopeList() ([]oidcScope, error) {

	var list []oidcScope
	resp, err := utils.SendGetListRequest(utils.OIDC_SCOPES, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving OIDC scope list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved OIDC scope list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved OIDC scope list. %w", err)
		}
		resp.Body.Close()

		return list, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving OIDC scope list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("error while retrieving OIDC scope list")
}

func getDeployedOidcScopeNames() []string {

	scopes, err := getOidcScopeList()
	if err != nil {
		return []string{}
	}

	var scopeNames []string
	for _, scope := range scopes {
		scopeNames = append(scopeNames, scope.Name)
	}
	return scopeNames
}

func getOidcScopeKeywordMapping(scopeName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.OidcScopeConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(scopeName, utils.KEYWORD_CONFIGS.OidcScopeConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func isScopeExists(scopeFilePath string, scopeName string) (bool, error) {

	fileContent, err := ioutil.ReadFile(scopeFilePath)
	if err != nil {
		return false, fmt.Errorf("error when reading the file for OIDC scope: %s. %s", scopeName, err)
	}
	var scopeConfig oidcScopeConfig
	err = yaml.Unmarshal(fileContent, &scopeConfig)
	if err != nil {
		return false, fmt.Errorf("invalid file content for OIDC scope: %s. %s", scopeName, err)
	}
	existingScopeList, err := getOidcScopeList()
	if err != nil {
		return false, fmt.Errorf("error when retrieving the deployed OIDC scope list: %s", err)
	}

	for _, scope := range existingScopeList {
		if scope.Name == scopeConfig.Name {
			return true, nil
		}
	}
	return false, nil
}
