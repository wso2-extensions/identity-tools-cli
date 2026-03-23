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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

const USERSTORE_SECRET_MASK = "ENCRYPTED PROPERTY"

type userStore struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type UserStoreConfigurations struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

func getUserStoreList() ([]userStore, error) {

	var list []userStore
	resp, err := utils.SendGetListRequest(utils.USERSTORES, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving user store list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved user store list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved user store list. %w", err)
		}
		resp.Body.Close()

		return list, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving user store list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("unexpected error while retrieving user store list")
}

func getDeployedUserstoreNames() []string {

	userstores, err := getUserStoreList()
	if err != nil {
		return []string{}
	}

	var userstoreNames []string
	for _, userstore := range userstores {
		userstoreNames = append(userstoreNames, userstore.Name)
	}
	return userstoreNames
}

func getUserStoreKeywordMapping(userStoreName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.UserStoreConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(userStoreName, utils.KEYWORD_CONFIGS.UserStoreConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getUserStoreId(userStoreName string) (string, error) {

	existingUserStoreList, err := getUserStoreList()
	if err != nil {
		return "", fmt.Errorf("error when retrieving the deployed user store list: %s", err)
	}

	for _, userstore := range existingUserStoreList {
		if userstore.Name == userStoreName {
			return userstore.Id, nil
		}
	}
	return "", nil
}
