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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing identity providers...")
	importFilePath := filepath.Join(inputDirPath, utils.IDENTITY_PROVIDERS.String())
	exportAPIExists := utils.ExportAPIExists(utils.IDENTITY_PROVIDERS)

	if utils.IsResourceTypeExcluded(utils.IDENTITY_PROVIDERS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No identity providers to import.")
		return
	}

	existingIdpList, err := getIdpList()
	if err != nil {
		log.Printf("error when retrieving the deployed identity provider list. %s", err)
	}

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing identity providers: ", err)
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedIdps(files, existingIdpList)
	}

	for _, file := range files {
		idpFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(idpFilePath)
		idpName := fileInfo.ResourceName

		if !utils.IsResourceExcluded(idpName, utils.TOOL_CONFIGS.IdpConfigs) {
			var idpId string
			if idpName == utils.RESIDENT_IDP_NAME {
				if !exportAPIExists {
					continue
				}
				idpId = utils.RESIDENT_IDP_NAME
			} else {
				idpId = getIdpId(idpName, existingIdpList)
			}

			err := importIdp(idpId, idpName, idpFilePath, exportAPIExists)
			if err != nil {
				log.Println("Error importing identity provider: ", err)
				utils.UpdateFailureSummary(utils.IDENTITY_PROVIDERS, idpName)
			}
		}
	}
}

func importIdp(idpId string, idpName string, importFilePath string, exportAPIExists bool) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for identity provider: %s", err)
	}

	idpKeywordMapping := getIdpKeywordMapping(idpName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), idpKeywordMapping)

	if exportAPIExists {
		if idpId == "" {
			return importIdentityProvider(idpName, importFilePath, modifiedFileData)
		}
		return updateIdentityProvider(idpId, idpName, importFilePath, modifiedFileData)
	}

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for identity provider: %w", err)
	}

	if idpId == "" {
		return createIdpWithCRUD(idpName, []byte(modifiedFileData), format)
	}
	return updateIdpWithCRUD(idpId, idpName, []byte(modifiedFileData), format)
}

func importIdentityProvider(idpName, importFilePath, modifiedFileData string) error {

	log.Println("Creating new identity provider: " + idpName)
	err := utils.SendImportRequest(importFilePath, modifiedFileData, utils.IDENTITY_PROVIDERS)
	if err != nil {
		return fmt.Errorf("error when importing identity provider: %s", err)
	}
	utils.UpdateSuccessSummary(utils.IDENTITY_PROVIDERS, utils.IMPORT)
	log.Println("Identity provider imported successfully.")
	return nil
}

func updateIdentityProvider(idpId, idpName, importFilePath, modifiedFileData string) error {

	log.Println("Updating identity provider: " + idpName)
	err := utils.SendUpdateRequest(idpId, importFilePath, modifiedFileData, utils.IDENTITY_PROVIDERS)
	if err != nil {
		return fmt.Errorf("error when updating identity provider: %s", err)
	}
	utils.UpdateSuccessSummary(utils.IDENTITY_PROVIDERS, utils.UPDATE)
	log.Println("Identity provider updated successfully.")
	return nil
}

func createIdpWithCRUD(idpName string, requestBody []byte, format utils.Format) error {

	log.Println("Creating new identity provider: " + idpName)

	idpMap, err := utils.DeserializeToMap(requestBody, format, utils.IDENTITY_PROVIDERS)
	if err != nil {
		return fmt.Errorf("error deserializing identity provider: %w", err)
	}
	newIdpId, err := createIdp(idpMap)
	if err != nil {
		return fmt.Errorf("error creating identity provider: %w", err)
	}

	isEnabled, exists := idpMap["isEnabled"]
	if exists {
		err := patchIdpIsEnabled(newIdpId, isEnabled)
		if err != nil {
			return fmt.Errorf("error setting isEnabled for identity provider: %w", err)
		}
	}

	utils.UpdateSuccessSummary(utils.IDENTITY_PROVIDERS, utils.IMPORT)
	log.Println("Identity provider imported successfully.")
	return nil
}

func updateIdpWithCRUD(idpId, idpName string, requestBody []byte, format utils.Format) error {

	return fmt.Errorf("not implemented")
}

func createIdp(idpMap map[string]interface{}) (string, error) {

	reqBody, err := createPostRequestBody(idpMap)
	if err != nil {
		return "", err
	}

	resp, err := utils.SendPostRequest(utils.IDENTITY_PROVIDERS, reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in create identity provider response")
	}
	return path.Base(location), nil
}

func patchIdpIsEnabled(idpId string, isEnabled interface{}) error {

	enabled, ok := isEnabled.(bool)
	if !ok {
		return fmt.Errorf("invalid value for isEnabled field: %v", isEnabled)
	}
	if enabled {
		return nil
	}

	patchBody := []map[string]interface{}{{
		"operation": "REPLACE",
		"path":      "/isEnabled",
		"value":     false,
	}}
	return patchIdp(idpId, patchBody)
}

func removeDeletedDeployedIdps(localFiles []os.FileInfo, deployedIdps []identityProvider) {

deployedResourcess:
	for _, idp := range deployedIdps {
		for _, file := range localFiles {
			if idp.Name == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedResourcess
			}
		}
		if utils.IsResourceExcluded(idp.Name, utils.TOOL_CONFIGS.IdpConfigs) || idp.Name == utils.RESIDENT_IDP_NAME {
			log.Println("Identity provider is excluded from deletion: ", idp.Name)
			continue
		}
		log.Printf("Identity provider: %s not found locally. Deleting idp.\n", idp.Name)
		err := utils.SendDeleteRequest(idp.Id, utils.IDENTITY_PROVIDERS)
		if err != nil {
			utils.UpdateFailureSummary(utils.IDENTITY_PROVIDERS, idp.Name)
			log.Println("Error deleting idp: ", idp.Name, err)
		}
		utils.UpdateSuccessSummary(utils.IDENTITY_PROVIDERS, utils.DELETE)
	}
}
