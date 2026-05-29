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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, "", "Importing certificates...")
	if utils.SERVER_CONFIGS.TenantDomain == utils.DEFAULT_TENANT_DOMAIN {
		utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, "", "Importing certificates for super tenant not supported.")
		return
	}
	importFilePath := filepath.Join(inputDirPath, utils.CERTIFICATES.String())

	if utils.ShouldSkip(utils.CERTIFICATES) {
		return
	}
	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, "", "No certificates to import.")
		return
	}

	existingCertList, err := getCertificateList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CERTIFICATES, "", fmt.Sprintf("Error retrieving the deployed certificate list: %s", err))
		utils.MarkResTypeFailure(utils.CERTIFICATES)
		return
	}

	files, err = ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CERTIFICATES, "", fmt.Sprintf("Error reading certificates directory: %s", err))
		utils.MarkResTypeFailure(utils.CERTIFICATES)
		return
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
				utils.PrintLog(utils.LogLevelError, utils.CERTIFICATES, alias, fmt.Sprintf("Error importing certificate: %s", err))
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

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for certificate: %w", err)
	}

	certKeywordMapping := getCertificateKeywordMapping(alias)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), certKeywordMapping)

	if !certExists {
		return createCertificate([]byte(modifiedFileData), format, alias)
	}

	utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, alias, "Already exists. Skipping.")
	return nil
}

func createCertificate(requestBody []byte, format utils.Format, alias string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, alias, "Creating new certificate")

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
	utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, alias, "Imported successfully")
	return nil
}

func removeDeletedDeployedCertificates(localFiles []os.FileInfo, deployedCerts []certificate) {

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
		if utils.IsResourceExcluded(cert.Alias, utils.TOOL_CONFIGS.CertificateConfigs) || cert.Alias == utils.SERVER_CONFIGS.TenantDomain {
			utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, cert.Alias, "Excluded from deletion.")
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, cert.Alias, "Not found locally. Deleting.")
		if err := utils.SendDeleteRequest(cert.Alias, utils.CERTIFICATES); err != nil {
			utils.UpdateFailureSummary(utils.CERTIFICATES, cert.Alias)
			utils.PrintLog(utils.LogLevelError, utils.CERTIFICATES, cert.Alias, fmt.Sprintf("Error deleting certificate: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.CERTIFICATES, utils.DELETE)
		}
	}
}
