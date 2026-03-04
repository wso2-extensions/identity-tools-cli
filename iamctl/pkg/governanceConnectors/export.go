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

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting governance connectors...")
	exportFilePath = filepath.Join(exportFilePath, utils.GOVERNANCE_CONNECTORS.String())

	if utils.IsResourceTypeExcluded(utils.GOVERNANCE_CONNECTORS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedCategoryNames := getDeployedCategoryNames()
			utils.RemoveDeletedLocalDirectories(exportFilePath, deployedCategoryNames)
		}
	}

	categories, err := getCategoryList()
	if err != nil {
		log.Println("Error retrieving governance connector categories:", err)
		return
	}
	for _, catInfo := range categories {
		if !utils.IsResourceExcluded(catInfo.Name, utils.TOOL_CONFIGS.GovernanceConnectorConfigs) {
			log.Println("Exporting governance connector category:", catInfo.Name)

			err := exportCategory(catInfo.Id, catInfo.Name, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.GOVERNANCE_CONNECTORS, catInfo.Name)
				log.Printf("Error while exporting governance connector category: %s. %s", catInfo.Name, err)
			} else {
				utils.UpdateSuccessSummary(utils.GOVERNANCE_CONNECTORS, utils.EXPORT)
				log.Println("Governance connector category exported successfully:", catInfo.Name)
			}
		}

	}
}

func exportCategory(catId, catName, parentDir, formatString string) error {

	connectors, err := getConnectorListForCategory(catId)
	if err != nil {
		return fmt.Errorf("error retrieving connectors: %w", err)
	}

	format := utils.FormatFromString(formatString)
	categoryDir := filepath.Join(parentDir, catName)

	if _, err := os.Stat(categoryDir); os.IsNotExist(err) {
		if err := os.MkdirAll(categoryDir, 0700); err != nil {
			return fmt.Errorf("error creating connector category directory: %w", err)
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			getDeployedConnectorNames := getDeployedConnectorNames(connectors)
			utils.RemoveDeletedLocalResources(categoryDir, getDeployedConnectorNames)
		}
	}

	keywordMapping := getGovernanceCategoryKeywordMapping(catName)
	for _, c := range connectors {
		err := exportConnector(c.Id, c.FriendlyName, catId, categoryDir, format, keywordMapping)
		if err != nil {
			return fmt.Errorf("error while exporting connector: %s. %w", c.FriendlyName, err)
		}
	}

	return nil
}

func exportConnector(connectorId, connectorName, categoryId, categoryDir string, format utils.Format, keywordMapping map[string]interface{}) error {

	connectorData, err := utils.GetResourceData(utils.GOVERNANCE_CONNECTORS, categoryId+"/connectors/"+connectorId)
	if err != nil {
		return fmt.Errorf("error while getting connector: %w", err)
	}

	exportedFileName := utils.GetExportedFilePath(categoryDir, connectorName, format)

	modifiedData, err := utils.ProcessExportedData(connectorData, exportedFileName, format, keywordMapping, utils.GOVERNANCE_CONNECTORS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, utils.GOVERNANCE_CONNECTORS)
	if err != nil {
		return fmt.Errorf("error while serializing connector: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error writing exported content to file: %w", err)
	}

	return nil
}
