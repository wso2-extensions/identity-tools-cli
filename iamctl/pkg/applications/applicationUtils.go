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
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v3"
)

type inboundProtocolRef struct {
	Self string `json:"self"`
}

type Application struct {
	Id               string               `json:"id"`
	Name             string               `json:"name"`
	InboundProtocols []inboundProtocolRef `json:"inboundProtocols"`
}

type AppList struct {
	AppCount     int           `json:"totalResults"`
	Applications []Application `json:"applications"`
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
			InboundAuthType              string `yaml:"inboundAuthType"`
			InboundAuthKey               string `yaml:"inboundAuthKey"`
			InboundConfigurationProtocol struct {
				OauthConsumerSecret string `yaml:"oauthConsumerSecret"`
			} `yaml:"inboundConfigurationProtocol"`
		} `yaml:"inboundAuthenticationRequestConfigs"`
	} `yaml:"inboundAuthenticationConfig"`
}

func getDeployedAppNames() []string {

	apps := getAppList()
	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}
	return appNames
}

func getAppList() (spIdList []Application) {

	totalAppCount, err := getTotalAppCount()
	if err != nil {
		log.Println("Error while retrieving application count. Retrieving only the default count.", err)
	}
	var list AppList
	resp, err := utils.SendGetListRequest(utils.APPLICATIONS, totalAppCount)
	if err != nil {
		log.Println("Error while retrieving application list", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		writer := new(tabwriter.Writer)
		writer.Init(os.Stdout, 8, 8, 0, '\t', 0)
		defer writer.Flush()

		err = json.Unmarshal(body, &list)
		if err != nil {
			log.Fatalln(err)
		}
		resp.Body.Close()

		spIdList = list.Applications
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		log.Println(error)
	} else {
		log.Println("Error while retrieving application list")
	}
	return spIdList
}

func getTotalAppCount() (count int, err error) {

	var list AppList
	resp, err := utils.SendGetListRequest(utils.APPLICATIONS, -1)
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve available app list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return -1, fmt.Errorf("error when reading the retrieved app list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return -1, fmt.Errorf("error when unmarshalling the retrieved app list. %w", err)
		}
		resp.Body.Close()

		return list.AppCount, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return -1, fmt.Errorf("error while retrieving app count. Status code: %d, Error: %s", statusCode, error)
	}
	return -1, fmt.Errorf("error while retrieving application count")
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

func isOauthApp(fileData string) (bool, error) {

	config, err := unmarshalAuthConfig([]byte(fileData))
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

func unmarshalAuthConfig(data []byte) (AuthConfig, error) {

	var config AuthConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return AuthConfig{}, fmt.Errorf("failed to unmarshal auth config: %s", err.Error())
	}
	return config, nil
}

func maskOAuthConsumerSecret(fileContent []byte) []byte {

	// Find and replace the value of oauthConsumerSecret with a mask.
	pattern := "(?m)(^\\s*oauthConsumerSecret:\\s*)null\\s*$"
	re := regexp.MustCompile(pattern)
	maskedContent := re.ReplaceAllString(string(fileContent), "${1}"+utils.SENSITIVE_FIELD_MASK)

	return []byte(maskedContent)
}

func isToolMgtApp(appId string) (bool, error) {

	body, err := utils.SendGetRequest(utils.APPLICATIONS, appId+"/inbound-protocols/oidc")
	if err != nil {
		// 404 means no OIDC config — not the tool management app
		if strings.Contains(err.Error(), "Resource not found") {
			return false, nil
		}
		return false, err
	}
	var oidcConfig map[string]interface{}
	if err := json.Unmarshal(body, &oidcConfig); err != nil {
		return false, fmt.Errorf("error unmarshalling OIDC config: %w", err)
	}

	clientId, _ := oidcConfig["clientId"].(string)
	if clientId == utils.SERVER_CONFIGS.ClientId {
		return true, nil
	}
	return false, nil
}

func isOauthSecretGiven(modifiedFileData string) (bool, error) {

	config, err := unmarshalAuthConfig([]byte(modifiedFileData))
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

		protocolConfig["self"] = self
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
		log.Printf("Warn: Skipped unsupported inbound protocols: %v", skippedProtocols)
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
				delete(itemMap, "self")
			}
			continue
		}

		configMap, ok := config.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("unexpected format for inbound protocol: %s", key)
		}
		delete(configMap, "self")

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

	self, ok := protocolMap["self"].(string)
	if !ok {
		return "", fmt.Errorf("self URL not found")
	}
	protocolPath = path.Base(self)
	delete(protocolMap, "self")

	if protocolPath == "oidc" || protocolPath == "passive-sts" {
		if err := injectDeployedReadOnlyFields(appId, protocolPath, protocolMap); err != nil {
			return "", err
		}
	}
	return protocolPath, nil
}

func injectDeployedReadOnlyFields(appId, protocolPath string, localConfig map[string]interface{}) error {

	body, err := utils.SendGetRequest(utils.APPLICATIONS, appId+"/inbound-protocols/"+protocolPath)
	if err != nil {
		return fmt.Errorf("error retrieving deployed %s config: %w", protocolPath, err)
	}
	var deployedConfig map[string]interface{}
	if err := json.Unmarshal(body, &deployedConfig); err != nil {
		return fmt.Errorf("error unmarshalling deployed %s config: %w", protocolPath, err)
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
