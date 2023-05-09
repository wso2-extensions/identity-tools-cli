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
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing Claim Dialects...")
	importFilePath := filepath.Join(inputDirPath, utils.CLAIMS)

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing Claim Dialects: ", err)
	}

	for _, file := range files {
		claimFilePath := filepath.Join(importFilePath, file.Name())
		dialectName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(dialectName, utils.TOOL_CONFIGS.ClaimDialectConfigs) {
			dialectId, err := getClaimDialectId(claimFilePath)
			if err != nil {
				log.Printf("Invalid file configurations for Claim Dialect: %s. %s", dialectName, err)
			} else {
				err := importClaimDialect(dialectId, claimFilePath)
				if err != nil {
					log.Println("error importing claim dialect: ", err)
				}
			}
		}
	}
}

func importClaimDialect(dialectId string, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for claim dialect: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	claimKeywordMapping := getClaimKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), claimKeywordMapping)

	sendImportRequest(dialectId, importFilePath, modifiedFileData)
	return nil
}

func sendImportRequest(dialectId string, importFilePath string, fileData string) error {

	fileInfo := utils.GetFileInfo(importFilePath)

	var requestMethod, reqUrl string
	if dialectId != "" {
		log.Println("Updating Claim Dialect: " + fileInfo.ResourceName)
		reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/claim-dialects/import"
		requestMethod = "PUT"
	} else {
		log.Println("Creating new Claim Dialect: " + fileInfo.ResourceName)
		reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/claim-dialects/import"
		requestMethod = "POST"
	}

	var buf bytes.Buffer
	var err error
	_, err = io.WriteString(&buf, fileData)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	mime.AddExtensionType(".yml", "application/yaml")
	mime.AddExtensionType(".xml", "application/xml")
	mime.AddExtensionType(".json", "application/json")

	mimeType := mime.TypeByExtension(fileInfo.FileExtension)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", fileInfo.FileName)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	_, err = io.Copy(part, &buf)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	request, err := http.NewRequest(requestMethod, reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error when sending the import request: %s", err)
	}

	statusCode := resp.StatusCode
	if statusCode == 201 {
		log.Println("Claim Dialect created successfully.")
		return nil
	} else if statusCode == 200 {
		log.Println("Claim Dialect updated successfully.")
		return nil
	} else if statusCode == 409 {
		log.Println("A Claim Dialect with the same name already exists. Please adjust the file accordingly.")
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing Claim Dialect: %s", resp.Status)
}

func getClaimDialectId(claimDialectFilePath string) (string, error) {

	fileContent, err := ioutil.ReadFile(claimDialectFilePath)
	if err != nil {
		return "", fmt.Errorf("error when reading the file: %s. %s", claimDialectFilePath, err)
	}
	var claimDialect claimDialect
	err = yaml.Unmarshal(fileContent, &claimDialect)
	if err != nil {
		return "", fmt.Errorf("invalid file content at: %s. %s", claimDialectFilePath, err)
	}

	existingClaimDialectList, err := getClaimDialectsList()
	if err != nil {
		return "", fmt.Errorf("error when retrieving the deployed claim dialect list: %s", err)
	}

	for _, dialect := range existingClaimDialectList {
		if dialect.Id == claimDialect.Id {
			return claimDialect.Id, nil
		}
	}
	return "", nil
}
