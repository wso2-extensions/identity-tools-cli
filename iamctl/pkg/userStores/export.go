/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

package userstores

import (
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all userstores to the UserStores folder.
	utils.PrintLog(utils.LogLevelInfo, utils.USERSTORES, "", "Exporting user stores...")
	exportFilePath = filepath.Join(exportFilePath, utils.USERSTORES.String())

	if utils.IsResourceTypeExcluded(utils.USERSTORES) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, utils.USERSTORES, "", fmt.Sprintf("Error creating user stores directory: %s", err))
			utils.MarkResTypeFailure(utils.USERSTORES)
			return
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, getDeployedUserstoreNames())
		}
	}

	exportAPIExists := utils.ExportAPIExists(utils.USERSTORES)
	userstores, err := getUserStoreList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.USERSTORES, "", fmt.Sprintf("Error retrieving the deployed user stores list: %s", err))
		utils.MarkResTypeFailure(utils.USERSTORES)
	} else {
		if !utils.AreSecretsExcluded(utils.TOOL_CONFIGS.UserStoreConfigs) {
			utils.PrintLog(utils.LogLevelWarn, utils.USERSTORES, "", "Secrets exclusion cannot be disabled for user stores. All secrets will be masked.")
		}
		for _, userstore := range userstores {
			if !utils.IsResourceExcluded(userstore.Name, utils.TOOL_CONFIGS.UserStoreConfigs) {
				utils.PrintLog(utils.LogLevelInfo, utils.USERSTORES, userstore.Name, "Exporting")

				if exportAPIExists {
					err = exportUserStore(userstore.Id, exportFilePath, format)
				} else {
					err = exportUserStoreWithCRUD(userstore.Id, userstore.Name, exportFilePath, format)
				}

				if err != nil {
					utils.UpdateFailureSummary(utils.USERSTORES, userstore.Name)
					utils.PrintLog(utils.LogLevelError, utils.USERSTORES, userstore.Name, fmt.Sprintf("Error while exporting: %s", err))
				} else {
					utils.UpdateSuccessSummary(utils.USERSTORES, utils.EXPORT)
					utils.PrintLog(utils.LogLevelInfo, utils.USERSTORES, userstore.Name, "Exported successfully")
				}
			}
		}
	}
	if (utils.IsResourceTypeExcluded(utils.CLAIMS) || utils.IsResourceExcluded(utils.LOCAL_CLAIM_DIALECT_URI, utils.TOOL_CONFIGS.ClaimConfigs)) && exportAPIExists {
		utils.PrintLog(utils.LogLevelWarn, utils.USERSTORES, "", "Local claim dialect is excluded from export. Export local claims to persist claim attribute mappings of user stores.")
	}
}

func exportUserStore(userStoreId string, outputDirPath string, format string) error {

	var fileType string
	// TODO: Extend support for json and xml formats.
	switch format {
	case "json":
		fileType = utils.MEDIA_TYPE_JSON
	case "xml":
		fileType = utils.MEDIA_TYPE_XML
	default:
		fileType = utils.MEDIA_TYPE_YAML
	}

	resp, err := utils.SendExportRequest(userStoreId, fileType, utils.USERSTORES, true)
	if err != nil {
		return fmt.Errorf("error while exporting the user store: %s", err)
	}

	var attachmentDetail = resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(attachmentDetail)
	if err != nil {
		return fmt.Errorf("error while parsing the content disposition header: %s", err)
	}

	fileName := params["filename"]
	exportedFileName := filepath.Join(outputDirPath, fileName)
	fileInfo := utils.GetFileInfo(exportedFileName)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response body when exporting user store: %s. %s", fileName, err)
	}

	// Use the common mask for senstive data.
	modifiedBody := []byte(strings.ReplaceAll(string(body), USERSTORE_SECRET_MASK, utils.SENSITIVE_FIELD_MASK))

	userStoreKeywordMapping := getUserStoreKeywordMapping(fileInfo.ResourceName)
	modifiedFile, err := utils.ProcessExportedContent(exportedFileName, modifiedBody, userStoreKeywordMapping, utils.USERSTORES)
	if err != nil {
		return fmt.Errorf("error while processing the exported content: %s", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing the exported content to file: %w", err)
	}
	return nil
}

func exportUserStoreWithCRUD(userStoreId, userStoreName, outputDirPath, formatString string) error {

	userStore, err := getUserStore(userStoreId)
	if err != nil {
		return fmt.Errorf("error while getting user store: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, userStoreName, format)

	storeKeywordMapping := getUserStoreKeywordMapping(userStoreName)
	modifiedStore, err := utils.ProcessExportedData(userStore, exportedFileName, format, storeKeywordMapping, utils.USERSTORES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedStore, format, utils.USERSTORES)
	if err != nil {
		return fmt.Errorf("error while serializing user store: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}

func getUserStore(userStoreId string) (interface{}, error) {

	resp, err := utils.SendGetRequest(utils.USERSTORES, userStoreId)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving user store with id: %s. %s", userStoreId, err)
	}

	modifiedResp := []byte(strings.ReplaceAll(string(resp), USERSTORE_SECRET_MASK, utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES))

	userstore, err := utils.Deserialize(modifiedResp, utils.FormatJSON, utils.USERSTORES)
	if err != nil {
		return nil, fmt.Errorf("error deserializing JSON response: %w", err)
	}

	return userstore, nil
}
