/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

package claims

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

type claimDialect struct {
	Id         string `json:"id"`
	DialectURI string `json:"dialectURI"`
}

type ClaimDialectConfigurations struct {
	URI string `yaml:"uri"`
	ID  string `yaml:"id"`
}

func getClaimDialectsList() ([]claimDialect, error) {

	var list []claimDialect
	resp, err := utils.SendGetListRequest(utils.CLAIMS)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving claim dialect list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrived claim dialect list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrived claim dialect list. %w", err)
		}
		resp.Body.Close()

		return list, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving claim dialect list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("unexpected error while retrieving claim dialect list")

}

func getClaimKeywordMapping(claimDialectName string) map[string]interface{} {

	if utils.TOOL_CONFIGS.ClaimDialectConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(claimDialectName, utils.TOOL_CONFIGS.ClaimDialectConfigs)
	}
	return utils.TOOL_CONFIGS.KeywordMappings
}

func getDeployedClaimDialectNames() []string {

	claimdialects, err := getClaimDialectsList()
	if err != nil {
		return []string{}
	}

	var claimdialectNames []string
	for _, claimdialect := range claimdialects {
		claimdialectNames = append(claimdialectNames, claimdialect.DialectURI)
	}
	return claimdialectNames
}

func getClaimDialectId(claimDialectFilePath string) (string, error) {

	fileContent, err := ioutil.ReadFile(claimDialectFilePath)
	if err != nil {
		return "", fmt.Errorf("error when reading the file: %s. %s", claimDialectFilePath, err)
	}
	var claimDialectConfig ClaimDialectConfigurations
	err = yaml.Unmarshal(fileContent, &claimDialectConfig)
	if err != nil {
		return "", fmt.Errorf("invalid file content at: %s. %s", claimDialectFilePath, err)
	}

	existingClaimDialectList, err := getClaimDialectsList()
	if err != nil {
		return "", fmt.Errorf("error when retrieving the deployed claim dialect list: %s", err)
	}

	for _, dialect := range existingClaimDialectList {
		if dialect.Id == claimDialectConfig.ID {
			return dialect.Id, nil
		}
	}
	return "", nil
}
