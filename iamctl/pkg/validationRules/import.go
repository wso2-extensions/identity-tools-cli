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
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing validation rules...")
	importFilePath := filepath.Join(inputDirPath, utils.VALIDATION_RULES.String())

	if !utils.IsEntitySupportedInVersion(utils.VALIDATION_RULES) || utils.IsResourceTypeExcluded(utils.VALIDATION_RULES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No validation rules to import.")
		return
	}
	filePath, err := getValidationRulesFilePath(importFilePath)

	if err == nil {
		err = importValidationRules(filePath)
	}
	if err != nil {
		utils.UpdateFailureSummary(utils.VALIDATION_RULES, resourceFileName)
		log.Println("Error importing validation rules:", err)
	}
}

func importValidationRules(filePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported format for validation rules file: %w", err)
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading validation rules file: %w", err)
	}

	keywordMapping := getValidationRuleKeywordMapping()
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	return updateValidationRules([]byte(modifiedFileData), format)
}

func updateValidationRules(requestBody []byte, format utils.Format) error {

	log.Println("Updating validation rules.")

	jsonBody, err := prepareValidationRulesRequestBody(requestBody, format)
	if err != nil {
		return fmt.Errorf("error preparing update request body: %w", err)
	}

	resp, err := utils.SendPutRequest(utils.VALIDATION_RULES, "", jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating validation rules: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.VALIDATION_RULES, utils.UPDATE)
	log.Println("Validation rules updated successfully.")
	return nil
}
