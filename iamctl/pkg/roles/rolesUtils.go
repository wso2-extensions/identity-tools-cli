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
	"io/ioutil"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/organizations"
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

	body, err := ioutil.ReadAll(resp.Body)
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

func getRoleId(displayName string, roleList []role) string {

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

func processExportedRole(roleData interface{}) (processedData interface{}, err error) {

	dataMap, ok := roleData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for response")
	}

	if dataMap, err = processPermissionsForExport(dataMap); err != nil {
		return nil, err
	}
	return processAudienceForExport(dataMap)
}

func processPermissionsForExport(roleMap map[string]interface{}) (map[string]interface{}, error) {

	permList, ok := roleMap["permissions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for permissions")
	}

	for i, perm := range permList {
		permMap, ok := perm.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for permission in list")
		}
		delete(permMap, "$ref")
		permList[i] = permMap
	}
	roleMap["permissions"] = permList
	return roleMap, nil
}

func processAudienceForExport(roleMap map[string]interface{}) (map[string]interface{}, error) {

	audience, ok := roleMap["audience"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for audience")
	}
	delete(audience, "value")
	roleMap["audience"] = audience
	return roleMap, nil
}

func processAudienceForImport(roleMap map[string]interface{}) (interface{}, error) {

	audience, ok := roleMap["audience"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for audience")
	}
	audType, ok := audience["type"].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected format for audience type")
	}

	switch audType {
	case "organization":
		superOrgId, err := organizations.GetCurrentOrganizationId()
		if err != nil {
			return nil, fmt.Errorf("error while retrieving organization ID: %w", err)
		}
		audience["value"] = superOrgId
		roleMap["audience"] = audience
		return roleMap, nil
	case "application":
		display, ok := audience["display"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected format for display key in audience")
		}
		audience["value"] = display
		roleMap["audience"] = audience
		return utils.ReplaceReferences(utils.ROLES, roleMap)
	default:
		return nil, fmt.Errorf("unsupported audience type")
	}
}

func escapeRoleName(filepath string) string {

	return strings.ReplaceAll(filepath, "Application/", "Application%2F")
}

func unescapeName(sanitizedFileName string) string {

	return strings.ReplaceAll(sanitizedFileName, "Application%2F", "Application/")
}

func setRolesV2ApiExists() {

	res, err := utils.CompareVersions(utils.SERVER_CONFIGS.ServerVersion, utils.MIN_VERSION_ROLES_V2_API)

	// Use the V2 API when the server version is not properly configured
	if err != nil || res >= 0 {
		utils.RolesV2ApiExists = true
	} else {
		utils.RolesV2ApiExists = false
	}
}
