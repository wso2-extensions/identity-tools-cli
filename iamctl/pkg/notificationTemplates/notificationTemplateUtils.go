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
	"net/http"
	"path/filepath"

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
		return "email templates"
	case utils.SMS_TEMPLATES:
		return "sms templates"
	}
	return string(rt)
}

func getTemplateChannel(rt utils.ResourceType) string {

	switch rt {
	case utils.EMAIL_TEMPLATES:
		return "EMAIL"
	case utils.SMS_TEMPLATES:
		return "SMS"
	}
	return ""
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

func getTemplateTypeId(displayName string, types []notificationTemplateType) string {

	for i := range types {
		if types[i].DisplayName == displayName {
			return types[i].ID
		}
	}
	return ""
}

func isTemplateExists(locale string, templates []notificationTemplate) bool {

	for _, t := range templates {
		if t.Locale == locale {
			return true
		}
	}
	return false
}

func createTemplateType(rt utils.ResourceType, displayName string) (string, error) {

	body := map[string]string{
		"displayName": displayName,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := utils.SendPostRequest(rt, jsonBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}
	var created notificationTemplateType
	if err := json.Unmarshal(respBody, &created); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	return created.ID, nil
}

func resetTemplateType(rt utils.ResourceType, typeId string) error {

	reqURL := utils.GetTenantBaseUrl() + "/api/server/v1/notification/reset-template-type"

	body := map[string]string{
		"templateTypeId": typeId,
		"channel":        getTemplateChannel(rt),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := utils.SendCustomRequest("POST", reqURL, jsonBody, utils.MEDIA_TYPE_JSON)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		if errMsg, ok := utils.ErrorCodes[resp.StatusCode]; ok {
			return fmt.Errorf("error response for reset request: %s", errMsg)
		}
		return fmt.Errorf("unexpected error when resetting: %s", resp.Status)
	}
	return nil
}

func writeTemplateTypesList(outputDirPath string, typeNames []string, rt utils.ResourceType, format utils.Format) error {

	exportedFileName := utils.GetExportedFilePath(outputDirPath, "TemplateTypes", format)
	data, err := utils.Serialize(typeNames, format, rt)
	if err != nil {
		return fmt.Errorf("error serializing list: %w", err)
	}
	if err := ioutil.WriteFile(exportedFileName, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}

func readLocalTemplateTypeNames(importDirPath string, rt utils.ResourceType) ([]string, error) {

	matches, err := filepath.Glob(filepath.Join(importDirPath, "TemplateTypes.*"))
	if err != nil {
		return nil, fmt.Errorf("error searching for file: %w", err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("file not found")
	}

	fileBytes, err := ioutil.ReadFile(matches[0])
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	format, err := utils.FormatFromExtension(filepath.Ext(matches[0]))
	if err != nil {
		return nil, fmt.Errorf("unsupported format for file: %w", err)
	}

	var names []string
	if _, err := utils.Deserialize(fileBytes, format, rt, &names); err != nil {
		return nil, fmt.Errorf("error deserializing file: %w", err)
	}
	return names, nil
}
