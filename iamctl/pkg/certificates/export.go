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

func ExportAll(exportFilePath string, format string) {

	utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, "", "Exporting certificates...")
	exportFilePath = filepath.Join(exportFilePath, utils.CERTIFICATES.String())

	if utils.ShouldSkip(utils.CERTIFICATES) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(exportFilePath, 0700); err != nil {
			utils.PrintLog(utils.LogLevelError, utils.CERTIFICATES, "", fmt.Sprintf("Error creating certificates directory: %s", err))
			utils.MarkResTypeFailure(utils.CERTIFICATES)
			return
		}
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedAliases := getDeployedCertificateAliases()
			utils.RemoveDeletedLocalResources(exportFilePath, deployedAliases)
		}
	}

	certs, err := getCertificateList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CERTIFICATES, "", fmt.Sprintf("Error retrieving the deployed certificates list: %s", err))
		utils.MarkResTypeFailure(utils.CERTIFICATES)
	} else {
		for _, cert := range certs {
			if !utils.IsResourceExcluded(cert.Alias, utils.TOOL_CONFIGS.CertificateConfigs) {
				utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, cert.Alias, "Exporting")

				err := exportCertificate(cert.Alias, exportFilePath, format)
				if err != nil {
					utils.UpdateFailureSummary(utils.CERTIFICATES, cert.Alias)
					utils.PrintLog(utils.LogLevelError, utils.CERTIFICATES, cert.Alias, fmt.Sprintf("Error while exporting: %s", err))
				} else {
					utils.UpdateSuccessSummary(utils.CERTIFICATES, utils.EXPORT)
					utils.PrintLog(utils.LogLevelInfo, utils.CERTIFICATES, cert.Alias, "Exported successfully")
				}
			}
		}
	}
}

func exportCertificate(alias string, outputDirPath string, formatString string) error {

	certData, err := getEncodedCertificate(alias)
	if err != nil {
		return fmt.Errorf("error while getting certificate data: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, alias, format)

	certKeywordMapping := getCertificateKeywordMapping(alias)
	modifiedCert, err := utils.ProcessExportedData(certData, exportedFileName, format, certKeywordMapping, utils.CERTIFICATES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedCert, format, utils.CERTIFICATES)
	if err != nil {
		return fmt.Errorf("error while serializing certificate: %w", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
