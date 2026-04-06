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
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/workflows/workflowAssociations"
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
	files, err := os.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing workflows:", err)
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedWorkflows(files, existingWorkflows)
	}

	for _, file := range files {
		wfFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(wfFilePath)
		workflowName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(workflowName, utils.TOOL_CONFIGS.WorkflowConfigs) {
			workflowId := getWorkflowId(workflowName, existingWorkflows)
			err := importWorkflow(workflowName, workflowId, wfFilePath)
			if err != nil {
				log.Println("Error importing workflow:", err)
				utils.UpdateFailureSummary(utils.WORKFLOWS, workflowName)
			}
		}
	}

	workflowAssociations.ImportAll(importFilePath)
}

func importWorkflow(workflowName string, workflowId string, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for workflow: %w", err)
	}

	fileBytes, err := os.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for workflow: %w", err)
	}

	keywordMapping := getWorkflowKeywordMapping(workflowName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)
	requestBody, err := utils.PrepareJSONRequestBody([]byte(modifiedFileData), format, utils.WORKFLOWS, "id")
	if err != nil {
		return err
	}

	if workflowId == "" {
		return createWorkflow(requestBody, format, workflowName)
	}
	return updateWorkflow(workflowId, requestBody, format, workflowName)
}

func createWorkflow(requestBody []byte, format utils.Format, workflowName string) error {

	log.Println("Creating new workflow:", workflowName)

	resp, err := utils.SendPostRequest(utils.WORKFLOWS, requestBody)
	if err != nil {
		return fmt.Errorf("error when importing workflow: %w", err)
	}
	defer resp.Body.Close()

	parsed, err := utils.ParseResponseBody(resp)
	if err != nil {
		log.Println("Warning: Error parsing response body. Skipping resource identifier map entry.")
	} else {
		utils.ExtractAndRegisterIdentifier(utils.WORKFLOWS, parsed, utils.IMPORT)
	}

	utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.IMPORT)
	log.Println("Workflow imported successfully.")
	return nil
}

func updateWorkflow(workflowId string, requestBody []byte, format utils.Format, workflowName string) error {

	log.Println("Updating workflow:", workflowName)

	resp, err := utils.SendPutRequest(utils.WORKFLOWS, workflowId, requestBody)
	if err != nil {
		return fmt.Errorf("error when updating workflow: %w", err)
	}
	defer resp.Body.Close()

	utils.AddToIdentifierMap(utils.WORKFLOWS, workflowId, workflowName, utils.IMPORT)
	utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.UPDATE)
	log.Println("Workflow updated successfully.")
	return nil
}

func removeDeletedDeployedWorkflows(localFiles []os.DirEntry, deployedWorkflows []workflow) {

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
