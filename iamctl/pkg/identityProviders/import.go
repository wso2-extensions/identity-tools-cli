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

package identityproviders

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
	"os"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing identity providers...")
	importFilePath := filepath.Join(inputDirPath, utils.IDENTITY_PROVIDERS)

	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No identity providers to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing identity providers: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedIdpList, err := getIdpList()
			if err != nil {
				log.Println("Error retrieving deployed identity providers: ", err)
			} else {
				removeDeletedDeployedIdps(files, deployedIdpList)
			}
		}

	}

	for _, file := range files {
		idpFilePath := filepath.Join(importFilePath, file.Name())
		idpName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(idpName, utils.TOOL_CONFIGS.IdpConfigs) {
			var idpId string
			var err error
			if idpName == utils.RESIDENT_IDP_NAME {
				idpId = utils.RESIDENT_IDP_NAME
			} else {
				idpId, err = getIdpId(idpFilePath, idpName)
			}

			if err != nil {
				log.Printf("Invalid file configurations for identity provider: %s. %s", idpName, err)
			} else {
				err := importIdp(idpId, idpFilePath)
				if err != nil {
					log.Println("Error importing identity provider: ", err)
				}
			}
		}
	}
}

func importIdp(idpId string, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for identity provider: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	idpKeywordMapping := getIdpKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), idpKeywordMapping)

	sendImportRequest(idpId, importFilePath, modifiedFileData)
	return nil
}

func sendImportRequest(idpId string, importFilePath string, fileData string) error {

	fileInfo := utils.GetFileInfo(importFilePath)

	var requestMethod, reqUrl string
	if idpId != "" {
		log.Println("Updating IdP: " + fileInfo.ResourceName)
		reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/identity-providers/" + idpId + "/import"
		requestMethod = "PUT"
	} else {
		log.Println("Creating new IdP: " + fileInfo.ResourceName)
		reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/identity-providers/import"
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
		log.Println("Identity provider created successfully.")
		return nil
	} else if statusCode == 200 {
		log.Println("Identity provider updated successfully.")
		return nil
	} else if statusCode == 409 {
		log.Println("An identity provider with the same name already exists. Please rename the file accordingly.")
		return importIdp(idpId, importFilePath)
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing identity provider: %s", resp.Status)
}

func getIdpId(idpFilePath string, idpName string) (string, error) {

	fileContent, err := ioutil.ReadFile(idpFilePath)
	if err != nil {
		return "", fmt.Errorf("error when reading the file for idp: %s. %s", idpName, err)
	}
	var idpConfig idpConfig
	err = yaml.Unmarshal(fileContent, &idpConfig)
	if err != nil {
		return "", fmt.Errorf("invalid file content for idp: %s. %s", idpName, err)
	}
	existingIdpList, err := getIdpList()
	if err != nil {
		return "", fmt.Errorf("error when retrieving the deployed idp list: %s", err)
	}

	for _, idp := range existingIdpList {
		if idp.Name == idpConfig.IdentityProviderName {
			return idp.Id, nil
		}
	}
	return "", nil
}

func removeDeletedDeployedIdps(localFiles []os.FileInfo, deployedIdps []identityProvider) {

	// Remove deployed identity providers that do not exist locally.
deployedResourcess:
	for _, idp := range deployedIdps {
		for _, file := range localFiles {
			if idp.Name == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedResourcess
			}
		}
		log.Println("Identity provider not found locally. Deleting idp: ", idp.Name)
		err := utils.DeleteResource(idp.Id, "identity-providers")
		if err != nil {
			log.Println("Error deleting identity provider: ", err)
		}
	}
}
