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

func ImportAll(parentDir string) {

	log.Println("Importing branding preferences...")
	importFilePath := filepath.Join(parentDir, utils.BRANDING_PREFERENCES.String())

	if utils.IsResourceTypeExcluded(utils.BRANDING_PREFERENCES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No branding preferences to import.")
		return
	}

	exists, err := isBrandingPreferencesExist()
	if err != nil {
		log.Println("Error retrieving deployed branding preferences: %w", err)
		return
	}

	filePath, err := getBrandingPreferencesFilePath(importFilePath)
	if err == nil {
		err = importBrandingPreferences(filePath, exists)
	}
	if err != nil {
		utils.UpdateFailureSummary(utils.BRANDING_PREFERENCES, resourceFileName)
		log.Println("Error while importing branding preferences:", err)
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

	log.Println("Creating branding preferences.")

	resp, err := utils.SendPostRequest(utils.BRANDING_PREFERENCES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating branding preferences: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.BRANDING_PREFERENCES, utils.IMPORT)
	log.Println("Branding preferences created successfully.")
	return nil
}

func updateBrandingPreferences(jsonBody []byte) error {

	log.Println("Updating branding preferences.")

	resp, err := utils.SendPutRequest(utils.BRANDING_PREFERENCES, "", jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating branding preferences: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.BRANDING_PREFERENCES, utils.UPDATE)
	log.Println("Branding preferences updated successfully.")
	return nil
}
