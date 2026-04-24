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

package applicationAuthorizedApis

import (
	"fmt"
	"io/ioutil"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAPIs(appId, appName, appsOutputDirPath, formatString string) error {

	if !SupportedInVersion {
		return nil
	}
	outputDirPath := GetOutputDirPath(appsOutputDirPath)

	apiData, err := utils.GetResourceData(utils.APPLICATIONS, appId+"/authorized-apis")
	if err != nil {
		return fmt.Errorf("error fetching authorized APIs: %w", err)
	}
	if err := validateAuthorizedApiReferences(apiData); err != nil {
		return err
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, appName, format)

	keywordMapping := getAuthorizedApisKeywordMapping(appName)
	modifiedData, err := utils.ProcessExportedData(apiData, exportedFileName, format, keywordMapping, utils.APPLICATION_AUTHORIZED_APIS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	fileContent, err := utils.Serialize(modifiedData, format, utils.APPLICATION_AUTHORIZED_APIS)
	if err != nil {
		return fmt.Errorf("error serializing authorized APIs: %w", err)
	}

	if err := ioutil.WriteFile(exportedFileName, fileContent, 0644); err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
