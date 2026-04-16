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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type apiScope struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiResource struct {
	ID         string        `json:"id"`
	Identifier string        `json:"identifier"`
	Scopes     []interface{} `json:"scopes"`
}

type apiResourceListResponse struct {
	APIResources []apiResource `json:"apiResources"`
}

var exportedScopesMap map[string]string

func getApiResourceList() ([]apiResource, error) {

	var listResponse apiResourceListResponse
	resp, err := utils.SendGetListRequest(utils.API_RESOURCES, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving API resource list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved API resource list. %w", err)
		}

		err = json.Unmarshal(body, &listResponse)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved API resource list. %w", err)
		}

		return listResponse.APIResources, nil
	} else if errMsg, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving API resource list. Status code: %d, Error: %s", statusCode, errMsg)
	}
	return nil, fmt.Errorf("error while retrieving API resource list")
}

func getDeployedApiResourceIdentifiers(resources []apiResource) []string {

	var identifiers []string
	for _, r := range resources {
		identifiers = append(identifiers, r.Identifier)
	}
	return identifiers
}

func getApiResourceId(identifier string, list []apiResource) string {

	for _, r := range list {
		if r.Identifier == identifier {
			return r.ID
		}
	}
	return ""
}

func getApiResourceKeywordMapping(resourceIdentifer string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ApiResourceConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(resourceIdentifer, utils.KEYWORD_CONFIGS.ApiResourceConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func processScopes(resourceMap map[string]interface{}, resourceIdentifier string) error {

	scopeList, ok := resourceMap["scopes"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for scopes in API resource data")
	}

	for _, scopeRaw := range scopeList {
		scopeEntry, ok := scopeRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for scope in API resource data")
		}
		delete(scopeEntry, "id")

		scopeName, ok := scopeEntry["name"].(string)
		if !ok {
			return fmt.Errorf("unexpected format for scope name")
		}
		exportedScopesMap[scopeName] = resourceIdentifier
	}

	return nil
}

func getApiResourceScopes(resourceId string) ([]apiScope, error) {

	var scopes []apiScope
	body, err := utils.SendGetRequest(utils.API_RESOURCES, resourceId+"/scopes")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &scopes); err != nil {
		return nil, fmt.Errorf("error unmarshalling scopes: %w", err)
	}
	return scopes, nil
}

func readLocalScopesMap(importDirPath string) (map[string]string, error) {

	matches, err := filepath.Glob(filepath.Join(importDirPath, "ApiResourceScopes.*"))
	if err != nil {
		return nil, fmt.Errorf("error searching for file: %w", err)
	}
	if len(matches) == 0 {
		return map[string]string{}, nil
	}

	fileBytes, err := ioutil.ReadFile(matches[0])
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	format, err := utils.FormatFromExtension(filepath.Ext(matches[0]))
	if err != nil {
		return nil, fmt.Errorf("unsupported format for file: %w", err)
	}
	var scopeMap map[string]string
	if _, err := utils.Deserialize(fileBytes, format, utils.API_RESOURCES, &scopeMap); err != nil {
		return nil, fmt.Errorf("error deserializing file: %w", err)
	}
	return scopeMap, nil
}

func writeScopesMap(outputDirPath string, scopeMap map[string]string, formatString string) error {

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, utils.API_RESOURCE_SCOPES.String(), format)

	data, err := utils.Serialize(scopeMap, format, utils.API_RESOURCES)
	if err != nil {
		return fmt.Errorf("error serializing scope name map: %w", err)
	}
	if err := ioutil.WriteFile(exportedFileName, data, 0644); err != nil {
		return fmt.Errorf("error writing scope name map: %w", err)
	}
	return nil
}

func updateApiResourceExportSummary(success bool, successCount int) {

	if !success {
		utils.UpdateFailureSummary(utils.API_RESOURCES, utils.API_RESOURCE_SCOPES.String())
		return
	}
	for i := 0; i < successCount; i++ {
		utils.UpdateSuccessSummary(utils.API_RESOURCES, utils.EXPORT)
	}
}
