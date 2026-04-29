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

package applicationNotificationTemplates

import (
	"encoding/json"
	"fmt"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

const ApplicationTemplatesDir = "applicationTemplates"

type appTemplate struct {
	Locale string `json:"locale"`
}

func getAppTemplatesList(rt utils.ResourceType, typeId, appId string) ([]appTemplate, error) {

	var list []appTemplate
	body, err := utils.SendGetRequest(rt, typeId+"/app-templates/"+appId)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("error unmarshalling app templates: %w", err)
	}
	return list, nil
}

func getDeployedAppTemplateLocales(templates []appTemplate) []string {

	var locales []string
	for _, t := range templates {
		locales = append(locales, t.Locale)
	}
	return locales
}

func isAppTemplateExists(locale string, templates []appTemplate) bool {

	for _, t := range templates {
		if t.Locale == locale {
			return true
		}
	}
	return false
}
