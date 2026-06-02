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
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	utils.PrintLog(utils.LogLevelInfo, utils.ROLES, "", "Exporting roles...")
	exportFilePath = filepath.Join(exportFilePath, utils.ROLES.String())
	setRolesV2ApiExists()

	if utils.ShouldSkip(utils.ROLES) {
		return
	}
	roles, err := GetRoleList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.ROLES, "", fmt.Sprintf("Error retrieving the deployed roles list: %s", err))
		utils.MarkResTypeFailure(utils.ROLES)
		return
	}

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, utils.ROLES, "", fmt.Sprintf("Error creating roles directory: %s", err))
			utils.MarkResTypeFailure(utils.ROLES)
			return
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedRoleNames := getDeployedRoleLocalFileNames(roles)
			utils.RemoveDeletedLocalResources(exportFilePath, deployedRoleNames)
		}
	}

	for _, r := range roles {
		if !utils.IsResourceExcluded(r.DisplayName, utils.TOOL_CONFIGS.RoleConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.ROLES, r.DisplayName, "Exporting")

			err := exportRole(r, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.ROLES, r.DisplayName)
				utils.PrintLog(utils.LogLevelError, utils.ROLES, r.DisplayName, fmt.Sprintf("Error while exporting: %s", err))
			} else {
				utils.AddToIdentifierMap(utils.ROLES, r.Id, r.DisplayName, utils.EXPORT)
				utils.UpdateSuccessSummary(utils.ROLES, utils.EXPORT)
				utils.PrintLog(utils.LogLevelInfo, utils.ROLES, r.DisplayName, "Exported successfully")
			}
		}
	}

}

func exportRole(r role, outputDirPath string, formatString string) error {

	roleData, err := utils.GetResourceData(utils.ROLES, r.Id)
	if err != nil {
		return fmt.Errorf("error while getting role: %w", err)
	}
	if utils.RolesV2ApiExists {
		roleData, err = processExportedRole(roleData)
		if err != nil {
			return fmt.Errorf("error while processing role permissions: %w", err)
		}
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
