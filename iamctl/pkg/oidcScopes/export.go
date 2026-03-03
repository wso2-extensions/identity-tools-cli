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

package oidcScopes

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting OIDC scopes...")
	exportFilePath = filepath.Join(exportFilePath, utils.OIDC_SCOPES.String())

	if utils.IsResourceTypeExcluded(utils.OIDC_SCOPES) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedScopeNames := getDeployedOidcScopeNames()
			utils.RemoveDeletedLocalResources(exportFilePath, deployedScopeNames)
		}
	}

	scopes, err := getOidcScopeList()
	if err != nil {
		log.Println("Error: when exporting OIDC scopes.", err)
	} else {
		for _, scope := range scopes {
			if !utils.IsResourceExcluded(scope.Name, utils.TOOL_CONFIGS.OidcScopeConfigs) {
				log.Println("Exporting OIDC scope: ", scope.Name)

				err := exportOidcScope(scope.Name, exportFilePath, format)
				if err != nil {
					utils.UpdateFailureSummary(utils.OIDC_SCOPES, scope.Name)
					log.Printf("Error while exporting OIDC scope: %s. %s", scope.Name, err)
				} else {
					utils.UpdateSuccessSummary(utils.OIDC_SCOPES, utils.EXPORT)
					log.Println("OIDC scope exported successfully: ", scope.Name)
				}
			}
		}
	}
}

func exportOidcScope(scopeName string, outputDirPath string, formatString string) error {

	scope, err := utils.SendGetRequest(utils.OIDC_SCOPES, scopeName)
	if err != nil {
		return fmt.Errorf("error while getting OIDC scope: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, scopeName, format)

	scopeKeywordMapping := getOidcScopeKeywordMapping(scopeName)
	modifiedScope, err := utils.ProcessExportedData(scope, exportedFileName, format, scopeKeywordMapping, utils.OIDC_SCOPES)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedScope, format, utils.OIDC_SCOPES)
	if err != nil {
		return fmt.Errorf("error while serializing scope: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
