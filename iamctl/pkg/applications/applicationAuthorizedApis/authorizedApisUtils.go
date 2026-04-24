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

package applicationAuthorizedApis

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/apiResources"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type AuthorizedAPI struct {
	ID               string            `json:"id"`
	Identifier       string            `json:"identifier"`
	Type             string            `json:"type"`
	AuthorizedScopes []AuthorizedScope `json:"authorizedScopes"`
}

type AuthorizedScope struct {
	Name string `json:"name"`
}

var SupportedInVersion bool
var apiResourcesMap map[string]string

func InitSupportedInVersion() {

	SupportedInVersion = utils.IsEntitySupportedInVersion(utils.APPLICATION_AUTHORIZED_APIS)
}

func GetAPIResources() error {

	list, err := apiResources.GetApiResourceList(false)
	if err != nil {
		return fmt.Errorf("error while retrieving API resource list: %w", err)
	}
	apiResourcesMap = make(map[string]string)
	for _, r := range list {
		apiResourcesMap[r.Identifier] = r.ID
	}
	return nil
}

func GetOutputDirPath(appsOutputDirPath string) string {

	return filepath.Join(appsOutputDirPath, utils.APPLICATION_AUTHORIZED_APIS.String())
}

func getAuthorizedAPIList(appId string) ([]AuthorizedAPI, error) {

	body, err := utils.SendGetRequest(utils.APPLICATIONS, appId+"/authorized-apis")
	if err != nil {
		return nil, err
	}
	var apis []AuthorizedAPI
	if err := json.Unmarshal(body, &apis); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}
	return apis, nil
}

func getAuthorizedApisKeywordMapping(appName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ApplicationConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(appName, utils.KEYWORD_CONFIGS.ApplicationConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getAPIByIdentifier(identifier string, deployedApis []AuthorizedAPI) *AuthorizedAPI {

	for i := range deployedApis {
		if deployedApis[i].Identifier == identifier {
			return &deployedApis[i]
		}
	}
	return nil
}

func findLocalFile(appsImportDirPath, appName string) (string, error) {

	matches, err := filepath.Glob(filepath.Join(GetOutputDirPath(appsImportDirPath), appName+".*"))
	if err != nil {
		return "", fmt.Errorf("error searching for authorized APIs file for app: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no authorized APIs file found for application")
	}
	return matches[0], nil
}

func extractScopeNames(apiMap map[string]interface{}) ([]string, error) {

	scopes, ok := apiMap["authorizedScopes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for authorizedScopes")
	}
	names := make([]string, 0, len(scopes))

	for _, s := range scopes {
		scopeMap, ok := s.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for scope")
		}
		name, ok := scopeMap["name"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected format for scope name")
		}
		names = append(names, name)
	}
	return names, nil
}

func getApiIdByIdentifier(identifier string) (string, error) {

	id, ok := apiResourcesMap[identifier]
	if !ok {
		return "", fmt.Errorf("referenced API resource with identifier '%s' is not found in the target environment", identifier)
	}
	return id, nil
}

func buildPostRequestBody(apiMap map[string]interface{}, apiId string, scopeNames []string) ([]byte, error) {

	delete(apiMap, "displayName")
	delete(apiMap, "type")

	policyIdentifier, ok := apiMap["policyId"].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected format for policy identifier")
	}
	apiMap["policyIdentifier"] = policyIdentifier
	delete(apiMap, "policyId")

	delete(apiMap, "identifier")
	apiMap["id"] = apiId

	delete(apiMap, "authorizedScopes")
	apiMap["scopes"] = scopeNames

	body, err := json.Marshal(apiMap)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}
	return body, nil
}

func buildPatchRequestBody(localScopes, deployedScopes map[string]struct{}) ([]byte, error) {

	addedScopes := make([]string, 0)
	removedScopes := make([]string, 0)
	for name := range localScopes {
		if _, exists := deployedScopes[name]; !exists {
			addedScopes = append(addedScopes, name)
		}
	}
	for name := range deployedScopes {
		if _, exists := localScopes[name]; !exists {
			removedScopes = append(removedScopes, name)
		}
	}

	if len(addedScopes) == 0 && len(removedScopes) == 0 {
		return nil, nil
	}
	patch := struct {
		AddedScopes   []string `json:"addedScopes"`
		RemovedScopes []string `json:"removedScopes"`
	}{
		AddedScopes:   addedScopes,
		RemovedScopes: removedScopes,
	}

	body, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	return body, nil
}
