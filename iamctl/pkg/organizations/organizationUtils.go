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
	"net/http"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type organization struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	OrgHandle string `json:"orgHandle"`
	Status    string `json:"status"`
}

type orgLink struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type organizationsResponse struct {
	Organizations []organization `json:"organizations"`
	Links         []orgLink      `json:"links"`
}

const (
	creatorIdKey       = "creator.id"
	creatorUsernameKey = "creator.username"
)

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

	body, err := utils.SendGetListRequest(utils.ORGANIZATIONS,
		utils.WithQueryParams(map[string]string{"recursive": "false"}))
	if err != nil {
		return nil, fmt.Errorf("error while retrieving organization list: %w", err)
	}
	var page organizationsResponse
	if err = json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("error when unmarshalling organization list: %w", err)
	}
	allResults := page.Organizations

	for {
		nextHref := ""
		for _, link := range page.Links {
			if link.Rel == "next" {
				nextHref = link.Href
				break
			}
		}
		if nextHref == "" {
			break
		}

		resp, err := utils.SendCustomRequest(http.MethodGet, utils.SERVER_CONFIGS.ServerUrl+nextHref, nil, "")
		if err != nil {
			return nil, fmt.Errorf("error retrieving page of organization list: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			if errMsg, ok := utils.ErrorCodes[resp.StatusCode]; ok {
				return nil, fmt.Errorf("error response for organization list page request: %s", errMsg)
			}
			return nil, fmt.Errorf("unexpected error when retrieving organization list page: %s", resp.Status)
		}

		nextBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error reading page of organization list: %w", err)
		}
		page = organizationsResponse{}
		if err = json.Unmarshal(nextBody, &page); err != nil {
			return nil, fmt.Errorf("error when unmarshalling organization list page: %w", err)
		}
		allResults = append(allResults, page.Organizations...)
	}

	return allResults, nil
}

func getDeployedOrgResourceNames(orgs []organization) []string {

	var names []string
	for _, o := range orgs {
		names = append(names, getOrgResourceName(o))
	}
	return names
}

func getOrganizationKeywordMapping(resourceName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.OrganizationConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(resourceName, utils.KEYWORD_CONFIGS.OrganizationConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getOrgId(resourceName string, list []organization) string {

	for _, o := range list {
		if getOrgResourceName(o) == resourceName {
			return o.Id
		}
	}
	return ""
}

func prepareOrganizationPostBody(requestBody []byte, format utils.Format, parentId string) (reqBody []byte, status interface{}, err error) {

	orgData, err := utils.DeserializeToMap(requestBody, format, utils.ORGANIZATIONS,
		"id", "parent", "version", "permissions", "created", "lastModified", "hasChildren", "ancestorPath")
	if err != nil {
		return nil, nil, fmt.Errorf("error deserializing organization: %w", err)
	}

	orgData["parentId"] = parentId
	status = orgData["status"]
	delete(orgData, "status")

	// orgHandle is removed from POST requests for Asgardeo
	if utils.SERVER_CONFIGS.ServerVersion == "" {
		delete(orgData, "orgHandle")
	}

	if err := addCreatorAttributes(orgData); err != nil {
		return nil, nil, err
	}

	jsonBody, err := utils.Serialize(orgData, utils.FormatJSON, utils.ORGANIZATIONS)
	if err != nil {
		return nil, nil, fmt.Errorf("error serializing to JSON: %w", err)
	}
	return jsonBody, status, nil
}

func addCreatorAttributes(orgData map[string]interface{}) error {

	if utils.TOOL_CONFIGS.OrganizationConfigs == nil {
		return nil
	}

	creatorId, hasId := utils.TOOL_CONFIGS.OrganizationConfigs["CREATOR_ID"]
	creatorUsername, hasUsername := utils.TOOL_CONFIGS.OrganizationConfigs["CREATOR_USERNAME"]
	if !hasId || !hasUsername {
		return nil
	}

	idStr, idOk := creatorId.(string)
	if !idOk {
		return fmt.Errorf("unexpected format for CREATOR_ID in organization configs")
	}
	usernameStr, usernameOk := creatorUsername.(string)
	if !usernameOk {
		return fmt.Errorf("unexpected format for CREATOR_USERNAME in organization configs")
	}
	if idStr == "" || usernameStr == "" {
		return nil
	}

	var attributes []interface{}
	if existing, hasAttr := orgData["attributes"]; hasAttr {
		attrArr, ok := existing.([]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for attributes field")
		}
		attributes = attrArr
	}

	attributes = append(attributes,
		map[string]interface{}{"key": creatorIdKey, "value": idStr},
		map[string]interface{}{"key": creatorUsernameKey, "value": usernameStr},
	)

	orgData["attributes"] = attributes
	return nil
}

func removeCreatorAttributes(orgData interface{}) (interface{}, error) {

	dataMap, ok := orgData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for organization data")
	}

	existing, hasAttr := dataMap["attributes"]
	if !hasAttr {
		return orgData, nil
	}
	attrArr, ok := existing.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for attributes field")
	}

	var filtered []interface{}
	for _, entry := range attrArr {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for attributes entry")
		}
		keyVal, ok := entryMap["key"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected format for key in attributes entry")
		}
		if keyVal != creatorIdKey && keyVal != creatorUsernameKey {
			filtered = append(filtered, entry)
		}
	}

	if len(filtered) == 0 {
		delete(dataMap, "attributes")
	} else {
		dataMap["attributes"] = filtered
	}
	return dataMap, nil
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

func getOrgResourceName(org organization) string {

	// Uses org name as the resource identifier in Asgardeo, org handle for IS.
	if utils.SERVER_CONFIGS.ServerVersion == "" {
		return org.Name
	}
	return org.OrgHandle
}
