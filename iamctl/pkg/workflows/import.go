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

package workflows

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

	log.Println("Importing workflows...")
	importFilePath := filepath.Join(inputDirPath, utils.WORKFLOWS.String())

	if utils.IsResourceTypeExcluded(utils.WORKFLOWS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No workflows to import.")
		return
	}

	existingWorkflows, err := getWorkflowList()
	if err != nil {
		log.Println("Error retrieving the deployed workflow list:", err)
		return
	}
	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing workflows:", err)
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedWorkflows(files, existingWorkflows)
	}

	localAssoc, err := readLocalAssociationNames(importFilePath)
	if err != nil {
		log.Println("Error reading local workflow associations list:", err)
		utils.UpdateFailureSummary(utils.WORKFLOWS, utils.WORKFLOW_ASSOCIATIONS.String())
		return
	}
	existingAssoc, err := getWorkflowAssociationsList()
	if err != nil {
		log.Println("Error retrieving the deployed workflow association list:", err)
		utils.UpdateFailureSummary(utils.WORKFLOWS, utils.WORKFLOW_ASSOCIATIONS.String())
		return
	}
	var failedWorkflows map[string]struct{}
	if utils.TOOL_CONFIGS.AllowDelete {
		failedWorkflows = removeDeletedDeployedWfAssociations(localAssoc, existingAssoc)
	}

	for _, file := range files {
		wfFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(wfFilePath)
		workflowName := fileInfo.ResourceName

		if workflowName == "WorkflowAssociations" {
			continue
		}
		if _, failed := failedWorkflows[workflowName]; failed {
			log.Printf("Skipping workflow %s: deleting stale workflow associations failed", workflowName)
			utils.UpdateFailureSummary(utils.WORKFLOWS, workflowName)
			continue
		}

		if !utils.IsResourceExcluded(workflowName, utils.TOOL_CONFIGS.WorkflowConfigs) {
			workflowId := getWorkflowId(workflowName, existingWorkflows)
			if err := importWorkflow(workflowName, workflowId, wfFilePath, existingAssoc); err != nil {
				log.Println("Error importing workflow:", err)
				utils.UpdateFailureSummary(utils.WORKFLOWS, workflowName)
			}
		}
	}
	log.Println("Warn: Users associated with workflow steps are removed during import")
}

func importWorkflow(workflowName string, workflowId string, importFilePath string, existingAssoc []workflowAssociation) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for workflow: %w", err)
	}
	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for workflow: %w", err)
	}

	keywordMapping := getWorkflowKeywordMapping(workflowName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	requestBody, associations, err := prepareWorkflowRequestBody([]byte(modifiedFileData), format)
	if err != nil {
		return err
	}

	if workflowId == "" {
		return createWorkflow(requestBody, workflowName, associations, existingAssoc)
	}
	return updateWorkflow(workflowId, requestBody, workflowName, associations, existingAssoc)
}

func createWorkflow(requestBody []byte, workflowName string, associations []map[string]interface{}, existingAssoc []workflowAssociation) error {

	log.Println("Creating new workflow:", workflowName)

	resp, err := utils.SendPostRequest(utils.WORKFLOWS, requestBody)
	if err != nil {
		return fmt.Errorf("error when importing workflow: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading create workflow response: %w", err)
	}

	var created workflow
	if err := json.Unmarshal(respBody, &created); err != nil {
		return fmt.Errorf("error parsing create workflow response: %w", err)
	}
	if err := syncWorkflowAssociations(created.ID, associations, existingAssoc); err != nil {
		return fmt.Errorf("error syncing workflow associations: %w", err)
	}

	utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.IMPORT)
	log.Println("Workflow imported successfully.")
	return nil
}

func updateWorkflow(workflowId string, requestBody []byte, workflowName string, associations []map[string]interface{}, existingAssoc []workflowAssociation) error {

	log.Println("Updating workflow:", workflowName)

	resp, err := utils.SendPutRequest(utils.WORKFLOWS, workflowId, requestBody)
	if err != nil {
		return fmt.Errorf("error when updating workflow: %w", err)
	}
	defer resp.Body.Close()

	if err := syncWorkflowAssociations(workflowId, associations, existingAssoc); err != nil {
		return fmt.Errorf("error syncing workflow associations: %w", err)
	}

	utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.UPDATE)
	log.Println("Workflow updated successfully.")
	return nil
}

func syncWorkflowAssociations(workflowId string, associations []map[string]interface{}, deployedAssoc []workflowAssociation) error {

	for _, assocMap := range associations {
		assocName, ok := assocMap["associationName"].(string)
		if !ok {
			return fmt.Errorf("invalid format for associationName")
		}
		requestBody, isEnabled, err := prepareAssociationRequestBody(assocMap, workflowId)
		if err != nil {
			return fmt.Errorf("error preparing request body for association %s: %w", assocName, err)
		}

		if assocId := getWfAssocId(assocName, deployedAssoc); assocId != "" {
			if err := updateAssociation(assocId, requestBody, assocName); err != nil {
				return err
			}
		} else {
			if err := createAssociation(requestBody, assocName, isEnabled); err != nil {
				return err
			}
		}
	}
	return nil
}

func createAssociation(requestBody []byte, associationName string, isEnabled bool) error {

	resp, err := utils.SendPostRequest(utils.WORKFLOW_ASSOCIATIONS, requestBody)
	if err != nil {
		return fmt.Errorf("error when creating workflow association %s: %w", associationName, err)
	}
	defer resp.Body.Close()

	if !isEnabled {
		if err := disableAssociation(resp, requestBody); err != nil {
			return fmt.Errorf("error when disabling workflow association %s: %w", associationName, err)
		}
	}
	return nil
}

func updateAssociation(associationId string, requestBody []byte, associationName string) error {

	resp, err := utils.SendPatchRequest(utils.WORKFLOW_ASSOCIATIONS, associationId, requestBody)
	if err != nil {
		return fmt.Errorf("error when updating workflow association %s: %w", associationName, err)
	}
	defer resp.Body.Close()

	return nil
}

func removeDeletedDeployedWorkflows(localFiles []os.FileInfo, deployedWorkflows []workflow) {

	if len(deployedWorkflows) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, wf := range deployedWorkflows {
		if _, existsLocally := localResourceNames[wf.Name]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(wf.Name, utils.TOOL_CONFIGS.WorkflowConfigs) {
			log.Println("Workflow is excluded from deletion:", wf.Name)
			continue
		}

		log.Printf("Workflow: %s not found locally. Deleting workflow.\n", wf.Name)
		if err := utils.SendDeleteRequest(wf.ID, utils.WORKFLOWS); err != nil {
			utils.UpdateFailureSummary(utils.WORKFLOWS, wf.Name)
			log.Println("Error deleting workflow:", wf.Name, err)
		} else {
			utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.DELETE)
		}
	}
}

func removeDeletedDeployedWfAssociations(localNames []string, deployedAssociations []workflowAssociation) map[string]struct{} {

	failedWorkflows := make(map[string]struct{})
	if len(deployedAssociations) == 0 {
		return failedWorkflows
	}

	localSet := make(map[string]struct{})
	for _, name := range localNames {
		localSet[name] = struct{}{}
	}

	for _, assoc := range deployedAssociations {
		if _, existsLocally := localSet[assoc.Name]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(assoc.WorkflowName, utils.TOOL_CONFIGS.WorkflowConfigs) {
			continue
		}
		if err := utils.SendDeleteRequest(assoc.ID, utils.WORKFLOW_ASSOCIATIONS); err != nil {
			log.Printf("Error deleting workflow association %s of workflow %s: %v", assoc.Name, assoc.WorkflowName, err)
			failedWorkflows[assoc.WorkflowName] = struct{}{}
		}
	}
	return failedWorkflows
}
