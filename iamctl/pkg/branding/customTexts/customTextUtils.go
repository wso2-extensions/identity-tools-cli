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

package customTexts

import (
	"fmt"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

const dotPlaceholder = "__DOT__"

var ScreenList = []string{"common", "login", "sms-otp", "email-otp", "totp", "push-auth", "sign-up", "password-recovery", "password-reset", "password-reset-success", "email-link-expiry", "username-recovery-claim", "username-recovery-channel-selection", "username-recovery-success"}
var LocaleList = []string{"en-US", "de-DE", "es-ES", "fr-FR", "ja-JP", "pt-BR", "pt-PT", "zh-CN"}

func getCustomTextList() (map[string]map[string]struct{}, error) {

	deployedTexts := make(map[string]map[string]struct{})

	for _, screen := range ScreenList {
		for _, locale := range LocaleList {
			_, err := getCustomText(screen, locale)
			if err == nil {
				if deployedTexts[screen] == nil {
					deployedTexts[screen] = make(map[string]struct{})
				}
				deployedTexts[screen][locale] = struct{}{}
			} else if !utils.IsResourceNotFound(err) {
				return nil, fmt.Errorf("error retrieving locale %s of screen %s: %w", locale, screen, err)
			}
		}
	}
	return deployedTexts, nil
}

func getCustomTextsKeywordMapping(screen string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.CustomTextConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(screen, utils.KEYWORD_CONFIGS.CustomTextConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getCustomText(screen, locale string) (interface{}, error) {

	return utils.GetResourceData(utils.CUSTOM_TEXTS, "",
		utils.WithQueryParams(map[string]string{"screen": screen, "locale": locale}))
}

func deleteCustomText(screen, locale string) error {

	return utils.SendDeleteRequest("", utils.CUSTOM_TEXTS,
		utils.WithQueryParams(map[string]string{"screen": screen, "locale": locale}))
}

func preprocessCustomTextKeys(data interface{}) (interface{}, error) {

	data = utils.ConvertToStringKeyMap(data)

	textMap, ok := data.(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for custom text data")
	}
	preference, ok := textMap["preference"].(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for preferences")
	}
	texts, ok := preference["text"].(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for preference texts")
	}

	encoded := make(map[string]interface{}, len(texts))
	for k, v := range texts {
		encoded[strings.ReplaceAll(k, ".", dotPlaceholder)] = v
	}
	preference["text"] = encoded

	return textMap, nil
}

func postprocessCustomTextKeys(data interface{}) (interface{}, error) {

	textMap, ok := data.(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for custom text data")
	}
	preference, ok := textMap["preference"].(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for preferences")
	}
	texts, ok := preference["text"].(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for preference texts")
	}

	decoded := make(map[string]interface{}, len(texts))
	for k, v := range texts {
		decoded[strings.ReplaceAll(k, dotPlaceholder, ".")] = v
	}
	preference["text"] = decoded

	return textMap, nil
}

func init() {

	utils.DataPreprocessFuncs[utils.CUSTOM_TEXTS] = preprocessCustomTextKeys
}
