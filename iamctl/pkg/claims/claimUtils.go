/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

package claims

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type claimDialect struct {
	Id         string `json:"id"`
	DialectURI string `json:"dialectURI"`
}

type ClaimDialectConfigurations struct {
	URI string `yaml:"dialectURI"`
	ID  string `yaml:"id"`
}

func getClaimDialectsList() ([]claimDialect, error) {

	var list []claimDialect
	resp, err := utils.SendGetListRequest(utils.CLAIMS, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving claim dialect list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved claim dialect list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved claim dialect list. %w", err)
		}
		resp.Body.Close()

		return list, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving claim dialect list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("unexpected error while retrieving claim dialect list")
}

func getClaimsList(dialectId string) ([]map[string]interface{}, error) {

	body, err := utils.SendGetRequest(utils.CLAIMS, dialectId+"/claims")
	if err != nil {
		return nil, fmt.Errorf("error while getting claims for dialect. %w", err)
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("error when unmarshalling claims list. %w", err)
	}
	return list, nil
}

func getClaimKeywordMapping(claimDialectName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ClaimConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(claimDialectName, utils.KEYWORD_CONFIGS.ClaimConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getDeployedDialectFileNames(claimDialects []claimDialect) []string {

	var claimDialectNames []string
	for _, claimDialect := range claimDialects {
		formattedName := formatFileName(claimDialect.DialectURI)
		claimDialectNames = append(claimDialectNames, formattedName)
	}
	return claimDialectNames
}

func parseClaims(data []byte, format utils.Format) ([]map[string]interface{}, error) {

	result, err := utils.Deserialize(data, format, utils.CLAIMS)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling claim dialect data: %w", err)
	}
	dialectMap, ok := utils.ConvertToStringKeyMap(result).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for claim dialect file")
	}
	rawClaims, ok := dialectMap["claims"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format: missing or invalid 'claims' array in claim dialect file")
	}
	var claims []map[string]interface{}
	for _, c := range rawClaims {
		m, ok := c.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format: claim entry is not a map")
		}
		claims = append(claims, m)
	}
	return claims, nil
}

func getDialectURIFromFile(filePath string) (string, error) {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return "", fmt.Errorf("unsupported file format: %w", err)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	var dialectConfig ClaimDialectConfigurations
	if _, err := utils.Deserialize(data, format, utils.CLAIMS, &dialectConfig); err != nil {
		return "", fmt.Errorf("error deserializing file: %w", err)
	}
	if dialectConfig.URI == "" {
		return "", fmt.Errorf("dialectURI not found in file")
	}
	return dialectConfig.URI, nil
}

func formatFileName(fileName string) string {

	formattedFileName := regexp.MustCompile(`[^\w\d]+`).ReplaceAllString(fileName, "_")
	if len(formattedFileName) > 255 {
		formattedFileName = formattedFileName[:255]
	}
	return formattedFileName
}

func getClaimDialectId(dialectURI string, existingClaimDialectList []claimDialect) string {

	for _, dialect := range existingClaimDialectList {
		if dialect.DialectURI == dialectURI {
			return dialect.Id
		}
	}
	return ""
}

func getClaimDialect(dialectId string) (interface{}, error) {

	dialectData, err := utils.GetResourceData(utils.CLAIMS, dialectId)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving claim dialect. %w", err)
	}
	dialectMap, ok := dialectData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for claim dialect response")
	}

	claims, err := utils.GetResourceData(utils.CLAIMS, dialectId+"/claims")
	if err != nil {
		return nil, fmt.Errorf("error while retrieving claims for dialect. %w", err)
	}
	dialectMap["claims"] = claims

	return dialectMap, nil
}

func createClaimReqBody(claim map[string]interface{}) ([]byte, error) {

	claimCopy := make(map[string]interface{}, len(claim))
	for k, v := range claim {
		if k != "id" && k != "claimDialectURI" {
			claimCopy[k] = v
		}
	}
	return json.Marshal(claimCopy)
}

func getClaimID(c map[string]interface{}) string {

	id, _ := c["id"].(string)
	return id
}

func getClaimURI(c map[string]interface{}) string {

	uri, _ := c["claimURI"].(string)
	return uri
}

func claimChanged(local, deployed map[string]interface{}) bool {

	localJSON, _ := createClaimReqBody(local)
	deployedJSON, _ := createClaimReqBody(deployed)
	return string(localJSON) != string(deployedJSON)
}

func exportAPIExists() bool {

	res, err := utils.CompareVersions(utils.SERVER_CONFIGS.ServerVersion, utils.MIN_VERSION_CLAIMS_EXPORT_API)
	if err != nil {
		// Use the export API when the server version is not properly configured for backward compatibility
		log.Println("Warn: Server version is not properly configured. For IS versions below 6.1, configure the server version properly to avoid failures.")
		return true
	}

	return res >= 0
}
