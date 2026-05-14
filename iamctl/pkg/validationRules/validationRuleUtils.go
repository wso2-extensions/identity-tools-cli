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

package validationRules

import (
	"fmt"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

const resourceFileName = "validationRules"

func getValidationRuleKeywordMapping() map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ValidationRuleConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(resourceFileName, utils.KEYWORD_CONFIGS.ValidationRuleConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getValidationRulesFilePath(importDirPath string) (string, error) {

	matches, err := filepath.Glob(filepath.Join(importDirPath, resourceFileName+".*"))
	if err != nil {
		return "", fmt.Errorf("error searching for validation rules file: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("validation rules file not found in %s", importDirPath)
	}
	return matches[0], nil
}

func prepareValidationRulesRequestBody(data []byte, format utils.Format) ([]byte, error) {

	parsed, err := utils.Deserialize(data, format, utils.VALIDATION_RULES)
	if err != nil {
		return nil, fmt.Errorf("error deserializing file: %w", err)
	}
	parsed = utils.ConvertToStringKeyMap(parsed)

	jsonBody, err := utils.Serialize(parsed, utils.FormatJSON, utils.VALIDATION_RULES)
	if err != nil {
		return nil, fmt.Errorf("error serializing to JSON: %w", err)
	}
	return jsonBody, nil
}
