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

package scriptLibraries

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting script libraries...")
	exportFilePath = filepath.Join(exportFilePath, utils.SCRIPT_LIBRARIES.String())

	if utils.IsResourceTypeExcluded(utils.SCRIPT_LIBRARIES) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedNames := getDeployedScriptLibraryNames()
			utils.RemoveDeletedLocalResources(exportFilePath, deployedNames)
		}
	}

	libraries, err := getScriptLibraryList()
	if err != nil {
		log.Println("Error: when exporting script libraries.", err)
	} else {
		for _, library := range libraries {
			if !utils.IsResourceExcluded(library.Name, utils.TOOL_CONFIGS.ScriptLibraryConfigs) {
				log.Println("Exporting script library: ", library.Name)

				err := exportScriptLibrary(library.Name, exportFilePath, format)
				if err != nil {
					utils.UpdateFailureSummary(utils.SCRIPT_LIBRARIES, library.Name)
					log.Printf("Error while exporting script library: %s. %s", library.Name, err)
				} else {
					utils.UpdateSuccessSummary(utils.SCRIPT_LIBRARIES, utils.EXPORT)
					log.Println("Script library exported successfully: ", library.Name)
				}
			}
		}
	}
}

func exportScriptLibrary(libraryName string, outputDirPath string, formatString string) error {

	libraryData, err := getScriptLibraryData(libraryName)
	if err != nil {
		return fmt.Errorf("error while getting script library: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, libraryName, format)

	keywordMapping := getScriptLibraryKeywordMapping(libraryName)
	modifiedData, err := utils.ProcessExportedData(libraryData, exportedFileName, format, keywordMapping, utils.SCRIPT_LIBRARIES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedData, format, utils.SCRIPT_LIBRARIES)
	if err != nil {
		return fmt.Errorf("error while serializing script library: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
