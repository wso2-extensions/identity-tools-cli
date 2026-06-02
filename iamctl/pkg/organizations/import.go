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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, "", "Importing organizations...")
	importFilePath := filepath.Join(inputDirPath, utils.ORGANIZATIONS.String())

	if utils.ShouldSkip(utils.ORGANIZATIONS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, "", "No organizations to import.")
		return
	}

	existingList, err := getOrganizationList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ORGANIZATIONS, "", fmt.Sprintf("Error retrieving the deployed organization list: %s", err))
		utils.MarkResTypeFailure(utils.ORGANIZATIONS)
		return
	}
	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ORGANIZATIONS, "", fmt.Sprintf("Error reading local organization files: %s", err))
		utils.MarkResTypeFailure(utils.ORGANIZATIONS)
		return
	}
	curOrgId, err = GetCurrentOrganizationId()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ORGANIZATIONS, "", "Error while retrieving current organization ID")
		utils.MarkResTypeFailure(utils.ORGANIZATIONS)
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedOrganizations(files, existingList)
	}

	for _, file := range files {
		orgFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(orgFilePath)
		resourceName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(resourceName, utils.TOOL_CONFIGS.OrganizationConfigs) {
			orgId := getOrgId(resourceName, existingList)
			err := importOrganization(resourceName, orgId, orgFilePath)
			if err != nil {
				utils.PrintLog(utils.LogLevelError, utils.ORGANIZATIONS, resourceName, fmt.Sprintf("Error importing organization: %s", err))
				utils.UpdateFailureSummary(utils.ORGANIZATIONS, resourceName)
			}
		}
	}
}

func importOrganization(resourceName, orgId, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for organization: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for organization: %w", err)
	}

	orgKeywordMapping := getOrganizationKeywordMapping(resourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), orgKeywordMapping)

	if orgId == "" {
		return createOrganization([]byte(modifiedFileData), format, resourceName)
	}
	return updateOrganization(orgId, []byte(modifiedFileData), format, resourceName)
}

func createOrganization(requestBody []byte, format utils.Format, resourceName string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, resourceName, "Creating new organization")

	jsonBody, status, err := prepareOrganizationPostBody(requestBody, format, curOrgId)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.ORGANIZATIONS, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating organization: %w", err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error reading create response: %w", err)
	}
	var createdOrg organization
	if err := json.Unmarshal(respBody, &createdOrg); err != nil {
		return fmt.Errorf("error parsing create response: %w", err)
	}

	if err := patchOrganizationStatus(createdOrg.Id, status); err != nil {
		return fmt.Errorf("error updating status field: %w", err)
	}

	utils.UpdateSuccessSummary(utils.ORGANIZATIONS, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, resourceName, "Created successfully")
	return nil
}

func updateOrganization(orgId string, requestBody []byte, format utils.Format, resourceName string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, resourceName, "Updating organization")

	updateBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.ORGANIZATIONS,
		"id", "orgHandle", "type", "parent", "permissions", "created", "lastModified", "hasChildren", "ancestorPath")
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(utils.ORGANIZATIONS, orgId, updateBody)
	if err != nil {
		return fmt.Errorf("error when updating organization: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.ORGANIZATIONS, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, resourceName, "Updated successfully")
	return nil
}

func removeDeletedDeployedOrganizations(localFiles []os.FileInfo, deployedOrgs []organization) {

	if len(deployedOrgs) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, org := range deployedOrgs {
		resourceName := getOrgResourceName(org)
		if _, existsLocally := localResourceNames[resourceName]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(resourceName, utils.TOOL_CONFIGS.OrganizationConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, resourceName, "Excluded from deletion")
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, utils.ORGANIZATIONS, resourceName, "Not found locally. Deleting organization.")
		if err := utils.SendDeleteRequest(org.Id, utils.ORGANIZATIONS); err != nil {
			utils.UpdateFailureSummary(utils.ORGANIZATIONS, resourceName)
			utils.PrintLog(utils.LogLevelError, utils.ORGANIZATIONS, resourceName, fmt.Sprintf("Error deleting organization: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.ORGANIZATIONS, utils.DELETE)
		}
	}
}
