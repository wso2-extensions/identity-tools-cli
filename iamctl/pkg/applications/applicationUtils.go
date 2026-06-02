/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

package applications

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/claims"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/roles"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type inboundProtocolRef struct {
	Self string `json:"self"`
}

type Application struct {
	Id               string               `json:"id"`
	Name             string               `json:"name"`
	InboundProtocols []inboundProtocolRef `json:"inboundProtocols"`
}

var protocolPathToKey = map[string]string{
	"saml":        "saml",
	"oidc":        "oidc",
	"passive-sts": "passiveSts",
	"ws-trust":    "wsTrust",
}

var unsupportedInboundProtocols = map[string]struct{}{
	"kerberos": {},
	"openid":   {},
}

type AuthConfig struct {
	InboundAuthenticationConfig struct {
		InboundAuthenticationRequestConfigs []struct {
			InboundAuthType              string `yaml:"inboundAuthType" json:"inboundAuthType"`
			InboundAuthKey               string `yaml:"inboundAuthKey" json:"inboundAuthKey"`
			InboundConfigurationProtocol struct {
				OauthConsumerSecret string `yaml:"oauthConsumerSecret" json:"oauthConsumerSecret"`
			} `yaml:"inboundConfigurationProtocol" json:"inboundConfigurationProtocol"`
		} `yaml:"inboundAuthenticationRequestConfigs" json:"inboundAuthenticationRequestConfigs"`
	} `yaml:"inboundAuthenticationConfig" json:"inboundAuthenticationConfig"`
}

var deployedRoleNameMap map[string]string

func getDeployedAppNames(apps []Application) []string {

	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}
	return appNames
}

func getAppList() ([]Application, error) {

	data, err := utils.SendPaginatedGetListRequest(
		utils.APPLICATIONS,
		"totalResults",
		"count",
		"offset",
		"limit",
		"applications",
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving application list: %w", err)
	}
	var apps []Application
	if err := json.Unmarshal(data, &apps); err != nil {
		return nil, fmt.Errorf("error when unmarshalling application list: %w", err)
	}
	return apps, nil
}

func getAppKeywordMapping(appName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ApplicationConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(appName, utils.KEYWORD_CONFIGS.ApplicationConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getAppId(appName string, appList []Application) string {

	for _, app := range appList {
		if app.Name == appName {
			return app.Id
		}
	}
	return ""
}

func isOauthApp(fileData string, format utils.Format) (bool, error) {

	config, err := unmarshalAuthConfig([]byte(fileData), format)
	if err != nil {
		return false, err
	}

	for _, requestConfig := range config.InboundAuthenticationConfig.InboundAuthenticationRequestConfigs {
		if strings.ToLower(requestConfig.InboundAuthType) == utils.OAUTH2 {
			return true, nil
		}
	}
	return false, nil
}

func unmarshalAuthConfig(data []byte, format utils.Format) (AuthConfig, error) {

	var config AuthConfig
	if _, err := utils.Deserialize(data, format, utils.APPLICATIONS, &config); err != nil {
		return AuthConfig{}, fmt.Errorf("failed to unmarshal auth config: %s", err.Error())
	}
	return config, nil
}

func maskOAuthConsumerSecret(fileContent []byte) []byte {

	// Find and replace the value of oauthConsumerSecret with a mask.
	yamlPattern := regexp.MustCompile(`(?m)(^\s*oauthConsumerSecret:\s*)null\s*$`)
	maskedContent := yamlPattern.ReplaceAllString(string(fileContent), "${1}"+utils.SENSITIVE_FIELD_MASK)

	jsonPattern := regexp.MustCompile(`("oauthConsumerSecret":\s*)null`)
	maskedContent = jsonPattern.ReplaceAllString(maskedContent, `${1}"`+utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES+`"`)

	return []byte(maskedContent)
}

func removeAssociatedRoles(fileContent []byte) []byte {

	yamlPattern := regexp.MustCompile(`(?m)(^\s+roles:)[^\n]*\n(\s+-[^\n]*\n)*`)
	result := yamlPattern.ReplaceAllString(string(fileContent), "${1} []\n")

	jsonPattern := regexp.MustCompile(`("roles"\s*:\s*)\[(?:\s*\{[^}]*\}\s*,?\s*)*\]`)
	result = jsonPattern.ReplaceAllString(result, `${1}[]`)

	return []byte(result)
}

func injectDeployedOAuthCredentials(appId, fileData string, format utils.Format) (string, error) {

	oidcConfig, err := getDeployedInboundProtocolConfig(appId, "oidc")
	if err != nil {
		return fileData, err
	}
	if oidcConfig == nil {
		return fileData, nil
	}
	clientSecret, ok := oidcConfig["clientSecret"].(string)
	if !ok {
		return fileData, fmt.Errorf("clientSecret not found in deployed oidc config")
	}
	clientId, ok := oidcConfig["clientId"].(string)
	if !ok {
		return fileData, fmt.Errorf("clientId not found in deployed oidc config")
	}
	result := fileData

	switch format {
	case utils.FormatYAML:
		re := regexp.MustCompile(`(?m)(^\s*oauthConsumerSecret:\s*)[^\n]*$`)
		result = re.ReplaceAllString(result, "${1}"+clientSecret)
		re = regexp.MustCompile(`(?m)(^\s*oauthConsumerKey:\s*)[^\n]*$`)
		result = re.ReplaceAllString(result, "${1}"+clientId)
		re = regexp.MustCompile(`(?m)(^\s*inboundAuthKey:\s*)[^\n]*$`)
		result = re.ReplaceAllString(result, "${1}"+clientId)
	case utils.FormatJSON:
		re := regexp.MustCompile(`("oauthConsumerSecret":\s*)(?:null|"[^"]*")`)
		result = re.ReplaceAllString(result, `${1}"`+clientSecret+`"`)
		re = regexp.MustCompile(`("oauthConsumerKey":\s*)"[^"]*"`)
		result = re.ReplaceAllString(result, `${1}"`+clientId+`"`)
		re = regexp.MustCompile(`("inboundAuthKey":\s*)"[^"]*"`)
		result = re.ReplaceAllString(result, `${1}"`+clientId+`"`)
	}
	return result, nil
}

func isToolMgtApp(appId string) (bool, error) {

	oidcConfig, err := getDeployedInboundProtocolConfig(appId, "oidc")
	if err != nil {
		return false, err
	}
	if oidcConfig == nil {
		return false, nil
	}

	clientId, ok := oidcConfig["clientId"].(string)
	if !ok {
		return false, fmt.Errorf("clientId not found in deployed oidc config")
	}
	return clientId == utils.SERVER_CONFIGS.ClientId, nil
}

func isOauthSecretGiven(modifiedFileData string, format utils.Format) (bool, error) {

	config, err := unmarshalAuthConfig([]byte(modifiedFileData), format)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	for _, requestConfig := range config.InboundAuthenticationConfig.InboundAuthenticationRequestConfigs {
		if strings.ToLower(requestConfig.InboundAuthType) == utils.OAUTH2 {
			if requestConfig.InboundConfigurationProtocol.OauthConsumerSecret != "" {
				return true, nil
			}
		}
	}
	return false, nil
}

func processInboundProtocolConfigs(appId string, inboundProtocols []inboundProtocolRef, appMap map[string]interface{}, excludeSecrets bool) error {

	result := make(map[string]interface{})
	var customProtocols []interface{}
	var skippedProtocols []string

	for _, protocol := range inboundProtocols {
		self := protocol.Self
		protocolPath := path.Base(self)

		if _, skip := unsupportedInboundProtocols[protocolPath]; skip {
			skippedProtocols = append(skippedProtocols, protocolPath)
			continue
		}

		body, err := utils.SendGetRequest(utils.APPLICATIONS, appId+"/inbound-protocols/"+protocolPath)
		if err != nil {
			return fmt.Errorf("error retrieving inbound protocol %s: %w", protocolPath, err)
		}
		var protocolConfig map[string]interface{}
		if err := json.Unmarshal(body, &protocolConfig); err != nil {
			return fmt.Errorf("error unmarshalling inbound protocol %s: %w", protocolPath, err)
		}

		if protocolPath == "oidc" && excludeSecrets {
			maskOIDCClientSecret(protocolConfig)
		}
		if protocolPath == "saml" {
			protocolConfig = map[string]interface{}{
				"manualConfiguration": protocolConfig,
			}
		}

		protocolConfig["_type"] = protocolPath
		if key, known := protocolPathToKey[protocolPath]; known {
			result[key] = protocolConfig
		} else {
			customProtocols = append(customProtocols, protocolConfig)
		}
	}

	if len(customProtocols) > 0 {
		result["custom"] = customProtocols
	}
	if len(skippedProtocols) > 0 {
		utils.PrintLog(utils.LogLevelWarn, utils.APPLICATIONS, "", fmt.Sprintf("Skipped unsupported inbound protocols: %v", skippedProtocols))
	}

	delete(appMap, "inboundProtocols")
	appMap["inboundProtocolConfiguration"] = result
	return nil
}

func maskOIDCClientSecret(oidcConfig map[string]interface{}) {

	oidcConfig["clientSecret"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
}

func processInboundProtocolsForPost(appMap map[string]interface{}) (newSecretCreated bool, err error) {

	inboundConfig, ok := appMap["inboundProtocolConfiguration"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected format for inboundProtocolConfiguration")
	}

	for key, config := range inboundConfig {
		if key == "custom" {
			customArr, ok := config.([]interface{})
			if !ok {
				return false, fmt.Errorf("unexpected format for custom inbound protocols")
			}
			for _, item := range customArr {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					return false, fmt.Errorf("unexpected format for custom inbound protocol")
				}
				delete(itemMap, "_type")
			}
			continue
		}

		configMap, ok := config.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("unexpected format for inbound protocol: %s", key)
		}
		delete(configMap, "_type")

		if key == "oidc" {
			secret, _ := configMap["clientSecret"].(string)
			if secret == "" {
				newSecretCreated = true
			}
		}
	}
	return newSecretCreated, nil
}

func getDeployedInboundProtocols(appId string) ([]inboundProtocolRef, error) {

	body, err := utils.SendGetRequest(utils.APPLICATIONS, appId+"/inbound-protocols")
	if err != nil {
		return nil, err
	}
	var deployedRefs []inboundProtocolRef
	if err := json.Unmarshal(body, &deployedRefs); err != nil {
		return nil, fmt.Errorf("error unmarshalling deployed inbound protocols: %w", err)
	}
	return deployedRefs, nil
}

func flattenInboundProtocols(localProtocolConfig map[string]interface{}) ([]map[string]interface{}, error) {

	var result []map[string]interface{}
	for key, config := range localProtocolConfig {
		if key == "custom" {
			customArr, ok := config.([]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected format for custom inbound protocols")
			}
			for _, item := range customArr {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("unexpected format for custom inbound protocol")
				}
				result = append(result, itemMap)
			}
		} else {
			configMap, ok := config.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected format for inbound protocol: %s", key)
			}
			result = append(result, configMap)
		}
	}
	return result, nil
}

func processInboundProtocolForUpdate(appId string, protocolMap map[string]interface{}) (protocolPath string, err error) {

	protocolPath, ok := protocolMap["_type"].(string)
	if !ok {
		return "", fmt.Errorf("_type not found in inbound protocol")
	}
	delete(protocolMap, "_type")

	if protocolPath == "oidc" || protocolPath == "passive-sts" {
		if err := injectDeployedReadOnlyFields(appId, protocolPath, protocolMap); err != nil {
			return "", err
		}
	}
	return protocolPath, nil
}

func getDeployedInboundProtocolConfig(appId, protocolPath string) (map[string]interface{}, error) {

	body, err := utils.SendGetRequest(utils.APPLICATIONS, appId+"/inbound-protocols/"+protocolPath)
	if err != nil {
		if strings.Contains(err.Error(), "Resource not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("error retrieving deployed %s config: %w", protocolPath, err)
	}
	var deployedConfig map[string]interface{}
	if err := json.Unmarshal(body, &deployedConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling deployed %s config: %w", protocolPath, err)
	}
	return deployedConfig, nil
}

func injectDeployedReadOnlyFields(appId, protocolPath string, localConfig map[string]interface{}) error {

	deployedConfig, err := getDeployedInboundProtocolConfig(appId, protocolPath)
	if err != nil {
		return err
	}
	if deployedConfig == nil {
		return nil
	}

	switch protocolPath {
	case "oidc":
		clientId, ok := deployedConfig["clientId"].(string)
		if !ok {
			return fmt.Errorf("clientId not found in deployed oidc config")
		}
		localConfig["clientId"] = clientId

		deployedSecret, ok := deployedConfig["clientSecret"].(string)
		if !ok {
			return fmt.Errorf("clientSecret not found in deployed oidc config")
		}
		localConfig["clientSecret"] = deployedSecret

	case "passive-sts":
		realm, ok := deployedConfig["realm"].(string)
		if !ok {
			return fmt.Errorf("realm not found in deployed passive-sts config")
		}
		localConfig["realm"] = realm
	}
	return nil
}

func removeRoleClaimUri(appMap map[string]interface{}) error {

	if !claims.RoleClaimUnsupported() {
		return nil
	}
	claimConfMap, ok := appMap["claimConfiguration"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for claimConfiguration")
	}
	role, ok := claimConfMap["role"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for role in claimConfiguration")
	}
	claim, ok := role["claim"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for claim in role")
	}
	uri, ok := claim["uri"].(string)
	if !ok {
		return fmt.Errorf("unexpected format for uri in role claim")
	}
	if uri == "http://wso2.org/claims/role" {
		claim["uri"] = ""
	}
	return nil
}

func filterAssociatedApplicationRoles(appMap map[string]interface{}) error {

	assocRolesRaw, exists := appMap["associatedRoles"]
	if !exists {
		return nil
	}
	assocRoles, ok := assocRolesRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for associatedRoles")
	}
	rolesRaw, exists := assocRoles["roles"]
	if !exists {
		return nil
	}
	rolesList, ok := rolesRaw.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for roles in associatedRoles")
	}

	filtered := make([]interface{}, 0, len(rolesList))
	for _, item := range rolesList {
		roleMap, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for role in associatedRoles")
		}
		name, ok := roleMap["name"].(string)
		if !ok {
			return fmt.Errorf("unexpected format for role name in associatedRoles")
		}
		deployedId, exists := deployedRoleNameMap[name]
		if !exists {
			continue
		}
		roleMap["id"] = deployedId
		filtered = append(filtered, roleMap)
	}
	assocRoles["roles"] = filtered
	return nil
}

func removeAssociatedApplicationRoles(appMap map[string]interface{}) error {

	assocRolesRaw, exists := appMap["associatedRoles"]
	if !exists {
		return nil
	}
	assocRoles, ok := assocRolesRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for associatedRoles")
	}
	delete(assocRoles, "roles")
	return nil
}

func removeAdditionalSpProperties(appMap map[string]interface{}) error {

	advConf, ok := appMap["advancedConfigurations"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for advancedConfigurations")
	}
	delete(advConf, "additionalSpProperties")
	return nil
}

func InitDeployedRoleIds() error {

	roleList, err := roles.GetRoleList()
	if err != nil {
		return fmt.Errorf("error retrieving role list: %w", err)
	}
	deployedRoleNameMap = make(map[string]string, len(roleList))
	for _, r := range roleList {
		deployedRoleNameMap[r.DisplayName] = r.Id
	}
	return nil
}
