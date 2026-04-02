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

package certificates

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing certificates...")
	if utils.SERVER_CONFIGS.TenantDomain == utils.DEFAULT_TENANT_DOMAIN {
		log.Println("Importing certificates for super tenant not supported.")
		return
	}
	importFilePath := filepath.Join(inputDirPath, utils.CERTIFICATES.String())

	if utils.IsResourceTypeExcluded(utils.CERTIFICATES) {
		return
	}
	var files []os.DirEntry
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No certificates to import.")
		return
	}

	existingCertList, err := getCertificateList()
	if err != nil {
		log.Println("Error retrieving the deployed certificate list: ", err)
		return
	}

	files, err = os.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing certificates: ", err)
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedCertificates(files, existingCertList)
	}

	for _, file := range files {
		certFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(certFilePath)
		alias := fileInfo.ResourceName

		if !utils.IsResourceExcluded(alias, utils.TOOL_CONFIGS.CertificateConfigs) {
			certExists := isCertificateExists(alias, existingCertList)
			err := importCertificate(alias, certExists, certFilePath)
			if err != nil {
				log.Println("Error importing certificate: ", err)
				utils.UpdateFailureSummary(utils.CERTIFICATES, alias)
			}
		}
	}
}

func importCertificate(alias string, certExists bool, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for certificate: %w", err)
	}

	fileBytes, err := os.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for certificate: %w", err)
	}

	certKeywordMapping := getCertificateKeywordMapping(alias)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), certKeywordMapping)

	if !certExists {
		return createCertificate([]byte(modifiedFileData), format, alias)
	}

	log.Printf("Certificate '%s' already exists. Skipping.\n", alias)
	return nil
}

func createCertificate(requestBody []byte, format utils.Format, alias string) error {

	log.Println("Creating new certificate: " + alias)

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.CERTIFICATES)
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.CERTIFICATES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating certificate: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.CERTIFICATES, utils.IMPORT)
	log.Println("Certificate imported successfully.")
	return nil
}

func removeDeletedDeployedCertificates(localFiles []os.DirEntry, deployedCerts []certificate) {

	if len(deployedCerts) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, cert := range deployedCerts {
		if _, existsLocally := localResourceNames[cert.Alias]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(cert.Alias, utils.TOOL_CONFIGS.CertificateConfigs) {
			log.Println("Certificate is excluded from deletion:", cert.Alias)
			continue
		}

		log.Printf("Certificate: %s not found locally. Deleting certificate.\n", cert.Alias)
		if err := utils.SendDeleteRequest(cert.Alias, utils.CERTIFICATES); err != nil {
			utils.UpdateFailureSummary(utils.CERTIFICATES, cert.Alias)
			log.Println("Error deleting certificate:", cert.Alias, err)
		} else {
			utils.UpdateSuccessSummary(utils.CERTIFICATES, utils.DELETE)
		}
	}
}
