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

package flows

import (
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var flowTypes = map[string]string{
	"Registration":            "REGISTRATION",
	"PasswordRecovery":        "PASSWORD_RECOVERY",
	"InvitedUserRegistration": "INVITED_USER_REGISTRATION",
}

func getFlowKeywordMapping(flowType string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.FlowConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(flowType, utils.KEYWORD_CONFIGS.FlowConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}
