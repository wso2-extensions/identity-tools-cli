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

func ExportAll(parentDir string, formatString string) {

	log.Println("Exporting custom texts...")
	exportFilePath := filepath.Join(parentDir, utils.CUSTOM_TEXTS.String())

	if utils.IsResourceTypeExcluded(utils.CUSTOM_TEXTS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	}

	var screensWithLocales []string
	for _, screen := range ScreenList {
		if utils.IsResourceExcluded(screen, utils.TOOL_CONFIGS.CustomTextConfigs) {
			screensWithLocales = append(screensWithLocales, screen)
			continue
		}
		log.Println("Exporting custom text for screen:", screen)
		hadLocales, err := exportCustomTextScreen(screen, exportFilePath, formatString)

		if err != nil {
			utils.UpdateFailureSummary(utils.CUSTOM_TEXTS, screen)
			log.Printf("Error while exporting custom text for screen: %s. %s", screen, err)
		} else {
			if hadLocales {
				screensWithLocales = append(screensWithLocales, screen)
			}
			utils.UpdateSuccessSummary(utils.CUSTOM_TEXTS, utils.EXPORT)
			log.Println("Custom text exported successfully for screen:", screen)
		}
	}

	if utils.TOOL_CONFIGS.AllowDelete {
		utils.RemoveDeletedLocalDirectories(exportFilePath, screensWithLocales)
	}
}

func exportCustomTextScreen(screen, exportFilePath, formatString string) (hadLocales bool, err error) {

	screenDir := filepath.Join(exportFilePath, screen)
	screenDirCreated := false

	format := utils.FormatFromString(formatString)
	keywordMapping := getCustomTextsKeywordMapping(screen)

	var exportedLocales []string
	for _, locale := range LocaleList {
		data, err := utils.GetResourceData(utils.CUSTOM_TEXTS, "",
			utils.WithQueryParams(map[string]string{"screen": screen, "locale": locale}))
		if err != nil {
			if utils.IsResourceNotFound(err) {
				continue
			}
			return false, fmt.Errorf("error while retrieving custom text locale: %s. %w", locale, err)
		}

		if !screenDirCreated {
			if err := os.MkdirAll(screenDir, 0700); err != nil {
				return false, fmt.Errorf("error creating directory for screen: %w", err)
			}
			screenDirCreated = true
		}

		err = exportCustomTextLocale(data, screenDir, locale, format, keywordMapping)
		if err != nil {
			return false, fmt.Errorf("error while exporting custom text locale: %s. %w", locale, err)
		}
		exportedLocales = append(exportedLocales, locale)
	}

	hadLocales = len(exportedLocales) > 0
	if utils.TOOL_CONFIGS.AllowDelete && hadLocales {
		utils.RemoveDeletedLocalResources(screenDir, exportedLocales)
	}
	return hadLocales, nil
}

func exportCustomTextLocale(data interface{}, screenDir, locale string, format utils.Format, keywordMapping map[string]interface{}) error {

	exportedFileName := utils.GetExportedFilePath(screenDir, locale, format)

	modifiedData, err := utils.ProcessExportedData(data, exportedFileName, format, keywordMapping, utils.CUSTOM_TEXTS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, utils.CUSTOM_TEXTS)
	if err != nil {
		return fmt.Errorf("error while serializing exported content: %w", err)
	}
	if err := ioutil.WriteFile(exportedFileName, modifiedFile, 0644); err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}
	return nil
}
