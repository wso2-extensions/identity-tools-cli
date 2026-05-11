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

package brandingPreferences

import (
	"fmt"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

const resourceFileName = "brandingPreferences"

func getBrandingPreferencesKeywordMapping() map[string]interface{} {

	if utils.KEYWORD_CONFIGS.BrandingPreferenceConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(resourceFileName,
			utils.KEYWORD_CONFIGS.BrandingPreferenceConfigs)
	}
	return nil
}

func isBrandingPreferencesExist() (bool, error) {

	_, err := utils.SendGetRequest(utils.BRANDING_PREFERENCES, "")

	if utils.IsResourceNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func getBrandingPreferencesFilePath(importDirPath string) (string, error) {

	matches, err := filepath.Glob(filepath.Join(importDirPath, resourceFileName+".*"))
	if err != nil {
		return "", fmt.Errorf("error searching for branding preferences file: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("branding preferences file not found")
	}
	return matches[0], nil
}
