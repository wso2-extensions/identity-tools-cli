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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, "", "Importing workflows...")
	importFilePath := filepath.Join(inputDirPath, utils.WORKFLOWS.String())
	setWorkflowVersionConfigs()

	if !utils.IsEntitySupportedInVersion(utils.WORKFLOWS) || !utils.IsEntitySupportedInOrg(utils.WORKFLOWS) || utils.IsResourceTypeExcluded(utils.WORKFLOWS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, "", "No workflows to import.")
		return
	}

	existingWorkflows, err := getWorkflowList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, "", fmt.Sprintf("Error retrieving the deployed workflow list: %s", err))
		return
	}
	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, "", fmt.Sprintf("Error importing workflows: %s", err))
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedWorkflows(files, existingWorkflows)
	}

	var existingAssoc []workflowAssociation
	failedWorkflows := make(map[string]struct{})
	if !assocSharingSupported {
		existingAssoc, err = getWorkflowAssociationsList()
		if err != nil {
			utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, "", fmt.Sprintf("Error retrieving the deployed workflow association list: %s", err))
			utils.UpdateFailureSummary(utils.WORKFLOWS, utils.WORKFLOW_ASSOCIATIONS.String())
			return
		}

		if utils.TOOL_CONFIGS.AllowDelete {
			localAssoc, err := readLocalAssociationNames(importFilePath)
			if err != nil {
				utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, "", fmt.Sprintf("Error reading local workflow association list: %s", err))
				utils.UpdateFailureSummary(utils.WORKFLOWS, utils.WORKFLOW_ASSOCIATIONS.String())
				return
			}
			failedWorkflows, _ = removeDeletedDeployedWfAssociations(localAssoc, existingAssoc)
		}
	}

	for _, file := range files {
		wfFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(wfFilePath)
		workflowName := fileInfo.ResourceName

		if workflowName == utils.WORKFLOW_ASSOCIATIONS.String() {
			continue
		}
		if _, failed := failedWorkflows[workflowName]; failed {
			utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, workflowName, "Skipping workflow: deleting stale workflow associations failed")
			utils.UpdateFailureSummary(utils.WORKFLOWS, workflowName)
			continue
		}

		if !utils.IsResourceExcluded(workflowName, utils.TOOL_CONFIGS.WorkflowConfigs) {
			workflowId := getWorkflowId(workflowName, existingWorkflows)
			if err := importWorkflow(workflowName, workflowId, wfFilePath, existingAssoc); err != nil {
				utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, workflowName, fmt.Sprintf("Error importing workflow: %s", err))
				utils.UpdateFailureSummary(utils.WORKFLOWS, workflowName)
			}
		}
	}
	utils.PrintLog(utils.LogLevelWarn, utils.WORKFLOWS, "", "Users associated with workflow steps are removed during import")
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

	utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, workflowName, "Creating new workflow")

	resp, err := utils.SendPostRequest(utils.WORKFLOWS, requestBody)
	if err != nil {
		return fmt.Errorf("error when importing workflow: %w", err)
	}
	defer resp.Body.Close()

	var created workflow
	if _, err := utils.ParseResponseBody(resp, &created); err != nil {
		return fmt.Errorf("error reading create workflow response: %w", err)
	}
	if err := syncWorkflowAssociations(created.ID, associations, existingAssoc); err != nil {
		return fmt.Errorf("error syncing workflow associations: %w", err)
	}

	utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, workflowName, "Imported successfully")
	return nil
}

func updateWorkflow(workflowId string, requestBody []byte, workflowName string, associations []map[string]interface{}, existingAssoc []workflowAssociation) error {

	utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, workflowName, "Updating workflow")

	resp, err := utils.SendPutRequest(utils.WORKFLOWS, workflowId, requestBody)
	if err != nil {
		return fmt.Errorf("error when updating workflow: %w", err)
	}
	defer resp.Body.Close()

	if err := syncWorkflowAssociations(workflowId, associations, existingAssoc); err != nil {
		return fmt.Errorf("error syncing workflow associations: %w", err)
	}

	utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, workflowName, "Updated successfully")
	return nil
}

func syncWorkflowAssociations(workflowId string, associations []map[string]interface{}, deployedAssoc []workflowAssociation) error {

	if assocSharingSupported {
		_, assocs, err := getAssociationsOfWorkflow(workflowId)
		if err != nil {
			return fmt.Errorf("error retrieving deployed workflow associations: %w", err)
		}
		deployedAssoc = assocs

		if utils.TOOL_CONFIGS.AllowDelete {
			if _, err := removeDeletedDeployedWfAssociations(getLocalWfAssocNames(associations), deployedAssoc); err != nil {
				return fmt.Errorf("error removing deleted workflow associations: %w", err)
			}
		}
	}

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
			utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, wf.Name, "Excluded from deletion.")
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, utils.WORKFLOWS, wf.Name, "Not found locally. Deleting workflow.")
		if err := utils.SendDeleteRequest(wf.ID, utils.WORKFLOWS); err != nil {
			utils.UpdateFailureSummary(utils.WORKFLOWS, wf.Name)
			utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, wf.Name, fmt.Sprintf("Error deleting workflow: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.DELETE)
		}
	}
}

func removeDeletedDeployedWfAssociations(localNames []string, deployedAssociations []workflowAssociation) (failedWorkflows map[string]struct{}, err error) {

	failedWorkflows = make(map[string]struct{})
	if len(deployedAssociations) == 0 {
		return failedWorkflows, nil
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
			if assocSharingSupported {
				return nil, fmt.Errorf("error deleting workflow association: %s. %w", assoc.Name, err)
			} else {
				utils.PrintLog(utils.LogLevelError, utils.WORKFLOWS, assoc.WorkflowName, fmt.Sprintf("Error deleting workflow association %s: %s", assoc.Name, err))
				failedWorkflows[assoc.WorkflowName] = struct{}{}
			}
		}
	}
	return failedWorkflows, nil
}
