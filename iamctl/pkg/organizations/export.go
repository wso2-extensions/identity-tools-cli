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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting organizations...")
	exportFilePath = filepath.Join(exportFilePath, utils.ORGANIZATIONS.String())

	if !utils.IsEntitySupportedInVersion(utils.ORGANIZATIONS) || !utils.IsEntitySupportedInOrg(utils.ORGANIZATIONS) || utils.IsResourceTypeExcluded(utils.ORGANIZATIONS) {
		return
	}
	orgs, err := getOrganizationList()
	if err != nil {
		log.Println("Error retrieving organizations list:", err)
		return
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			log.Println("Error creating organizations directory:", err)
			return
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedResourceNames := getDeployedOrgResourceNames(orgs)
			utils.RemoveDeletedLocalResources(exportFilePath, deployedResourceNames)
		}
	}

	for _, org := range orgs {
		resourceName := getOrgResourceName(org)
		if !utils.IsResourceExcluded(resourceName, utils.TOOL_CONFIGS.OrganizationConfigs) {
			log.Println("Exporting organization: ", resourceName)

			err := exportOrganization(org.Id, resourceName, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.ORGANIZATIONS, resourceName)
				log.Printf("Error while exporting organization: %s. %s", resourceName, err)
			} else {
				utils.UpdateSuccessSummary(utils.ORGANIZATIONS, utils.EXPORT)
				log.Println("Organization exported successfully: ", resourceName)
			}
		}
	}
}

func exportOrganization(orgId, resourceName, outputDirPath, formatString string) error {

	org, err := utils.GetResourceData(utils.ORGANIZATIONS, orgId)
	if err != nil {
		return fmt.Errorf("error while getting organization: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, resourceName, format)

	orgKeywordMapping := getOrganizationKeywordMapping(resourceName)
	modifiedOrg, err := utils.ProcessExportedData(org, exportedFileName, format, orgKeywordMapping, utils.ORGANIZATIONS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedOrg, format, utils.ORGANIZATIONS)
	if err != nil {
		return fmt.Errorf("error while serializing organization: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
