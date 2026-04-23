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

package authorizedApis

import (
	"encoding/json"
	"fmt"

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

var supportedInVersion bool

func InitSupportedInVersion() {

	supportedInVersion = utils.IsEntitySupportedInVersion(utils.APPLICATION_AUTHORIZED_APIS)
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
