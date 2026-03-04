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

package governanceConnectors

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing governance connectors...")
	importFilePath := filepath.Join(inputDirPath, utils.GOVERNANCE_CONNECTORS.String())

	if utils.IsResourceTypeExcluded(utils.GOVERNANCE_CONNECTORS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No governance connectors to import.")
		return
	}

	deployedCategories, err := getCategoryList()
	if err != nil {
		log.Println("Error retrieving governance connector categories:", err)
		return
	}

	localCategoryDirs, err := os.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error reading governance connectors directory:", err)
		return
	}

	for _, entry := range localCategoryDirs {
		if !entry.IsDir() {
			continue
		}
		catName := entry.Name()
		localCategoryPath := filepath.Join(importFilePath, catName)

		if !utils.IsResourceExcluded(catName, utils.TOOL_CONFIGS.GovernanceConnectorConfigs) {
			err := importCategory(localCategoryPath, catName, deployedCategories)
			if err != nil {
				utils.UpdateFailureSummary(utils.GOVERNANCE_CONNECTORS, catName)
				log.Printf("Error importing governance connector category %s: %s", catName, err)
			}
		}
	}
}

func importCategory(localCategoryPath, catName string, deployedCategories []connectorCategory) error {

	catInfo := IsCategoryExists(catName, deployedCategories)
	if catInfo == nil {
		log.Printf("Governance connector category %s not found on server, skipping.", catName)
		return nil
	}

	deployedConnectors, err := getConnectorListForCategory(catInfo.Id)
	if err != nil {
		return fmt.Errorf("error retrieving connector list: %w", err)
	}

	localFiles, err := os.ReadDir(localCategoryPath)
	if err != nil {
		return fmt.Errorf("error reading local connector files: %w", err)
	}

	keywordMapping := getGovernanceCategoryKeywordMapping(catName)

	for _, file := range localFiles {
		filePath := filepath.Join(localCategoryPath, file.Name())
		fileInfo := utils.GetFileInfo(filePath)
		connectorName := fileInfo.ResourceName

		conId := getConnectorId(connectorName, deployedConnectors)
		if conId == "" {
			log.Printf("Connector %s not found on server, skipping.", connectorName)
			continue
		}

		err := importConnector(conId, catInfo.Id, filePath, keywordMapping)
		if err != nil {
			return fmt.Errorf("error importing connector: %s. %w", connectorName, err)
		}
	}

	utils.UpdateSuccessSummary(utils.GOVERNANCE_CONNECTORS, utils.UPDATE)
	log.Println("Governance connector category imported successfully:", catName)

	return nil
}

func importConnector(connectorId, categoryId, filePath string, keywordMapping map[string]interface{}) error {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for connector: %w", err)
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for connector: %w", err)
	}

	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	patchBody, err := buildPatchRequestBody([]byte(modifiedFileData), format)
	if err != nil {
		return err
	}

	resp, err := utils.SendPatchRequest(utils.GOVERNANCE_CONNECTORS, categoryId+"/connectors/"+connectorId, patchBody)
	if err != nil {
		return fmt.Errorf("error when updating connector: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
