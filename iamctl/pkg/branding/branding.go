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

package branding

import (
	"path/filepath"

	brandingPreferences "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/branding/brandingPreferences"
	customTexts "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/branding/customTexts"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING, "", "Exporting branding...")
	exportFilePath = filepath.Join(exportFilePath, utils.BRANDING.String())

	brandingPreferences.ExportAll(exportFilePath, format)
	customTexts.ExportAll(exportFilePath, format)
}

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.BRANDING, "", "Importing branding...")
	inputDirPath = filepath.Join(inputDirPath, utils.BRANDING.String())

	brandingPreferences.ImportAll(inputDirPath)
	customTexts.ImportAll(inputDirPath)
}
