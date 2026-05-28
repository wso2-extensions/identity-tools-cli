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

package brandingPreferences

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(parentDir string) {

	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "Importing branding preferences...")
	importFilePath := filepath.Join(parentDir, utils.BRANDING_PREFERENCES.String())

	if !utils.IsEntitySupportedInVersion(utils.BRANDING_PREFERENCES) || !utils.IsEntitySupportedInOrg(utils.BRANDING_PREFERENCES) || utils.IsResourceTypeExcluded(utils.BRANDING_PREFERENCES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "No branding preferences to import.")
		return
	}

	isDeployed, err := isBrandingPreferencesExist()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.BRANDING_PREFERENCES, "", fmt.Sprintf("Error retrieving deployed branding preferences: %s", err))
		return
	}
	filePath, fileExists, err := getBrandingPreferencesFilePath(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.BRANDING_PREFERENCES, "", fmt.Sprintf("Error reading branding preferences file path: %s", err))
		return
	}

	if !fileExists {
		if utils.TOOL_CONFIGS.AllowDelete && isDeployed {
			removeDeletedDeployedBrandingPreferences()
		}
		return
	}

	err = importBrandingPreferences(filePath, isDeployed)
	if err != nil {
		utils.UpdateFailureSummary(utils.BRANDING_PREFERENCES, resourceFileName)
		utils.PrintLog(utils.LogLevelError, utils.BRANDING_PREFERENCES, "", fmt.Sprintf("Error while importing branding preferences: %s", err))
	}
}

func importBrandingPreferences(filePath string, exists bool) error {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for branding preferences: %w", err)
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error when reading branding preferences file: %w", err)
	}

	keywordMapping := getBrandingPreferencesKeywordMapping()
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	jsonBody, err := utils.PrepareJSONRequestBody([]byte(modifiedFileData), format, utils.BRANDING_PREFERENCES)
	if err != nil {
		return err
	}

	if !exists {
		return createBrandingPreferences(jsonBody)
	}
	return updateBrandingPreferences(jsonBody)
}

func createBrandingPreferences(jsonBody []byte) error {

	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "Creating branding preferences")

	resp, err := utils.SendPostRequest(utils.BRANDING_PREFERENCES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating branding preferences: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.BRANDING_PREFERENCES, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "Created successfully")
	return nil
}

func updateBrandingPreferences(jsonBody []byte) error {

	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "Updating branding preferences")

	resp, err := utils.SendPutRequest(utils.BRANDING_PREFERENCES, "", jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating branding preferences: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.BRANDING_PREFERENCES, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "Updated successfully")
	return nil
}

func removeDeletedDeployedBrandingPreferences() {

	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "Not found locally. Deleting preferences.")

	if err := utils.SendDeleteRequest("", utils.BRANDING_PREFERENCES); err != nil {
		utils.UpdateFailureSummary(utils.BRANDING_PREFERENCES, resourceFileName)
		utils.PrintLog(utils.LogLevelError, utils.BRANDING_PREFERENCES, "", fmt.Sprintf("Error while deleting branding preferences: %s", err))
	} else {
		utils.UpdateSuccessSummary(utils.BRANDING_PREFERENCES, utils.DELETE)
		utils.PrintLog(utils.LogLevelInfo, utils.BRANDING_PREFERENCES, "", "Deleted successfully")
	}
}
