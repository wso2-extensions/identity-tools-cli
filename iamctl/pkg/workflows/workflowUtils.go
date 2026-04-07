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
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type workflow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type workflowListResponse struct {
	Workflows []workflow `json:"workflows"`
}

type workflowAssociation struct {
	ID           string `json:"id"`
	Name         string `json:"associationName"`
	WorkflowName string `json:"workflowName"`
}

type workflowAssociationListResponse struct {
	WorkflowAssociations []workflowAssociation `json:"workflowAssociations"`
}

type associationsOfWorkflowResponse struct {
	WorkflowAssociations []interface{} `json:"workflowAssociations"`
}

var exportedAssociationNames []string

func getWorkflowList() ([]workflow, error) {

	resp, err := utils.SendGetListRequest(utils.WORKFLOWS, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving workflow list: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		if errMsg, ok := utils.ErrorCodes[statusCode]; ok {
			return nil, fmt.Errorf("error while retrieving workflow list. Status code: %d, Error: %s", statusCode, errMsg)
		}
		return nil, fmt.Errorf("error while retrieving workflow list. Status code: %d", statusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error when reading the retrieved workflow list: %w", err)
	}

	var listResponse workflowListResponse
	err = json.Unmarshal(body, &listResponse)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling the retrieved workflow list: %w", err)
	}

	return listResponse.Workflows, nil
}

func getWorkflowAssociationsList() ([]workflowAssociation, error) {

	resp, err := utils.SendGetListRequest(utils.WORKFLOW_ASSOCIATIONS, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving workflow association list: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		if errMsg, ok := utils.ErrorCodes[statusCode]; ok {
			return nil, fmt.Errorf("error while retrieving workflow association list. Status code: %d, Error: %s", statusCode, errMsg)
		}
		return nil, fmt.Errorf("error while retrieving workflow association list. Status code: %d", statusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error when reading the retrieved workflow association list: %w", err)
	}

	var listResponse workflowAssociationListResponse
	if err := json.Unmarshal(body, &listResponse); err != nil {
		return nil, fmt.Errorf("error when unmarshalling the retrieved workflow association list: %w", err)
	}

	return listResponse.WorkflowAssociations, nil
}

func getDeployedWorkflowNames(workflows []workflow) []string {

	var names []string
	for _, wf := range workflows {
		names = append(names, wf.Name)
	}
	return names
}

func getWorkflowKeywordMapping(workflowName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.WorkflowConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(workflowName, utils.KEYWORD_CONFIGS.WorkflowConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getWorkflowId(name string, list []workflow) string {

	for _, wf := range list {
		if wf.Name == name {
			return wf.ID
		}
	}
	return ""
}

func getWfAssocId(name string, list []workflowAssociation) string {

	for _, assoc := range list {
		if assoc.Name == name {
			return assoc.ID
		}
	}
	return ""
}

func getAssociationsOfWorkflow(workflowId string) ([]interface{}, error) {

	resp, err := utils.SendGetListRequest(utils.WORKFLOW_ASSOCIATIONS, -1,
		utils.WithQueryParams(map[string]string{"filter": "workflowId eq " + workflowId}))
	if err != nil {
		return nil, fmt.Errorf("error while retrieving associations: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		if errMsg, ok := utils.ErrorCodes[statusCode]; ok {
			return nil, fmt.Errorf("error while retrieving associations. Status code: %d, Error: %s", statusCode, errMsg)
		}
		return nil, fmt.Errorf("error while retrieving associations for workflow. Status code: %d", statusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error when reading associations: %w", err)
	}
	var listResp associationsOfWorkflowResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("error when unmarshalling associations: %w", err)
	}

	for _, association := range listResp.WorkflowAssociations {
		if err := processWorkflowAssociation(association); err != nil {
			return nil, err
		}
	}

	return listResp.WorkflowAssociations, nil
}

func processWorkflowAssociation(association interface{}) error {

	assocMap, ok := association.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for association")
	}
	name, ok := assocMap["associationName"].(string)
	if !ok {
		return fmt.Errorf("unexpected format for associationName in association")
	}
	exportedAssociationNames = append(exportedAssociationNames, name)
	delete(assocMap, "workflowName")
	return nil
}

func removeUserStepOptions(wfMap map[string]interface{}) error {

	template, ok := wfMap["template"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for workflow template")
	}
	steps, ok := template["steps"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for workflow template steps")
	}

	for _, stepRaw := range steps {
		step, ok := stepRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for workflow step")
		}
		options, ok := step["options"].([]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for options in workflow step")
		}

		var filtered []interface{}
		for _, optRaw := range options {
			opt, ok := optRaw.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected format for option in workflow step")
			}
			entity, ok := opt["entity"].(string)
			if !ok {
				return fmt.Errorf("unexpected format for entity in workflow step option")
			}
			if entity != "users" {
				filtered = append(filtered, opt)
			}
		}
		step["options"] = filtered
	}
	return nil
}

func prepareWorkflowRequestBody(data []byte, format utils.Format) ([]byte, []map[string]interface{}, error) {

	wfMap, err := utils.DeserializeToMap(data, format, utils.WORKFLOWS, "id")
	if err != nil {
		return nil, nil, fmt.Errorf("error deserializing workflow file: %w", err)
	}

	if _, err := utils.ReplaceReferences(utils.WORKFLOWS, wfMap); err != nil {
		return nil, nil, err
	}

	assocRaw, ok := wfMap["associations"].([]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("unexpected format for workflow associations")
	}

	var associations []map[string]interface{}
	for _, item := range assocRaw {
		assocMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("unexpected format for association item in workflow file")
		}
		associations = append(associations, assocMap)
	}

	delete(wfMap, "associations")
	requestBody, err := utils.Serialize(wfMap, utils.FormatJSON, utils.WORKFLOWS)
	if err != nil {
		return nil, nil, fmt.Errorf("error serializing workflow request body: %w", err)
	}
	return requestBody, associations, nil
}

func prepareAssociationRequestBody(assocMap map[string]interface{}, workflowId string) ([]byte, error) {

	delete(assocMap, "id")
	delete(assocMap, "workflowName")
	assocMap["workflowId"] = workflowId

	body, err := json.Marshal(assocMap)
	if err != nil {
		return nil, fmt.Errorf("error serializing request body: %w", err)
	}
	return body, nil
}

func readLocalAssociationNames(importDirPath string) ([]string, error) {

	matches, err := filepath.Glob(filepath.Join(importDirPath, "WorkflowAssociations.*"))
	if err != nil {
		return nil, fmt.Errorf("error searching for file: %w", err)
	}
	if len(matches) == 0 {
		return []string{}, nil
	}

	fileBytes, err := ioutil.ReadFile(matches[0])
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	format, err := utils.FormatFromExtension(filepath.Ext(matches[0]))
	if err != nil {
		return nil, fmt.Errorf("unsupported format for file: %w", err)
	}

	var names []string
	if _, err := utils.Deserialize(fileBytes, format, utils.WORKFLOW_ASSOCIATIONS, &names); err != nil {
		return nil, fmt.Errorf("error deserializing file: %w", err)
	}
	return names, nil
}

func writeWorkflowAssociationsList(outputDirPath string, formatString string) error {

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, "WorkflowAssociations", format)

	data, err := utils.Serialize(exportedAssociationNames, format, utils.WORKFLOW_ASSOCIATIONS)
	if err != nil {
		return fmt.Errorf("error serializing workflow associations list: %w", err)
	}

	if err := ioutil.WriteFile(exportedFileName, data, 0644); err != nil {
		return fmt.Errorf("error writing workflow associations list: %w", err)
	}
	return nil
}

func updateWorkflowExportSummary(success bool, successCount int) {

	if !success {
		utils.UpdateFailureSummary(utils.WORKFLOW_ASSOCIATIONS, utils.WORKFLOW_ASSOCIATIONS.String())
		return
	}
	for i := 0; i < successCount; i++ {
		utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.EXPORT)
	}
}
