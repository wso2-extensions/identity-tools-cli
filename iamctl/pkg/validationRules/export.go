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

package validationRules

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	utils.PrintLog(utils.LogLevelInfo, utils.VALIDATION_RULES, "", "Exporting validation rules...")
	exportFilePath = filepath.Join(exportFilePath, utils.VALIDATION_RULES.String())

	if utils.ShouldSkip(utils.VALIDATION_RULES) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, utils.VALIDATION_RULES, "", fmt.Sprintf("Error creating validation rules directory: %s", err))
			utils.MarkResTypeFailure(utils.VALIDATION_RULES)
			return
		}
	}

	err := exportValidationRules(exportFilePath, format)
	if err != nil {
		utils.UpdateFailureSummary(utils.VALIDATION_RULES, resourceFileName)
		utils.PrintLog(utils.LogLevelError, utils.VALIDATION_RULES, "", fmt.Sprintf("Error while exporting validation rules: %s", err))
	} else {
		utils.UpdateSuccessSummary(utils.VALIDATION_RULES, utils.EXPORT)
		utils.PrintLog(utils.LogLevelInfo, utils.VALIDATION_RULES, "", "Exported successfully")
	}
}

func exportValidationRules(outputDirPath string, formatString string) error {

	rules, err := utils.GetResourceData(utils.VALIDATION_RULES, "")
	if err != nil {
		return fmt.Errorf("error while getting validation rules: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, resourceFileName, format)

	keywordMapping := getValidationRuleKeywordMapping()
	modifiedRules, err := utils.ProcessExportedData(rules, exportedFileName, format, keywordMapping, utils.VALIDATION_RULES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedRules, format, utils.VALIDATION_RULES)
	if err != nil {
		return fmt.Errorf("error while serializing validation rules: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
