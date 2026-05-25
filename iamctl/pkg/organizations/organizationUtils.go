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

package organizations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type organization struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type organizationsResponse struct {
	Organizations []organization `json:"organizations"`
}

var curOrgId string

func GetCurrentOrganizationId() (id string, err error) {

	org, err := utils.SendGetRequest(utils.ORGANIZATIONS, "self")
	if err != nil {
		return "", fmt.Errorf("error while getting organization: %w", err)
	}

	var curOrg organization
	if _, err := utils.Deserialize(org, utils.FormatJSON, utils.ORGANIZATIONS, &curOrg); err != nil {
		return "", fmt.Errorf("error while deserializing JSON response: %w", err)
	}
	return curOrg.Id, nil
}

func getOrganizationList() ([]organization, error) {

	resp, err := utils.SendGetListRequest(utils.ORGANIZATIONS, -1,
		utils.WithQueryParams(map[string]string{"recursive": "false"}))
	if err != nil {
		return nil, fmt.Errorf("error while retrieving list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved list. %w", err)
		}

		var wrapper organizationsResponse
		err = json.Unmarshal(body, &wrapper)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved list. %w", err)
		}
		return wrapper.Organizations, nil

	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("unknown error while retrieving list")
}

func getDeployedOrganizationNames(orgs []organization) []string {

	var orgNames []string
	for _, o := range orgs {
		orgNames = append(orgNames, o.Name)
	}
	return orgNames
}

func getOrganizationKeywordMapping(orgName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.OrganizationConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(orgName, utils.KEYWORD_CONFIGS.OrganizationConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getOrgId(orgName string, list []organization) string {

	for _, o := range list {
		if o.Name == orgName {
			return o.Id
		}
	}
	return ""
}

func prepareOrganizationPostBody(requestBody []byte, format utils.Format, parentId string) ([]byte, interface{}, error) {

	orgData, err := utils.DeserializeToMap(requestBody, format, utils.ORGANIZATIONS,
		"id", "parent", "version", "permissions", "created", "lastModified", "hasChildren", "ancestorPath")
	if err != nil {
		return nil, nil, fmt.Errorf("error deserializing organization: %w", err)
	}

	orgData["parentId"] = parentId
	status := orgData["status"]
	delete(orgData, "status")

	// orgHandle is removed from POST requests for Asgardeo
	if utils.SERVER_CONFIGS.ServerVersion == "" {
		delete(orgData, "orgHandle")
	}

	jsonBody, err := utils.Serialize(orgData, utils.FormatJSON, utils.ORGANIZATIONS)
	if err != nil {
		return nil, nil, fmt.Errorf("error serializing to JSON: %w", err)
	}
	return jsonBody, status, nil
}

func patchOrganizationStatus(orgId string, rawStatus interface{}) error {

	status, ok := rawStatus.(string)
	if !ok {
		return fmt.Errorf("unexpected format for status field")
	}

	patchBody := []map[string]string{
		{
			"operation": "REPLACE",
			"path":      "/status",
			"value":     status,
		},
	}
	jsonBody, err := json.Marshal(patchBody)
	if err != nil {
		return fmt.Errorf("error serializing PATCH body: %w", err)
	}

	resp, err := utils.SendPatchRequest(utils.ORGANIZATIONS, orgId, jsonBody)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}
