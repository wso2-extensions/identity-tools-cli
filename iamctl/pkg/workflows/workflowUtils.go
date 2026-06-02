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
	"net/http"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type workflow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

var assocSharingSupported bool
var assocRulesSupported bool
var exportedAssociationNames []string

func getWorkflowList() ([]workflow, error) {

	data, err := utils.SendPaginatedGetListRequest(
		utils.WORKFLOWS,
		"totalResults",
		"count",
		"offset",
		"limit",
		"workflows",
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving workflow list: %w", err)
	}
	var workflows []workflow
	if err := json.Unmarshal(data, &workflows); err != nil {
		return nil, fmt.Errorf("error when unmarshalling workflow list: %w", err)
	}
	return workflows, nil
}

func getWorkflowAssociationsList() ([]workflowAssociation, error) {

	data, err := utils.SendPaginatedGetListRequest(
		utils.WORKFLOW_ASSOCIATIONS,
		"totalResults",
		"count",
		"offset",
		"limit",
		"workflowAssociations",
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving workflow association list: %w", err)
	}
	var associations []workflowAssociation
	if err := json.Unmarshal(data, &associations); err != nil {
		return nil, fmt.Errorf("error when unmarshalling workflow association list: %w", err)
	}
	return associations, nil
}

func getDeployedWorkflowNames(workflows []workflow) []string {

	var names []string
	for _, wf := range workflows {
		names = append(names, wf.Name)
	}
	return names
}

func getLocalWfAssocNames(associations []map[string]interface{}) []string {

	var names []string
	for _, a := range associations {
		if name, ok := a["associationName"].(string); ok {
			names = append(names, name)
		}
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

func getAssociationsOfWorkflow(workflowId string) ([]interface{}, []workflowAssociation, error) {

	body, err := utils.SendGetListRequest(utils.WORKFLOW_ASSOCIATIONS,
		utils.WithQueryParams(map[string]string{"filter": "workflowId eq " + workflowId}))
	if err != nil {
		return nil, nil, fmt.Errorf("error while retrieving associations: %w", err)
	}
	var rawResp associationsOfWorkflowResponse
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, nil, fmt.Errorf("error when unmarshalling associations data: %w", err)
	}
	for _, association := range rawResp.WorkflowAssociations {
		if err := processWorkflowAssociation(association); err != nil {
			return nil, nil, err
		}
	}

	var structuredResp workflowAssociationListResponse
	if err := json.Unmarshal(body, &structuredResp); err != nil {
		return nil, nil, fmt.Errorf("error when parsing associations to struct: %w", err)
	}

	return rawResp.WorkflowAssociations, structuredResp.WorkflowAssociations, nil
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
	delete(assocMap, "rule")
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

	var filteredSteps []interface{}
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

		if len(filtered) > 0 {
			step["options"] = filtered
			filteredSteps = append(filteredSteps, step)
		}
	}

	if len(filteredSteps) == 0 {
		return fmt.Errorf("no valid steps remain after removing user options from workflow")
	}
	template["steps"] = filteredSteps
	return nil
}

func prepareWorkflowRequestBody(data []byte, format utils.Format) (requestBody []byte, associations []map[string]interface{}, err error) {

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

	for _, item := range assocRaw {
		assocMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("unexpected format for association item in workflow file")
		}
		associations = append(associations, assocMap)
	}

	delete(wfMap, "associations")
	requestBody, err = utils.Serialize(wfMap, utils.FormatJSON, utils.WORKFLOWS)
	if err != nil {
		return nil, nil, fmt.Errorf("error serializing workflow request body: %w", err)
	}
	return requestBody, associations, nil
}

func prepareAssociationRequestBody(assocMap map[string]interface{}, workflowId string) (requestBody []byte, isEnabled bool, err error) {

	delete(assocMap, "id")
	delete(assocMap, "workflowName")
	assocMap["workflowId"] = workflowId

	isEnabled, ok := assocMap["isEnabled"].(bool)
	if !ok {
		return nil, false, fmt.Errorf("unexpected format for isEnabled field")
	}

	body, err := json.Marshal(assocMap)
	if err != nil {
		return nil, false, fmt.Errorf("error serializing request body: %w", err)
	}
	return body, isEnabled, nil
}

func disableAssociation(resp *http.Response, requestBody []byte) error {

	var created workflowAssociation
	if _, err := utils.ParseResponseBody(resp, &created); err != nil {
		return fmt.Errorf("error reading create association response: %w", err)
	}

	patchResp, err := utils.SendPatchRequest(utils.WORKFLOW_ASSOCIATIONS, created.ID, requestBody)
	if err != nil {
		return fmt.Errorf("error updating association: %w", err)
	}
	defer patchResp.Body.Close()

	return nil
}

func readLocalAssociationNames(importDirPath string) ([]string, error) {

	matches, err := filepath.Glob(filepath.Join(importDirPath, utils.WORKFLOW_ASSOCIATIONS.String()+".*"))
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
	exportedFileName := utils.GetExportedFilePath(outputDirPath, utils.WORKFLOW_ASSOCIATIONS.String(), format)

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
		utils.UpdateFailureSummary(utils.WORKFLOWS, utils.WORKFLOW_ASSOCIATIONS.String())
		return
	}
	for i := 0; i < successCount; i++ {
		utils.UpdateSuccessSummary(utils.WORKFLOWS, utils.EXPORT)
	}
}

func setWorkflowVersionConfigs() {

	res, err := utils.CompareVersions(utils.SERVER_CONFIGS.ServerVersion, utils.MIN_VERSION_ASSOCIATION_SHARING_ACROSS_WORKFLOWS)

	// Assume association sharing is supported if the server version is "" (Asgardeo)
	if err != nil || res >= 0 {
		assocSharingSupported = true
	} else {
		assocSharingSupported = false
	}

	res, err = utils.CompareVersions(utils.SERVER_CONFIGS.ServerVersion, utils.MIN_VERSION_WORKFLOW_ASSOCIATION_RULES)

	// Assume association rules are supported if the server version is "" (Asgardeo)
	if err != nil || res >= 0 {
		assocRulesSupported = true
	} else {
		assocRulesSupported = false
	}
}
