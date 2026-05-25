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
		log.Println("Error reading local organization files: ", err)
		return
	}
	curOrgId, err = GetCurrentOrganizationId()
	if err != nil {
		log.Println("error while retrieving current organization ID")
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedOrganizations(files, existingList)
	}

	for _, file := range files {
		orgFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(orgFilePath)
		orgName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(orgName, utils.TOOL_CONFIGS.OrganizationConfigs) {
			orgId := getOrgId(orgName, existingList)
			err := importOrganization(orgName, orgId, orgFilePath)
			if err != nil {
				log.Println("Error importing organization: ", err)
				utils.UpdateFailureSummary(utils.ORGANIZATIONS, orgName)
			}
		}
	}
}

func importOrganization(orgName, orgId, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for organization: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for organization: %w", err)
	}

	orgKeywordMapping := getOrganizationKeywordMapping(orgName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), orgKeywordMapping)

	if orgId == "" {
		return createOrganization([]byte(modifiedFileData), format, orgName)
	}
	return updateOrganization(orgId, []byte(modifiedFileData), format, orgName)
}

func createOrganization(requestBody []byte, format utils.Format, orgName string) error {

	log.Println("Creating new organization: " + orgName)

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
	log.Println("Organization created successfully.")
	return nil
}

func updateOrganization(orgId string, requestBody []byte, format utils.Format, orgName string) error {

	log.Println("Updating organization: " + orgName)

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
		if _, existsLocally := localResourceNames[org.Name]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(org.Name, utils.TOOL_CONFIGS.OrganizationConfigs) {
			log.Println("Organization is excluded from deletion:", org.Name)
			continue
		}

		log.Printf("Organization: %s not found locally. Deleting organization.\n", org.Name)
		if err := utils.SendDeleteRequest(org.Id, utils.ORGANIZATIONS); err != nil {
			utils.UpdateFailureSummary(utils.ORGANIZATIONS, org.Name)
			log.Println("Error deleting organization:", org.Name, err)
		} else {
			utils.UpdateSuccessSummary(utils.ORGANIZATIONS, utils.DELETE)
		}
	}
}
