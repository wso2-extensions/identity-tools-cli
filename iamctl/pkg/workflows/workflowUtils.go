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
	"io"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type workflow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type workflowListResponse struct {
	Workflows []workflow `json:"workflows"`
}

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

	body, err := io.ReadAll(resp.Body)
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
