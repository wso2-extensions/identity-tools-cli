/**
* Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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

package claims

import (
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all claim dialects with related claims.
	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, "", "Exporting claims...")
	exportFilePath = filepath.Join(exportFilePath, utils.CLAIMS.String())

	if !utils.IsEntitySupportedInOrg(utils.CLAIMS) || utils.IsResourceTypeExcluded(utils.CLAIMS) {
		return
	}

	claimDialects, err := getClaimDialectsList()
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, utils.CLAIMS, "", fmt.Sprintf("Error creating claims directory: %s", err))
			return
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, getDeployedDialectFileNames(claimDialects))
		}
	}

	// Min version requirement for claims export api is removed. CRUD apis used for all versions
	exportAPIExists := utils.ExportAPIExists(utils.CLAIMS)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CLAIMS, "", fmt.Sprintf("Error while retrieving Claim Dialect list: %s", err))
	} else {
		for _, dialect := range claimDialects {
			if !utils.IsResourceExcluded(dialect.DialectURI, utils.TOOL_CONFIGS.ClaimConfigs) {
				utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialect.DialectURI, "Exporting")

				var err error
				if exportAPIExists {
					err = exportClaimDialect(dialect.Id, dialect.DialectURI, exportFilePath, format)
				} else {
					err = exportClaimDialectWithCRUD(dialect.Id, dialect.DialectURI, exportFilePath, format)
				}

				if err != nil {
					utils.UpdateFailureSummary(utils.CLAIMS, dialect.DialectURI)
					utils.PrintLog(utils.LogLevelError, utils.CLAIMS, dialect.DialectURI, fmt.Sprintf("Error while exporting: %s", err))
				} else {
					utils.UpdateSuccessSummary(utils.CLAIMS, utils.EXPORT)
					utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialect.DialectURI, "Exported successfully")
				}
			}
		}
	}
}

func exportClaimDialect(dialectId, dialectUri, outputDirPath, format string) error {

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

	resp, err := utils.SendExportRequest(dialectId, fileType, utils.CLAIMS, true)
	if err != nil {
		return fmt.Errorf("error while exporting the claim dialect: %s", err)
	}

	var attachmentDetail = resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(attachmentDetail)
	if err != nil {
		return fmt.Errorf("error while parsing the content disposition header: %s", err)
	}

	fileName := params["filename"]
	exportedFileName := filepath.Join(outputDirPath, fileName)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response body when exporting claim dialect: %s. %s", fileName, err)
	}

	claimDialectKeywordMapping := getClaimKeywordMapping(dialectUri)
	modifiedFile, err := utils.ProcessExportedContent(exportedFileName, body, claimDialectKeywordMapping, utils.CLAIMS)
	if err != nil {
		return fmt.Errorf("error while processing the exported content: %s", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing the exported content to file: %w", err)
	}
	return nil
}

func exportClaimDialectWithCRUD(dialectId, dialectUri, outputDirPath, formatString string) error {

	claimDialect, err := getClaimDialect(dialectId)
	if err != nil {
		return fmt.Errorf("error while getting claim dialect: %w", err)
	}

	format := utils.FormatFromString(formatString)
	fileName := formatFileName(dialectUri)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, fileName, format)

	dialectKeywordMapping := getClaimKeywordMapping(dialectUri)
	modifiedDialect, err := utils.ProcessExportedData(claimDialect, exportedFileName, format, dialectKeywordMapping, utils.CLAIMS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedDialect, format, utils.CLAIMS)
	if err != nil {
		return fmt.Errorf("error while serializing claim dialect: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
