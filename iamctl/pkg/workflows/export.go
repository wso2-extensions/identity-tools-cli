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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting workflows...")
	exportFilePath = filepath.Join(exportFilePath, utils.WORKFLOWS.String())

	if utils.IsResourceTypeExcluded(utils.WORKFLOWS) {
		return
	}
	workflows, err := getWorkflowList()
	if err != nil {
		log.Println("Error: when retrieving workflow list.", err)
		return
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedWorkflowNames := getDeployedWorkflowNames(workflows)
			utils.RemoveDeletedLocalResources(exportFilePath, deployedWorkflowNames)
		}
	}

	exportedAssociationNames = []string{}
	successCount := 0

	for _, wf := range workflows {
		if !utils.IsResourceExcluded(wf.Name, utils.TOOL_CONFIGS.WorkflowConfigs) {
			log.Println("Exporting workflow:", wf.Name)
			err := exportWorkflow(wf.ID, wf.Name, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.WORKFLOWS, wf.Name)
				log.Printf("Error while exporting workflow: %s. %s", wf.Name, err)
			} else {
				successCount++
				log.Println("Workflow exported successfully:", wf.Name)
			}
		}
	}

	err = writeWorkflowAssociationsList(exportFilePath, format)
	updateWorkflowExportSummary(err == nil, successCount)
	if err != nil {
		log.Println("Error writing workflow associations list:", err)
	}
}

func exportWorkflow(workflowId string, workflowName string, outputDirPath string, formatString string) error {

	wf, err := getWorkflowData(workflowId)
	if err != nil {
		return fmt.Errorf("error while getting workflow: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, workflowName, format)

	keywordMapping := getWorkflowKeywordMapping(workflowName)
	modifiedWf, err := utils.ProcessExportedData(wf, exportedFileName, format, keywordMapping, utils.WORKFLOWS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedWf, format, utils.WORKFLOWS)
	if err != nil {
		return fmt.Errorf("error while serializing workflow: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}

func getWorkflowData(workflowId string) (interface{}, error) {

	wf, err := utils.GetResourceData(utils.WORKFLOWS, workflowId)
	if err != nil {
		return nil, fmt.Errorf("error while getting workflow: %w", err)
	}

	wfMap, ok := wf.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for workflow data")
	}

	associations, err := getAssociationsOfWorkflow(workflowId)
	if err != nil {
		return nil, fmt.Errorf("error while getting workflow associations: %w", err)
	}
	wfMap["associations"] = associations

	return wfMap, nil
}
