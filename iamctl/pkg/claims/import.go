/**
* Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing claims...")
	if !utils.IsEntitySupportedInVersion(configs.CLAIMS) {
		return
	}
	if utils.IsSubOrganization() {
		log.Println("Importing claims for sub organization not supported.")
		return
	}
	importFilePath := filepath.Join(inputDirPath, configs.CLAIMS)

	if utils.IsResourceTypeExcluded(configs.CLAIMS) {
		return
	}
	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No claim dialects to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing claim dialects: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			removeDeletedDeployedClaimdialect(files, importFilePath)
		}
	}

	// Move the local claims file to the front of the array to import it first
	for i, file := range files {
		if file.Name() == "http_wso2_org_claims.yml" {
			files[0], files[i] = files[i], files[0]
			break
		}
	}

	for _, file := range files {
		claimFilePath := filepath.Join(importFilePath, file.Name())
		dialectName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(dialectName, utils.TOOL_CONFIGS.ClaimConfigs) {
			dialectId, err := getClaimDialectId(claimFilePath)
			if err != nil {
				log.Printf("Invalid file configurations for Claim Dialect: %s. %s", dialectName, err)
			} else {
				err := importClaimDialect(dialectId, claimFilePath)
				if err != nil {
					log.Println("error importing claim dialect:", err)
				}
			}
		}
	}
}

func importClaimDialect(dialectId string, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for claim dialect: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	claimKeywordMapping := getClaimKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), claimKeywordMapping)

	// Unmarshal the file data to get the dialect URI as the resource name.
	var claimDialectConfigurations ClaimDialectConfigurations
	error := yaml.Unmarshal([]byte(modifiedFileData), &claimDialectConfigurations)
	if error != nil {
		return fmt.Errorf("error when unmarshalling the file for claim dialect: %s", err)
	}
	fileInfo.ResourceName = claimDialectConfigurations.URI

	if dialectId == "" {
		return importDialect(importFilePath, modifiedFileData, fileInfo)
	}
	return updateDialect(dialectId, importFilePath, modifiedFileData, fileInfo)
}

func importDialect(importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Creating new claim dialect: " + fileInfo.ResourceName)
	err := utils.SendImportRequest(importFilePath, modifiedFileData, configs.CLAIMS)
	if err != nil {
		utils.UpdateFailureSummary(configs.CLAIMS, fileInfo.ResourceName)
		return fmt.Errorf("error when importing claim dialect: %s", err)
	}
	utils.UpdateSuccessSummary(configs.CLAIMS, utils.IMPORT)
	log.Println("Claim dialect imported successfully.")
	return nil
}

func updateDialect(dialectId string, importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Updating claim dialect: " + fileInfo.ResourceName)
	err := utils.SendUpdateRequest(dialectId, importFilePath, modifiedFileData, configs.CLAIMS)
	if err != nil {
		utils.UpdateFailureSummary(configs.CLAIMS, fileInfo.ResourceName)
		return fmt.Errorf("error when updating claim dialect: %s", err)
	}
	utils.UpdateSuccessSummary(configs.CLAIMS, utils.UPDATE)
	log.Println("Claim dialect updated successfully.")
	return nil
}

func removeDeletedDeployedClaimdialect(localFiles []os.FileInfo, importFilePath string) {

	// Remove deployed claim dialects that do not exist locally.
	deployedClaimDialects, err := getClaimDialectsList()
	if err != nil {
		log.Println("Error retrieving deployed claim dialects: ", err)
		return
	}
deployedResourcess:
	for _, claimDialect := range deployedClaimDialects {
		for _, file := range localFiles {

			var claimDialectConfigurations ClaimDialectConfigurations
			claimFilePath := filepath.Join(importFilePath, file.Name())

			content, err := readFileContent(claimFilePath)
			if err != nil {
				log.Println("error when reading file content: ", err)
			}

			err = yaml.Unmarshal(content, &claimDialectConfigurations)
			if err != nil {
				log.Println("error when unmarshalling the file for claim dialect: ", err)
			}

			localResourceName := claimDialectConfigurations.URI
			if claimDialect.DialectURI == localResourceName {
				continue deployedResourcess
			}
		}
		if utils.IsResourceExcluded(claimDialect.DialectURI, utils.TOOL_CONFIGS.ClaimConfigs) {
			log.Printf("Claim dialect: %s is excluded from deletion.\n", claimDialect.DialectURI)
			continue
		}
		log.Println("Claim dialect not found locally. Deleting claim dialect: ", claimDialect.DialectURI)
		err := utils.SendDeleteRequest(claimDialect.Id, configs.CLAIMS)
		if err != nil {
			log.Println("Error deleting claim dialect: ", err)
		}
	}
}

func readFileContent(filename string) ([]byte, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return content, nil
}
