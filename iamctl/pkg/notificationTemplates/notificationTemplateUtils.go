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

package notificationTemplates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type notificationTemplateType struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type notificationTemplate struct {
	Locale string `json:"locale"`
}

func getTemplateResourceConfig(rt utils.ResourceType) map[string]interface{} {

	switch rt {
	case utils.EMAIL_TEMPLATES:
		return utils.TOOL_CONFIGS.EmailTemplateConfigs
	case utils.SMS_TEMPLATES:
		return utils.TOOL_CONFIGS.SmsTemplateConfigs
	}
	return nil
}

func getTemplateKeywordConfig(rt utils.ResourceType) map[string]interface{} {

	switch rt {
	case utils.EMAIL_TEMPLATES:
		return utils.KEYWORD_CONFIGS.EmailTemplateConfigs
	case utils.SMS_TEMPLATES:
		return utils.KEYWORD_CONFIGS.SmsTemplateConfigs
	}
	return nil
}

func getTemplateLogName(rt utils.ResourceType) string {

	switch rt {
	case utils.EMAIL_TEMPLATES:
		return "Email Templates"
	case utils.SMS_TEMPLATES:
		return "SMS Templates"
	}
	return string(rt)
}

func getTemplateTypeList(rt utils.ResourceType) ([]notificationTemplateType, error) {

	var list []notificationTemplateType
	resp, err := utils.SendGetListRequest(rt, -1)
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

func getTemplatesList(rt utils.ResourceType, typeId string) ([]notificationTemplate, error) {

	var list []notificationTemplate
	body, err := utils.SendGetRequest(rt, typeId+"/org-templates")
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("error unmarshalling templates: %w", err)
	}
	return list, nil
}

func getDeployedTemplateLocales(templates []notificationTemplate) []string {

	var locales []string
	for _, t := range templates {
		locales = append(locales, t.Locale)
	}
	return locales
}

func getTemplateKeywordMapping(rt utils.ResourceType, typeName string) map[string]interface{} {

	kwConfig := getTemplateKeywordConfig(rt)
	if kwConfig != nil {
		return utils.ResolveAdvancedKeywordMapping(typeName, kwConfig)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}
