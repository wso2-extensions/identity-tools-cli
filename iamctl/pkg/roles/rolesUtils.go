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

package roles

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type role struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type roleListResponse struct {
	Resources []role `json:"Resources"`
}

type patchOperation struct {
	Op    string                 `json:"op"`
	Value map[string]interface{} `json:"value"`
}

type rolePatchRequest struct {
	Operations []patchOperation `json:"Operations"`
	Schemas    []string         `json:"schemas"`
}

func getRoleList() ([]role, error) {

	resp, err := utils.SendGetListRequest(utils.ROLES, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving role list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if errMsg, ok := utils.ErrorCodes[resp.StatusCode]; ok {
			return nil, fmt.Errorf("error while retrieving roles list. Status: %d, Error: %s", resp.StatusCode, errMsg)
		}
		return nil, fmt.Errorf("error while retrieving roles list. Status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error when reading the retrieved roles list: %w", err)
	}

	var listResp roleListResponse
	err = json.Unmarshal(body, &listResp)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling the retrieved roles list: %w", err)
	}

	return listResp.Resources, nil
}

func getDeployedRoleLocalFileNames(roles []role) []string {

	var names []string
	for _, r := range roles {
		fileName := escapeRoleName(r.DisplayName)
		names = append(names, fileName)
	}
	return names
}

func getRoleKeywordMapping(roleName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.RoleConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(roleName, utils.KEYWORD_CONFIGS.RoleConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func isRoleExists(displayName string, existingRoleList []role) bool {

	for _, r := range existingRoleList {
		if r.DisplayName == displayName {
			return true
		}
	}
	return false
}

func getRoleIdByDisplayName(displayName string, roleList []role) string {

	for _, r := range roleList {
		if r.DisplayName == displayName {
			return r.Id
		}
	}
	return ""
}

func buildRolePermissionsPatchBody(fileBytes []byte, format utils.Format) ([]byte, error) {

	parsed, err := utils.Deserialize(fileBytes, format, utils.ROLES)
	if err != nil {
		return nil, fmt.Errorf("error deserializing role file: %w", err)
	}

	if interfaceMap, ok := parsed.(map[interface{}]interface{}); ok {
		parsed = utils.ConvertToStringKeyMap(interfaceMap)
	}
	dataMap, ok := parsed.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected role file format")
	}

	permissions := dataMap["permissions"]
	if permissions == nil {
		permissions = []interface{}{}
	}

	patchBody := rolePatchRequest{
		Operations: []patchOperation{
			{
				Op: "replace",
				Value: map[string]interface{}{
					"permissions": permissions,
				},
			},
		},
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
	}

	return utils.Serialize(patchBody, utils.FormatJSON, utils.ROLES)
}

func escapeRoleName(filepath string) string {

	return strings.ReplaceAll(filepath, "Application/", "Application%2F")
}

func unescapeName(sanitizedFileName string) string {

	return strings.ReplaceAll(sanitizedFileName, "Application%2F", "Application/")
}
