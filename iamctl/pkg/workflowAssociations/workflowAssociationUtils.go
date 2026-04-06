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
	"encoding/json"
	"fmt"
	"io"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type workflowAssociation struct {
	ID   string `json:"id"`
	Name string `json:"associationName"`
}

type workflowAssociationListResponse struct {
	WorkflowAssociations []workflowAssociation `json:"workflowAssociations"`
}

func getWorkflowAssociationList() ([]workflowAssociation, error) {

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error when reading the retrieved workflow association list: %w", err)
	}

	var listResponse workflowAssociationListResponse
	err = json.Unmarshal(body, &listResponse)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling the retrieved workflow association list: %w", err)
	}

	return listResponse.WorkflowAssociations, nil
}

func getDeployedWorkflowAssociationNames(associations []workflowAssociation) []string {

	var names []string
	for _, assoc := range associations {
		names = append(names, assoc.Name)
	}
	return names
}

func getWorkflowAssociationKeywordMapping(associationName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.WorkflowAssociationConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(associationName, utils.KEYWORD_CONFIGS.WorkflowAssociationConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getAssociationId(name string, list []workflowAssociation) string {

	for _, assoc := range list {
		if assoc.Name == name {
			return assoc.ID
		}
	}
	return ""
}

func prepareAssociationRequestBody(requestBody []byte, format utils.Format, excludeFields ...string) ([]byte, error) {

	dataMap, err := utils.DeserializeToMap(requestBody, format, utils.WORKFLOW_ASSOCIATIONS, excludeFields...)
	if err != nil {
		return nil, fmt.Errorf("error deserializing association data: %w", err)
	}

	replaced, err := utils.ReplaceReferences(utils.WORKFLOW_ASSOCIATIONS, dataMap)
	if err != nil {
		return nil, err
	}
	dataMap = replaced.(map[string]interface{})

	val, ok := dataMap["workflowName"]
	if !ok {
		return nil, fmt.Errorf("workflowName not found in the workflow association")
	}
	dataMap["workflowId"] = val
	delete(dataMap, "workflowName")

	jsonBody, err := utils.Serialize(dataMap, utils.FormatJSON, utils.WORKFLOW_ASSOCIATIONS)
	if err != nil {
		return nil, fmt.Errorf("error serializing to json: %w", err)
	}
	return jsonBody, nil
}
