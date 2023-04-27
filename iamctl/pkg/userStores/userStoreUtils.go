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

package userstores

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type userStore struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// type userStoreConfig struct {
// 	userStoreName string `yaml:"name"`
// 	userStoreId   string
// }

func getUserStoreList() ([]userStore, error) {

	var reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/userstores"
	var list []userStore

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, _ := http.NewRequest("GET", reqUrl, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	req.Header.Set("accept", "*/*")
	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve available userstore list. %w", err)
	}

	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrived userstore list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrived userstore list. %w", err)
		}
		resp.Body.Close()

		return list, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving userstore list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("unexpected error while retrieving userstore list")
}

func getUserStoreKeywordMapping(userStoreName string) map[string]interface{} {

	if utils.TOOL_CONFIGS.UserStoreConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(userStoreName, utils.TOOL_CONFIGS.UserStoreConfigs)
	}
	return utils.TOOL_CONFIGS.KeywordMappings
}
