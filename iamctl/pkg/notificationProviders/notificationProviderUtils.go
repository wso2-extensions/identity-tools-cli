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

package notificationProviders

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type notificationProvider struct {
	Name string `json:"name"`
}

func getProviderResourceConfig(resType utils.ResourceType) map[string]interface{} {

	switch resType {
	case utils.EMAIL_PROVIDERS:
		return utils.TOOL_CONFIGS.EmailProviderConfigs
	case utils.SMS_PROVIDERS:
		return utils.TOOL_CONFIGS.SmsProviderConfigs
	}
	return nil
}

func getProviderKeywordConfig(resType utils.ResourceType) map[string]interface{} {

	switch resType {
	case utils.EMAIL_PROVIDERS:
		return utils.KEYWORD_CONFIGS.EmailProviderConfigs
	case utils.SMS_PROVIDERS:
		return utils.KEYWORD_CONFIGS.SmsProviderConfigs
	}
	return nil
}

func getProviderLogName(resType utils.ResourceType) string {

	switch resType {
	case utils.EMAIL_PROVIDERS:
		return "Email Providers"
	case utils.SMS_PROVIDERS:
		return "SMS Providers"
	}
	return string(resType)
}

func getProviderList(resType utils.ResourceType) ([]notificationProvider, error) {

	var list []notificationProvider
	resp, err := utils.SendGetListRequest(resType, -1)
	if err != nil {
		return nil, fmt.Errorf("error while getting the list: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the list: %w", err)
		}
		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the list: %w", err)
		}
		return list, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("unknown error while getting the list")
}

func getDeployedProviderNames(providers []notificationProvider) []string {

	var names []string
	for _, provider := range providers {
		names = append(names, provider.Name)
	}
	return names
}

func getProviderKeywordMapping(resType utils.ResourceType, name string) map[string]interface{} {

	kwConfig := getProviderKeywordConfig(resType)
	if kwConfig != nil {
		return utils.ResolveAdvancedKeywordMapping(name, kwConfig)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func isProviderExists(name string, exisitingProviderList []notificationProvider) bool {

	for _, provider := range exisitingProviderList {
		if provider.Name == name {
			return true
		}
	}
	return false
}
