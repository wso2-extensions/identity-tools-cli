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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, "", "Importing script libraries...")
	importFilePath := filepath.Join(inputDirPath, utils.SCRIPT_LIBRARIES.String())

	if !utils.IsEntitySupportedInVersion(utils.SCRIPT_LIBRARIES) || !utils.IsEntitySupportedInOrg(utils.SCRIPT_LIBRARIES) || utils.IsResourceTypeExcluded(utils.SCRIPT_LIBRARIES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, "", "No script libraries to import.")
		return
	}

	existingList, err := getScriptLibraryList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.SCRIPT_LIBRARIES, "", fmt.Sprintf("Error retrieving the deployed script library list: %s", err))
		return
	}

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.SCRIPT_LIBRARIES, "", fmt.Sprintf("Error importing script libraries: %s", err))
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedScriptLibraries(files, existingList)
	}

	for _, file := range files {
		libraryFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(libraryFilePath)
		libraryName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(libraryName, utils.TOOL_CONFIGS.ScriptLibraryConfigs) {
			libraryExists := isScriptLibraryExists(libraryName, existingList)
			err := importScriptLibrary(libraryName, libraryExists, libraryFilePath)
			if err != nil {
				utils.PrintLog(utils.LogLevelError, utils.SCRIPT_LIBRARIES, libraryName, fmt.Sprintf("Error importing script library: %s", err))
				utils.UpdateFailureSummary(utils.SCRIPT_LIBRARIES, libraryName)
			}
		}
	}
}

func importScriptLibrary(libraryName string, libraryExists bool, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for script library: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for script library: %w", err)
	}

	keywordMapping := getScriptLibraryKeywordMapping(libraryName)
	modifiedFileData := []byte(utils.ReplaceKeywords(string(fileBytes), keywordMapping))

	if !libraryExists {
		return createScriptLibrary(libraryName, modifiedFileData, format)
	}
	return updateScriptLibrary(libraryName, modifiedFileData, format)
}

func createScriptLibrary(name string, data []byte, format utils.Format) error {

	utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, name, "Creating new script library")

	body, contentType, err := utils.PrepareMultipartFormBody(data, format, utils.SCRIPT_LIBRARIES)
	if err != nil {
		return fmt.Errorf("error building multipart form: %w", err)
	}

	resp, err := utils.SendPostRequest(utils.SCRIPT_LIBRARIES, body, utils.WithContentType(contentType))
	if err != nil {
		return fmt.Errorf("error when creating script library: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.SCRIPT_LIBRARIES, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, name, "Created successfully")
	return nil
}

func updateScriptLibrary(name string, data []byte, format utils.Format) error {

	utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, name, "Updating script library")

	body, contentType, err := utils.PrepareMultipartFormBody(data, format, utils.SCRIPT_LIBRARIES, "name")
	if err != nil {
		return fmt.Errorf("error building multipart form: %w", err)
	}

	resp, err := utils.SendPutRequest(utils.SCRIPT_LIBRARIES, name, body, utils.WithContentType(contentType))
	if err != nil {
		return fmt.Errorf("error when updating script library: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.SCRIPT_LIBRARIES, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, name, "Updated successfully")
	return nil
}

func removeDeletedDeployedScriptLibraries(localFiles []os.FileInfo, deployedLibraries []scriptLibrary) {

	if len(deployedLibraries) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, library := range deployedLibraries {
		if _, existsLocally := localResourceNames[library.Name]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(library.Name, utils.TOOL_CONFIGS.ScriptLibraryConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, library.Name, "Excluded from deletion.")
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, utils.SCRIPT_LIBRARIES, library.Name, "Not found locally. Deleting library.")
		if err := utils.SendDeleteRequest(library.Name, utils.SCRIPT_LIBRARIES); err != nil {
			utils.UpdateFailureSummary(utils.SCRIPT_LIBRARIES, library.Name)
			utils.PrintLog(utils.LogLevelError, utils.SCRIPT_LIBRARIES, library.Name, fmt.Sprintf("Error deleting script library: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.SCRIPT_LIBRARIES, utils.DELETE)
		}
	}
}
