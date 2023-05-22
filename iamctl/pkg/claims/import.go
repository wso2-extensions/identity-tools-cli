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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing Claim Dialects...")
	importFilePath := filepath.Join(inputDirPath, utils.CLAIMS)

	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No claim dialects to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing claim dialects: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			removeDeletedDeployedClaimdialect(files)
		}
	}

	for i, file := range files {
		if file.Name() == "http_wso2_org_claims.yml" {
			files[0], files[i] = files[i], files[0]
			break
		}
	}

	for _, file := range files {
		claimFilePath := filepath.Join(importFilePath, file.Name())
		dialectName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(dialectName, utils.TOOL_CONFIGS.ClaimDialectConfigs) {
			dialectId, err := getClaimDialectId(claimFilePath)
			if err != nil {
				log.Printf("Invalid file configurations for Claim Dialect: %s. %s", dialectName, err)
			} else {
				err := importClaimDialect(dialectId, claimFilePath)
				if err != nil {
					log.Println("error importing claim dialect: ", err)
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

	if dialectId == "" {
		log.Println("Creating new claim dialect: " + fileInfo.ResourceName)
		err = utils.SendImportRequest(importFilePath, modifiedFileData, utils.CLAIMS)
	} else {
		log.Println("Updating claim dialect: " + fileInfo.ResourceName)
		err = utils.SendUpdateRequest(dialectId, importFilePath, modifiedFileData, utils.CLAIMS)
	}
	if err != nil {
		return fmt.Errorf("error when importing claim dialect: %s", err)
	}
	log.Println("Claim dialects imported successfully.")
	return nil
}

func removeDeletedDeployedClaimdialect(localFiles []os.FileInfo) {

	// Remove deployed claim dialects that do not exist locally.
	deployedClaimDialects, err := getClaimDialectsList()
	if err != nil {
		log.Println("Error retrieving deployed claim dialects: ", err)
		return
	}
deployedResourcess:
	for _, claimdialect := range deployedClaimDialects {
		for _, file := range localFiles {
			if claimdialect.DialectURI == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedResourcess
			}
		}
		if utils.IsResourceExcluded(claimdialect.DialectURI, utils.TOOL_CONFIGS.ClaimDialectConfigs) {
			log.Printf("Claim dialect: %s is excluded from deletion.\n", claimdialect.DialectURI)
			continue
		}
		log.Println("Claim dialect not found locally. Deleting claim dialect: ", claimdialect.DialectURI)
		err := utils.SendDeleteRequest(claimdialect.Id, utils.CLAIMS)
		if err != nil {
			log.Println("Error deleting claim dialect: ", err)
		}
	}
}
