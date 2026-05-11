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

package roles

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting roles...")
	exportFilePath = filepath.Join(exportFilePath, utils.ROLES.String())

	if utils.IsResourceTypeExcluded(utils.ROLES) {
		return
	}
	roles, err := getRoleList()
	if err != nil {
		log.Println("Error: when exporting roles.", err)
		return
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedRoleNames := getDeployedRoleLocalFileNames(roles)
			utils.RemoveDeletedLocalResources(exportFilePath, deployedRoleNames)
		}
	}

	for _, r := range roles {
		if !utils.IsResourceExcluded(r.DisplayName, utils.TOOL_CONFIGS.RoleConfigs) {
			log.Println("Exporting role:", r.DisplayName)

			err := exportRole(r, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.ROLES, r.DisplayName)
				log.Printf("Error while exporting role: %s. %s", r.DisplayName, err)
			} else {
				utils.AddToIdentifierMap(utils.ROLES, r.Id, r.DisplayName, utils.EXPORT)
				utils.UpdateSuccessSummary(utils.ROLES, utils.EXPORT)
				log.Println("Role exported successfully:", r.DisplayName)
			}
		}
	}

}

func exportRole(r role, outputDirPath string, formatString string) error {

	roleData, err := utils.GetResourceData(utils.ROLES, r.Id)
	if err != nil {
		return fmt.Errorf("error while getting role: %w", err)
	}
	roleData, err = utils.RemoveResponseFields(roleData, "id")
	if err != nil {
		return fmt.Errorf("error while processing response fields: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, escapeRoleName(r.DisplayName), format)

	roleKeywordMapping := getRoleKeywordMapping(r.DisplayName)
	modifiedRole, err := utils.ProcessExportedData(roleData, exportedFileName, format, roleKeywordMapping, utils.ROLES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedRole, format, utils.ROLES)
	if err != nil {
		return fmt.Errorf("error while serializing role: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
