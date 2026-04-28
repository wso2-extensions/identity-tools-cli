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
	"log"

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
		return "email providers"
	case utils.SMS_PROVIDERS:
		return "sms providers"
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

func processAuthSecrets(resType utils.ResourceType, data interface{}) (interface{}, error) {

	providerMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for provider data")
	}
	provider, ok := providerMap["provider"].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected format for provider")
	}

	var authType string
	var authentication map[string]interface{}
	var properties interface{}

	switch resType {
	case utils.EMAIL_PROVIDERS:
		if !utils.AreSecretsExcluded(utils.TOOL_CONFIGS.EmailProviderConfigs) {
			log.Printf("Warn: Secrets exclusion cannot be disabled for Email Providers. All secrets will be masked.")
		}

		authType, ok = providerMap["authType"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected format for authType")
		}
		emailProps, ok := providerMap["properties"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for properties")
		}
		properties = emailProps

	case utils.SMS_PROVIDERS:
		if provider != "Custom" {
			delete(providerMap, "authentication")
			if utils.AreSecretsExcluded(utils.TOOL_CONFIGS.SmsProviderConfigs) {
				providerMap["secret"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
			}
			return providerMap, nil
		}

		if !utils.AreSecretsExcluded(utils.TOOL_CONFIGS.SmsProviderConfigs) {
			log.Printf("Warn: Secrets exclusion cannot be disabled for Custom SMS Providers. All secrets will be masked.")
		}
		delete(providerMap, "secret")

		authentication, ok = providerMap["authentication"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for authentication")
		}
		authType, ok = authentication["type"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected format for authentication type")
		}
		smsProps, ok := authentication["properties"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for authentication properties")
		}
		properties = smsProps
	}

	maskedProps, err := maskAuthSecrets(resType, authType, properties)
	if err != nil {
		return nil, fmt.Errorf("error masking auth secrets: %w", err)
	}

	switch resType {
	case utils.EMAIL_PROVIDERS:
		providerMap["properties"] = maskedProps
	case utils.SMS_PROVIDERS:
		authentication["properties"] = maskedProps
	}

	return providerMap, nil
}

func maskAuthSecrets(resType utils.ResourceType, authType string, properties interface{}) (interface{}, error) {

	switch authType {
	case "NONE":
		return properties, nil
	case "BASIC":
		return addSecretWithMask(resType, properties, "password")
	case "CLIENT_CREDENTIAL":
		return addSecretWithMask(resType, properties, "clientSecret")
	case "BEARER":
		return addSecretWithMask(resType, properties, "accessToken")
	case "API_KEY":
		switch resType {
		case utils.SMS_PROVIDERS:
			return addSecretWithMask(resType, properties, "value")
		case utils.EMAIL_PROVIDERS:
			return addSecretWithMask(resType, properties, "apiKeyValue")
		}
	}
	return properties, nil
}

func addSecretWithMask(resType utils.ResourceType, properties interface{}, key string) (interface{}, error) {

	switch resType {
	case utils.EMAIL_PROVIDERS:
		emailProps, ok := properties.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for email provider properties")
		}
		return append(emailProps, map[string]interface{}{
			"key":   key,
			"value": utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES,
		}), nil

	case utils.SMS_PROVIDERS:
		smsProps, ok := properties.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for SMS provider properties")
		}
		smsProps[key] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		return smsProps, nil
	}
	return properties, nil
}
