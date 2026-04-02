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

package emailTemplates

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type emailTemplateType struct {
	ID          string          `json:"id"`
	DisplayName string          `json:"displayName"`
	Templates   []emailTemplate `json:"templates,omitempty"`
}

type emailTemplate struct {
	ID string `json:"id"`
}

func getEmailTemplateTypeList() ([]emailTemplateType, error) {

	var list []emailTemplateType
	resp, err := utils.SendGetListRequest(utils.EMAIL_TEMPLATES, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving email template type list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved email template type list. %w", err)
		}
		if err = json.Unmarshal(body, &list); err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved email template type list. %w", err)
		}
		return list, nil
	} else if errMsg, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving email template type list. Status code: %d, Error: %s", statusCode, errMsg)
	}
	return nil, fmt.Errorf("error while retrieving email template type list")
}

func getDeployedEmailTemplateTypeNames() []string {

	types, err := getEmailTemplateTypeList()
	if err != nil {
		return []string{}
	}

	var typeNames []string
	for _, t := range types {
		typeNames = append(typeNames, t.DisplayName)
	}
	return typeNames
}

func getDeployedEmailTemplatesList(templateType emailTemplateType) []string {

	var templateIds []string
	for _, t := range templateType.Templates {
		templateIds = append(templateIds, t.ID)
	}
	return templateIds
}

func getEmailTemplateTypeDetails(typeId string) (*emailTemplateType, error) {

	jsonBytes, err := utils.SendGetRequest(utils.EMAIL_TEMPLATES, typeId)
	if err != nil {
		return nil, fmt.Errorf("error getting email template type details: %w", err)
	}

	var templateType emailTemplateType
	if err := json.Unmarshal(jsonBytes, &templateType); err != nil {
		return nil, fmt.Errorf("error deserializing email template type details: %w", err)
	}
	return &templateType, nil
}

func createEmailTemplateType(displayName string) (*emailTemplateType, error) {

	body := map[string]string{"displayName": displayName}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling create type request: %w", err)
	}

	resp, err := utils.SendPostRequest(utils.EMAIL_TEMPLATES, jsonBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading create type response: %w", err)
	}

	var created emailTemplateType
	if err := json.Unmarshal(respBody, &created); err != nil {
		return nil, fmt.Errorf("error parsing create type response: %w", err)
	}

	return &created, nil
}

func isEmailTemplateTypeExists(displayName string, types []emailTemplateType) *emailTemplateType {

	for i := range types {
		if types[i].DisplayName == displayName {
			return &types[i]
		}
	}
	return nil
}

func isTemplateExists(templateId string, deployedTemplates []emailTemplate) bool {

	for _, t := range deployedTemplates {
		if t.ID == templateId {
			return true
		}
	}
	return false
}

func getEmailTemplateKeywordMapping(typeName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.EmailTemplateConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(typeName, utils.KEYWORD_CONFIGS.EmailTemplateConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

