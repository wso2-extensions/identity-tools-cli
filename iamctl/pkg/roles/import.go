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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing roles...")
	importFilePath := filepath.Join(inputDirPath, utils.ROLES.String())

	if utils.IsResourceTypeExcluded(utils.ROLES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No roles to import.")
		return
	}

	existingRoleList, err := getRoleList()
	if err != nil {
		log.Println("Error retrieving the deployed role list:", err)
		return
	}

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing roles:", err)
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
				log.Println("Error importing role:", err)
				utils.UpdateFailureSummary(utils.ROLES, displayName)
			}
		}
	}
}

func importRole(displayName string, roleId string, importFilePath string) error {

	if displayName == utils.ADMIN {
		log.Println("Role: admin is a system role. Skipping import.")
		utils.AddToIdentifierMap(utils.ROLES, roleId, displayName, utils.IMPORT)
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

	log.Println("Creating new role:", displayName)

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.ROLES)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.ROLES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating role: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading role response: %w", err)
	}
	var created role
	if err := json.Unmarshal(respBody, &created); err != nil {
		return fmt.Errorf("error parsing create role response: %w", err)
	}
	utils.AddToIdentifierMap(utils.ROLES, created.Id, created.DisplayName, utils.IMPORT)

	utils.UpdateSuccessSummary(utils.ROLES, utils.IMPORT)
	log.Println("Role created successfully:", displayName)
	return nil
}

func updateRole(roleId string, requestBody []byte, format utils.Format, displayName string) error {

	log.Println("Updating role:", displayName)

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
	log.Println("Role updated successfully:", displayName)
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
		if utils.IsResourceExcluded(r.DisplayName, utils.TOOL_CONFIGS.RoleConfigs) || r.DisplayName == utils.ADMIN {
			log.Println("Role is excluded from deletion:", r.DisplayName)
			continue
		}

		log.Printf("Role: %s not found locally. Deleting role.\n", r.DisplayName)
		if err := utils.SendDeleteRequest(r.Id, utils.ROLES); err != nil {
			utils.UpdateFailureSummary(utils.ROLES, r.DisplayName)
			log.Println("Error deleting role:", r.DisplayName, err)
		} else {
			utils.UpdateSuccessSummary(utils.ROLES, utils.DELETE)
		}
	}
}
