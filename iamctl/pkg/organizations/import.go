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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing organizations...")
	importFilePath := filepath.Join(inputDirPath, utils.ORGANIZATIONS.String())

	if utils.IsResourceTypeExcluded(utils.ORGANIZATIONS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No organizations to import.")
		return
	}

	existingList, err := getOrganizationList()
	if err != nil {
		log.Println("Error retrieving the deployed organization list: ", err)
		return
	}
	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error reading local organization  files: ", err)
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedOrganizations(files, existingList)
	}

	for _, file := range files {
		orgFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(orgFilePath)
		orgHandle := fileInfo.ResourceName

		if !utils.IsResourceExcluded(orgHandle, utils.TOOL_CONFIGS.OrganizationConfigs) {
			orgId := getOrgId(orgHandle, existingList)
			err := importOrganization(orgHandle, orgId, orgFilePath)
			if err != nil {
				log.Println("Error importing organization: ", err)
				utils.UpdateFailureSummary(utils.ORGANIZATIONS, orgHandle)
			}
		}
	}
}

func importOrganization(orgHandle, orgId, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for organization: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for organization: %w", err)
	}

	orgKeywordMapping := getOrganizationKeywordMapping(orgHandle)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), orgKeywordMapping)

	if orgId == "" {
		return createOrganization([]byte(modifiedFileData), format, orgHandle)
	}
	return updateOrganization(orgId, []byte(modifiedFileData), format, orgHandle)
}

func createOrganization(requestBody []byte, format utils.Format, orgHandle string) error {

	log.Println("Creating new organization: " + orgHandle)

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.ORGANIZATIONS,
		"id", "parent", "version", "status", "permissions", "created", "lastModified", "hasChildren", "ancestorPath")
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.ORGANIZATIONS, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating organization: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.ORGANIZATIONS, utils.IMPORT)
	log.Println("Organization created successfully.")
	return nil
}

func updateOrganization(orgId string, requestBody []byte, format utils.Format, orgHandle string) error {

	log.Println("Updating organization: " + orgHandle)

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
	log.Println("Organization updated successfully.")
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
		if _, existsLocally := localResourceNames[org.OrgHandle]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(org.OrgHandle, utils.TOOL_CONFIGS.OrganizationConfigs) {
			log.Println("Organization is excluded from deletion:", org.OrgHandle)
			continue
		}

		log.Printf("Organization: %s not found locally. Deleting organization.\n", org.OrgHandle)
		if err := utils.SendDeleteRequest(org.Id, utils.ORGANIZATIONS); err != nil {
			utils.UpdateFailureSummary(utils.ORGANIZATIONS, org.OrgHandle)
			log.Println("Error deleting organization:", org.OrgHandle, err)
		} else {
			utils.UpdateSuccessSummary(utils.ORGANIZATIONS, utils.DELETE)
		}
	}
}
