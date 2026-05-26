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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(parentDir string, formatString string) {

	log.Println("Exporting branding preferences...")
	exportFilePath := filepath.Join(parentDir, utils.BRANDING_PREFERENCES.String())

	if !utils.IsEntitySupportedInVersion(utils.BRANDING_PREFERENCES) || !utils.IsEntitySupportedInOrg(utils.BRANDING_PREFERENCES) || utils.IsResourceTypeExcluded(utils.BRANDING_PREFERENCES) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			log.Println("Error creating branding preferences directory:", err)
			return
		}
	}

	err := exportBrandingPreferences(exportFilePath, formatString)
	if err != nil {
		if utils.IsResourceNotFound(err) {
			log.Println("No branding preferences configured.")
			if utils.TOOL_CONFIGS.AllowDelete {
				utils.RemoveDeletedLocalResources(exportFilePath, []string{})
			}
			return
		}
		utils.UpdateFailureSummary(utils.BRANDING_PREFERENCES, resourceFileName)
		log.Println("Error while exporting branding preferences:", err)
	} else {
		utils.UpdateSuccessSummary(utils.BRANDING_PREFERENCES, utils.EXPORT)
		log.Println("Branding preferences exported successfully.")
	}
}

func exportBrandingPreferences(outputDirPath string, formatString string) error {

	data, err := utils.GetResourceData(utils.BRANDING_PREFERENCES, "")
	if err != nil {
		return err
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, resourceFileName, format)
	keywordMapping := getBrandingPreferencesKeywordMapping()

	modifiedData, err := utils.ProcessExportedData(data, exportedFileName, format, keywordMapping, utils.BRANDING_PREFERENCES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, utils.BRANDING_PREFERENCES)
	if err != nil {
		return fmt.Errorf("error while serializing exported content: %w", err)
	}

	if err := ioutil.WriteFile(exportedFileName, modifiedFile, 0644); err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
