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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.ROLES, "", "Importing roles...")
	importFilePath := filepath.Join(inputDirPath, utils.ROLES.String())
	setRolesV2ApiExists()

	if utils.ShouldSkip(utils.ROLES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.ROLES, "", "No roles to import.")
		return
	}

	existingRoleList, err := getRoleList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ROLES, "", fmt.Sprintf("Error retrieving the deployed role list: %s", err))
		utils.MarkResTypeFailure(utils.ROLES)
		return
	}

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ROLES, "", fmt.Sprintf("Error reading roles directory: %s", err))
		utils.MarkResTypeFailure(utils.ROLES)
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedRoles(files, existingRoleList)
	}

	for _, file := range files {
		roleFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(roleFilePath)
		displayName := unescapeName(fileInfo.ResourceName)

		if !utils.IsResourceExcluded(displayName, utils.TOOL_CONFIGS.RoleConfigs) {
			roleId := getRoleId(displayName, existingRoleList)
			err := importRole(displayName, roleId, roleFilePath)
			if err != nil {
				utils.PrintLog(utils.LogLevelError, utils.ROLES, displayName, fmt.Sprintf("Error importing role: %s", err))
				utils.UpdateFailureSummary(utils.ROLES, displayName)
			}
		}
	}
}

func importRole(displayName string, roleId string, importFilePath string) error {

	if displayName == utils.ADMIN_ROLE || (utils.RolesV2ApiExists && (displayName == utils.ADMINISTRATOR_ROLE || displayName == utils.IMPERSONATOR_ROLE)) {
		utils.PrintLog(utils.LogLevelInfo, utils.ROLES, displayName, "System role. Skipping import.")
		if roleId != "" {
			utils.AddToIdentifierMap(utils.ROLES, roleId, displayName, utils.IMPORT)
		}
		return nil
	}

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for role: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for role: %w", err)
	}

	roleKeywordMapping := getRoleKeywordMapping(displayName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), roleKeywordMapping)

	if roleId == "" {
		return createRole([]byte(modifiedFileData), format, displayName)
	}
	return updateRole(roleId, []byte(modifiedFileData), format, displayName)
}

func createRole(requestBody []byte, format utils.Format, displayName string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.ROLES, displayName, "Creating new role")

	roleMap, err := utils.DeserializeToMap(requestBody, format, utils.ROLES, "id")
	if err != nil {
		return fmt.Errorf("error deserializing role: %w", err)
	}

	var roleData interface{} = roleMap
	if utils.RolesV2ApiExists {
		roleData, err = processAudienceForImport(roleMap)
		if err != nil {
			return fmt.Errorf("error processing role audience: %w", err)
		}
	}

	jsonBody, err := utils.Serialize(roleData, utils.FormatJSON, utils.ROLES)
	if err != nil {
		return fmt.Errorf("error serializing to JSON: %w", err)
	}

	resp, err := utils.SendPostRequest(utils.ROLES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating role: %w", err)
	}
	defer resp.Body.Close()

	var created role
	if _, err := utils.ParseResponseBody(resp, &created); err != nil {
		return fmt.Errorf("error reading create role response: %w", err)
	}
	utils.AddToIdentifierMap(utils.ROLES, created.Id, created.DisplayName, utils.IMPORT)

	utils.UpdateSuccessSummary(utils.ROLES, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.ROLES, displayName, "Created successfully")
	return nil
}

func updateRole(roleId string, requestBody []byte, format utils.Format, displayName string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.ROLES, displayName, "Updating role")

	patchBody, err := buildRolePermissionsPatchBody(requestBody, format)
	if err != nil {
		return fmt.Errorf("error building patch body for role: %w", err)
	}

	resp, err := utils.SendPatchRequest(utils.ROLES, roleId, patchBody)
	if err != nil {
		return fmt.Errorf("error when updating role: %w", err)
	}
	defer resp.Body.Close()

	utils.AddToIdentifierMap(utils.ROLES, roleId, displayName, utils.IMPORT)

	utils.UpdateSuccessSummary(utils.ROLES, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.ROLES, displayName, "Updated successfully")
	return nil
}

func removeDeletedDeployedRoles(localFiles []os.FileInfo, deployedRoles []role) {

	if len(deployedRoles) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, r := range deployedRoles {
		fileName := escapeRoleName(r.DisplayName)
		if _, existsLocally := localResourceNames[fileName]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(r.DisplayName, utils.TOOL_CONFIGS.RoleConfigs) || r.DisplayName == utils.ADMIN_ROLE {
			utils.PrintLog(utils.LogLevelInfo, utils.ROLES, r.DisplayName, "Excluded from deletion.")
			continue
		}
		if utils.RolesV2ApiExists && (r.DisplayName == utils.ADMINISTRATOR_ROLE || r.DisplayName == utils.IMPERSONATOR_ROLE) {
			utils.PrintLog(utils.LogLevelInfo, utils.ROLES, r.DisplayName, "Excluded from deletion.")
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, utils.ROLES, r.DisplayName, "Not found locally. Deleting role.")
		if err := utils.SendDeleteRequest(r.Id, utils.ROLES); err != nil {
			utils.UpdateFailureSummary(utils.ROLES, r.DisplayName)
			utils.PrintLog(utils.LogLevelError, utils.ROLES, r.DisplayName, fmt.Sprintf("Error deleting role: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.ROLES, utils.DELETE)
		}
	}
}
