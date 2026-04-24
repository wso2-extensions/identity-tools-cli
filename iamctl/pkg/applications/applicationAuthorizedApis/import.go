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
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAPIs(appId, appName, appsImportDirPath string) error {

	if !SupportedInVersion {
		return nil
	}

	filePath, err := findLocalFile(appsImportDirPath, appName)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading authorized APIs file: %w", err)
	}

	keywordMapping := getAuthorizedApisKeywordMapping(appName)
	fileContent := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for authorized APIs file: %w", err)
	}

	authApis, err := utils.Deserialize([]byte(fileContent), format, utils.APPLICATION_AUTHORIZED_APIS)
	if err != nil {
		return fmt.Errorf("error deserializing authorized APIs: %w", err)
	}
	apiList, ok := authApis.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for authorized APIs file content")
	}

	deployedAPIs, err := getAuthorizedAPIList(appId)
	if err != nil {
		return fmt.Errorf("error fetching deployed authorized APIs: %w", err)
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		if err := removeDeletedAuthorizedAPIs(appId, apiList, deployedAPIs); err != nil {
			return fmt.Errorf("error removing deleted APIs: %w", err)
		}
	}

	for _, api := range apiList {
		if err := importAuthorizedAPI(appId, api, deployedAPIs); err != nil {
			return fmt.Errorf("error importing API: %w", err)
		}
	}
	return nil
}

func importAuthorizedAPI(appId string, api interface{}, deployedAPIs []AuthorizedAPI) error {

	apiMap, ok := api.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for API")
	}
	identifier, ok := apiMap["identifier"].(string)
	if !ok {
		return fmt.Errorf("unexpected format for API identifier")
	}
	deployedApi := getAPIByIdentifier(identifier, deployedAPIs)

	if deployedApi != nil {
		err := updateAuthorizedAPI(appId, apiMap, *deployedApi)
		if err != nil {
			return fmt.Errorf("error updating API %q: %w", identifier, err)
		}
	} else {
		err := createAuthorizedAPI(appId, identifier, apiMap)
		if err != nil {
			return fmt.Errorf("error creating API %q: %w", identifier, err)
		}
	}
	return nil
}

func createAuthorizedAPI(appId, identifier string, apiMap map[string]interface{}) error {

	apiId, err := getApiIdByIdentifier(identifier)
	if err != nil {
		return err
	}
	scopeNames, err := extractScopeNames(apiMap)
	if err != nil {
		return err
	}

	body, err := buildPostRequestBody(apiMap, apiId, scopeNames)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.APPLICATIONS, body,
		utils.WithPathSuffix(appId+"/authorized-apis"))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func updateAuthorizedAPI(appId string, apiMap map[string]interface{}, deployedApi AuthorizedAPI) error {

	localScopeNames, err := extractScopeNames(apiMap)
	if err != nil {
		return err
	}

	localScopes := make(map[string]struct{}, len(localScopeNames))
	for _, name := range localScopeNames {
		localScopes[name] = struct{}{}
	}
	deployedScopes := make(map[string]struct{}, len(deployedApi.AuthorizedScopes))
	for _, s := range deployedApi.AuthorizedScopes {
		deployedScopes[s.Name] = struct{}{}
	}

	reqbody, err := buildPatchRequestBody(localScopes, deployedScopes)
	if err != nil {
		return err
	}
	if reqbody == nil {
		return nil
	}

	resp, err := utils.SendPatchRequest(utils.APPLICATIONS, appId+"/authorized-apis/"+deployedApi.ID, reqbody)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func removeDeletedAuthorizedAPIs(appId string, localApis []interface{}, deployedApis []AuthorizedAPI) error {

	localIdents := make(map[string]struct{})
	for _, api := range localApis {
		apiMap, ok := api.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for API")
		}
		identifier, ok := apiMap["identifier"].(string)
		if !ok {
			return fmt.Errorf("unexpected format for API identifier")
		}
		localIdents[identifier] = struct{}{}
	}

	for _, dep := range deployedApis {
		if _, exists := localIdents[dep.Identifier]; exists {
			continue
		}
		if err := utils.SendDeleteRequest(appId+"/authorized-apis/"+dep.ID, utils.APPLICATIONS); err != nil {
			return fmt.Errorf("error deleting API %q: %w", dep.Identifier, err)
		}
	}
	return nil
}
