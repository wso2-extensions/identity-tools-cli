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
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all claim dialects with related claims.
	log.Println("Exporting claims...")
	if utils.IsSubOrganization() {
		log.Println("Exporting claims for sub organization not supported.")
		return
	}
	exportFilePath = filepath.Join(exportFilePath, utils.CLAIMS)

	if utils.IsResourceTypeExcluded(utils.CLAIMS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, getDeployedClaimDialectNames())
		}
	}

	claimDialects, err := getClaimDialectsList()
	if err != nil {
		log.Println("Error while retrieving Claim Dialect list.", err)
	} else {
		for _, dialect := range claimDialects {
			if !utils.IsResourceExcluded(dialect.DialectURI, utils.TOOL_CONFIGS.ClaimConfigs) {
				log.Println("Exporting Claim Dialect: ", dialect.DialectURI)

				err := exportClaimDialect(dialect.Id, exportFilePath, format)
				if err != nil {
					utils.UpdateFailureSummary(utils.CLAIMS, dialect.DialectURI)
					log.Printf("Error while exporting Claim Dialect: %s. %s", dialect.DialectURI, err)
				} else {
					utils.UpdateSuccessSummary(utils.CLAIMS, dialect.DialectURI)
					log.Println("Claim Dialect exported successfully: ", dialect.DialectURI)
				}
			}
		}
	}
}

func exportClaimDialect(dialectId string, outputDirPath string, format string) error {

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
	fileInfo := utils.GetFileInfo(exportedFileName)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response body when exporting claim dialect: %s. %s", fileName, err)
	}

	claimDialectKeywordMapping := getClaimKeywordMapping(fileInfo.ResourceName)
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
