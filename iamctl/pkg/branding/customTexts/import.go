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

package customTexts

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(parentDir string) {

	log.Println("Importing custom texts...")
	importFilePath := filepath.Join(parentDir, utils.CUSTOM_TEXTS.String())

	if !utils.IsEntitySupportedInVersion(utils.CUSTOM_TEXTS) || utils.IsResourceTypeExcluded(utils.CUSTOM_TEXTS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No custom texts to import.")
		return
	}

	deployedTexts, err := getCustomTextList()
	if err != nil {
		log.Printf("Error while retrieving deployed custom text list: %s", err)
		return
	}
	localScreenDirs, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Printf("Error reading custom texts directory: %s", err)
		return
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedScreens(localScreenDirs, deployedTexts)
	}

	for _, entry := range localScreenDirs {
		if !entry.IsDir() {
			continue
		}
		screen := entry.Name()
		screenDir := filepath.Join(importFilePath, screen)

		if !utils.IsResourceExcluded(screen, utils.TOOL_CONFIGS.CustomTextConfigs) {
			if err := importCustomTextScreen(screen, screenDir, deployedTexts[screen]); err != nil {
				utils.UpdateFailureSummary(utils.CUSTOM_TEXTS, screen)
				log.Printf("Error while importing custom text for screen: %s. %s", screen, err)
			}
		}
	}
}

func importCustomTextScreen(screen, screenDir string, deployedLocales map[string]struct{}) error {

	if len(deployedLocales) == 0 {
		log.Printf("Importing custom text for screen: %s", screen)
	} else {
		log.Printf("Updating custom text for screen: %s", screen)
	}

	localFiles, err := ioutil.ReadDir(screenDir)
	if err != nil {
		return fmt.Errorf("error reading local custom text files: %w", err)
	}
	keywordMapping := getCustomTextsKeywordMapping(screen)

	if utils.TOOL_CONFIGS.AllowDelete {
		if err := removeDeletedDeployedLocales(screen, localFiles, deployedLocales); err != nil {
			return fmt.Errorf("error removing deleted deployed locales: %w", err)
		}
	}

	for _, file := range localFiles {
		filePath := filepath.Join(screenDir, file.Name())
		locale := utils.GetFileInfo(file.Name()).ResourceName

		_, srvExists := deployedLocales[locale]
		if err := importCustomTextLocale(filePath, srvExists, keywordMapping); err != nil {
			return fmt.Errorf("error importing locale %s: %w", locale, err)
		}
	}

	if len(deployedLocales) == 0 {
		utils.UpdateSuccessSummary(utils.CUSTOM_TEXTS, utils.IMPORT)
		log.Printf("Custom text for screen imported successfully")
	} else {
		utils.UpdateSuccessSummary(utils.CUSTOM_TEXTS, utils.UPDATE)
		log.Printf("Custom text for screen updated successfully")
	}
	return nil
}

func importCustomTextLocale(filePath string, srvExists bool, keywordMapping map[string]interface{}) error {

	format, err := utils.FormatFromExtension(filepath.Ext(filePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for custom text: %w", err)
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for custom text: %w", err)
	}

	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	if !srvExists {
		return createLocale([]byte(modifiedFileData), format)
	}
	return updateLocale([]byte(modifiedFileData), format)
}

func createLocale(requestBody []byte, format utils.Format) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.CUSTOM_TEXTS)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.CUSTOM_TEXTS, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating locale: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func updateLocale(requestBody []byte, format utils.Format) error {

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.CUSTOM_TEXTS)
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(utils.CUSTOM_TEXTS, "", jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating locale: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func removeDeletedDeployedScreens(localScreenDirs []os.FileInfo, deployedTexts map[string]map[string]struct{}) {

	localScreenNames := make(map[string]struct{})
	for _, dir := range localScreenDirs {
		if dir.IsDir() {
			localScreenNames[dir.Name()] = struct{}{}
		}
	}

	for screen, locales := range deployedTexts {
		if _, existsLocally := localScreenNames[screen]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(screen, utils.TOOL_CONFIGS.CustomTextConfigs) {
			log.Printf("Custom text for screen %s is excluded from deletion.", screen)
			continue
		}

		log.Printf("Custom text for screen %s not found locally. Deleting all locales", screen)
		for locale := range locales {
			if err := deleteCustomText(screen, locale); err != nil {
				log.Printf("Error deleting locale %s of screen %s: %s", locale, screen, err)
				utils.UpdateFailureSummary(utils.CUSTOM_TEXTS, screen+"/"+locale)
				continue
			} else {
				utils.UpdateSuccessSummary(utils.CUSTOM_TEXTS, utils.DELETE)
			}
		}
	}
}

func removeDeletedDeployedLocales(screen string, localFiles []os.FileInfo, deployedLocales map[string]struct{}) error {

	localLocales := make(map[string]struct{})
	for _, file := range localFiles {
		locale := utils.GetFileInfo(file.Name()).ResourceName
		localLocales[locale] = struct{}{}
	}

	for locale := range deployedLocales {
		if _, existsLocally := localLocales[locale]; existsLocally {
			continue
		}
		log.Printf("Locale %s not found locally. Deleting locale.", locale)
		if err := deleteCustomText(screen, locale); err != nil {
			return fmt.Errorf("error deleting locale: %s. %w", locale, err)
		}
	}
	return nil
}
