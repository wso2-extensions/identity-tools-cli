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

package workflowAssociations

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing workflow associations...")
	importFilePath := filepath.Join(inputDirPath, utils.WORKFLOW_ASSOCIATIONS.String())

	if utils.IsResourceTypeExcluded(utils.WORKFLOW_ASSOCIATIONS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No workflow associations to import.")
		return
	}

	existingAssociations, err := getWorkflowAssociationList()
	if err != nil {
		log.Println("Error retrieving the deployed workflow association list:", err)
		return
	}

	files, err := os.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing workflow associations:", err)
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedAssociations(files, existingAssociations)
	}

	for _, file := range files {
		assocFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(assocFilePath)
		associationName := fileInfo.ResourceName

		associationId := getAssociationId(associationName, existingAssociations)
		err := importWorkflowAssociation(associationName, associationId, assocFilePath)
		if err != nil {
			log.Println("Error importing workflow association:", err)
			utils.UpdateFailureSummary(utils.WORKFLOW_ASSOCIATIONS, associationName)
		}
	}
}

func importWorkflowAssociation(associationName string, associationId string, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for workflow association: %w", err)
	}

	fileBytes, err := os.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for workflow association: %w", err)
	}

	keywordMapping := getWorkflowAssociationKeywordMapping(associationName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)
	requestBody, err := prepareAssociationRequestBody([]byte(modifiedFileData), format, "id")
	if err != nil {
		return fmt.Errorf("error preparing request body: %w", err)
	}

	if associationId == "" {
		return createAssociation(requestBody, format, associationName)
	}
	return updateAssociation(associationId, requestBody, format, associationName)
}

func createAssociation(requestBody []byte, format utils.Format, associationName string) error {

	resp, err := utils.SendPostRequest(utils.WORKFLOW_ASSOCIATIONS, requestBody)
	if err != nil {
		return fmt.Errorf("error when importing workflow association: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.WORKFLOW_ASSOCIATIONS, utils.IMPORT)
	return nil
}

func updateAssociation(associationId string, requestBody []byte, format utils.Format, associationName string) error {

	resp, err := utils.SendPatchRequest(utils.WORKFLOW_ASSOCIATIONS, associationId, requestBody)
	if err != nil {
		return fmt.Errorf("error when updating workflow association: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.WORKFLOW_ASSOCIATIONS, utils.UPDATE)
	return nil
}

func removeDeletedDeployedAssociations(localFiles []os.DirEntry, deployedAssociations []workflowAssociation) {

	if len(deployedAssociations) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, assoc := range deployedAssociations {
		if _, existsLocally := localResourceNames[assoc.Name]; existsLocally {
			continue
		}
		if err := utils.SendDeleteRequest(assoc.ID, utils.WORKFLOW_ASSOCIATIONS); err != nil {
			utils.UpdateFailureSummary(utils.WORKFLOW_ASSOCIATIONS, assoc.Name)
			log.Println("Error deleting workflow association:", assoc.Name, err)
		} else {
			utils.UpdateSuccessSummary(utils.WORKFLOW_ASSOCIATIONS, utils.DELETE)
		}
	}
}
