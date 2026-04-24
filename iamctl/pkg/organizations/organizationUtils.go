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

package organizations

import (
	"fmt"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type organization struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func GetCurrentOrganizationId() (id string, err error) {

	org, err := utils.SendGetRequest(utils.ORGANIZATIONS, "self")
	if err != nil {
		return "", fmt.Errorf("error while getting organization: %w", err)
	}

	var curOrg organization
	if _, err := utils.Deserialize(org, utils.FormatJSON, utils.ORGANIZATIONS, &curOrg); err != nil {
		return "", fmt.Errorf("error while deserializing JSON response: %w", err)
	}
	return curOrg.Id, nil
}
