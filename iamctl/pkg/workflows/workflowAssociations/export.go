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

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting workflow associations...")
	exportFilePath = filepath.Join(exportFilePath, utils.WORKFLOW_ASSOCIATIONS.String())

	if utils.IsResourceTypeExcluded(utils.WORKFLOW_ASSOCIATIONS) {
		return
	}
	associations, err := getWorkflowAssociationList()
	if err != nil {
		log.Println("Error: when exporting workflow associations.", err)
		return
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedNames := getDeployedWorkflowAssociationNames(associations)
			utils.RemoveDeletedLocalResources(exportFilePath, deployedNames)
		}
	}

	for _, assoc := range associations {
		err := exportWorkflowAssociation(assoc.ID, assoc.Name, exportFilePath, format)
		if err != nil {
			utils.UpdateFailureSummary(utils.WORKFLOW_ASSOCIATIONS, assoc.Name)
			log.Printf("Error while exporting workflow association %s of workflow %s: %s", assoc.Name, assoc.WorkflowName, err)
		} else {
			utils.UpdateSuccessSummary(utils.WORKFLOW_ASSOCIATIONS, utils.EXPORT)
		}

	}
}

func exportWorkflowAssociation(associationId string, associationName string, outputDirPath string, formatString string) error {

	assoc, err := utils.GetResourceData(utils.WORKFLOW_ASSOCIATIONS, associationId)
	if err != nil {
		return fmt.Errorf("error while getting workflow association: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, associationName, format)

	keywordMapping := getWorkflowAssociationKeywordMapping(associationName)
	modifiedAssoc, err := utils.ProcessExportedData(assoc, exportedFileName, format, keywordMapping, utils.WORKFLOW_ASSOCIATIONS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedAssoc, format, utils.WORKFLOW_ASSOCIATIONS)
	if err != nil {
		return fmt.Errorf("error while serializing workflow association: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
