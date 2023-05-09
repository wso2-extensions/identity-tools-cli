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

package claims

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all claim dialects with related claims.
	log.Println("Exporting Claim Dialects...")
	exportFilePath = filepath.Join(exportFilePath, utils.CLAIMS)
	os.MkdirAll(exportFilePath, 0700)

	dialects, err := getClaimDialectsList()
	if err != nil {
		log.Println("Error: when getting Claim Dialect IDs.", err)
	} else {
		for _, dialect := range dialects {
			if !utils.IsResourceExcluded(dialect.DialectURI, utils.TOOL_CONFIGS.ClaimDialectConfigs) {
				log.Println("Exporting Claim Dialect: ", dialect.DialectURI)

				err := exportClaimDialect(dialect.Id, exportFilePath, format)
				if err != nil {
					log.Printf("Error while exporting Claim Dialect: %s. %s", dialect.DialectURI, err)
				} else {
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

	var reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/claim-dialects/" + dialectId + "/export"

	var err error
	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("error while creating the request to export Claim Dialect: %s", err)
	}
	req.Header.Set("Content-Type", utils.MEDIA_TYPE_FORM)
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)

	defer req.Body.Close()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error while exporting Claim Dialect: %s", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		var attachmentDetail = resp.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(attachmentDetail)
		if err != nil {
			return fmt.Errorf("error while parsing the content disposition header: %s", err)
		}

		fileName := params["filename"]
		fileName = strings.ReplaceAll(fileName, "http://", "")
		fileName = strings.ReplaceAll(fileName, ":", ".")
		fileName = strings.ReplaceAll(fileName, "/", ".")

		exportedFileName := filepath.Join(outputDirPath, fileName)
		fileInfo := utils.GetFileInfo(exportedFileName)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error while reading the response body when exporting Claim Dialect: %s. %s", fileName, err)
		}

		// Handle Environment Specific Variables.
		claimKeywordMapping := getClaimKeywordMapping(fileInfo.ResourceName)
		modifiedFile := utils.HandleESVs(exportedFileName, body, claimKeywordMapping)

		err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
		if err != nil {
			return fmt.Errorf("error when writing the exported content to file: %w", err)
		}
		return nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error while exporting the Claim Dialect: %s", error)
	}
	return fmt.Errorf("unexpected error while exporting the Claim with status code: %s", strconv.FormatInt(int64(statusCode), 10))
}
