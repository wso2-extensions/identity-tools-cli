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
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing identity providers...")
	importFilePath := filepath.Join(inputDirPath, configs.IDENTITY_PROVIDERS)
	if !utils.IsEntitySupportedInVersion(configs.IDENTITY_PROVIDERS) {
		return
	}

	if utils.IsResourceTypeExcluded(configs.IDENTITY_PROVIDERS) {
		return
	}
	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No identity providers to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing identity providers: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			removeDeletedDeployedIdps(files)
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

	if idpId == "" {
		return importIdentityProvider(importFilePath, modifiedFileData, fileInfo)
	}
	return updateIdentityProvider(idpId, importFilePath, modifiedFileData, fileInfo)
}

func importIdentityProvider(importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Creating new identity provider: " + fileInfo.ResourceName)
	err := utils.SendImportRequest(importFilePath, modifiedFileData, configs.IDENTITY_PROVIDERS)
	if err != nil {
		utils.UpdateFailureSummary(configs.IDENTITY_PROVIDERS, fileInfo.ResourceName)
		return fmt.Errorf("error when importing identity provider: %s", err)
	}
	utils.UpdateSuccessSummary(configs.IDENTITY_PROVIDERS, utils.IMPORT)
	log.Println("Identity provider imported successfully.")
	return nil
}

func updateIdentityProvider(idpId string, importFilePath string, modifiedFileData string, fileInfo utils.FileInfo) error {

	log.Println("Updating identity provider: " + fileInfo.ResourceName)
	err := utils.SendUpdateRequest(idpId, importFilePath, modifiedFileData, configs.IDENTITY_PROVIDERS)
	if err != nil {
		utils.UpdateFailureSummary(configs.IDENTITY_PROVIDERS, fileInfo.ResourceName)
		return fmt.Errorf("error when updating identity provider: %s", err)
	}
	utils.UpdateSuccessSummary(configs.IDENTITY_PROVIDERS, utils.UPDATE)
	log.Println("Identity provider updated successfully.")
	return nil
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

func removeDeletedDeployedIdps(localFiles []os.FileInfo) {

	// Remove deployed identity providers that do not exist locally.
	deployedIdps, err := getIdpList()
	if err != nil {
		log.Println("Error retrieving deployed identity providers: ", err)
		return
	}
deployedResourcess:
	for _, idp := range deployedIdps {
		for _, file := range localFiles {
			if idp.Name == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedResourcess
			}
		}
		if utils.IsResourceExcluded(idp.Name, utils.TOOL_CONFIGS.ApplicationConfigs) || idp.Name == utils.RESIDENT_IDP_NAME {
			log.Println("Identity provider is excluded from deletion: ", idp.Name)
			continue
		}
		log.Printf("Identity provider: %s not found locally. Deleting idp.\n", idp.Name)
		err := utils.SendDeleteRequest(idp.Id, configs.IDENTITY_PROVIDERS)
		if err != nil {
			utils.UpdateFailureSummary(configs.IDENTITY_PROVIDERS, idp.Name)
			log.Println("Error deleting idp: ", idp.Name, err)
		}
		utils.UpdateSuccessSummary(configs.IDENTITY_PROVIDERS, utils.DELETE)
	}
}
