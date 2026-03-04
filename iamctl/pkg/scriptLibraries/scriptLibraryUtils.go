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

package scriptLibraries

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type scriptLibrary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type scriptLibraryListResponse struct {
	ScriptLibraries []scriptLibrary `json:"scriptLibraries"`
}

func getScriptLibraryList() ([]scriptLibrary, error) {

	resp, err := utils.SendGetListRequest(utils.SCRIPT_LIBRARIES, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving script library list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved script library list. %w", err)
		}

		var listResponse scriptLibraryListResponse
		err = json.Unmarshal(body, &listResponse)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved script library list. %w", err)
		}

		return listResponse.ScriptLibraries, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving script library list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("error while retrieving script library list")
}

func getDeployedScriptLibraryNames() []string {

	libraries, err := getScriptLibraryList()
	if err != nil {
		return []string{}
	}

	var names []string
	for _, library := range libraries {
		names = append(names, library.Name)
	}
	return names
}

func getScriptLibraryKeywordMapping(libraryName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ScriptLibraryConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(libraryName, utils.KEYWORD_CONFIGS.ScriptLibraryConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func isScriptLibraryExists(libraryName string, existingList []scriptLibrary) bool {

	for _, library := range existingList {
		if library.Name == libraryName {
			return true
		}
	}
	return false
}

func getScriptLibraryData(libraryName string) (map[string]interface{}, error) {

	libraryData, err := utils.GetResourceData(utils.SCRIPT_LIBRARIES, libraryName)
	if err != nil {
		return nil, fmt.Errorf("error while getting script library: %w", err)
	}

	contentBytes, err := utils.SendGetRequest(utils.SCRIPT_LIBRARIES, libraryName+"/content")
	if err != nil {
		return nil, fmt.Errorf("error retrieving script library content: %w", err)
	}

	dataMap, ok := libraryData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for script library response")
	}
	delete(dataMap, "content-ref")
	dataMap["content"] = string(contentBytes)
	return dataMap, nil
}

