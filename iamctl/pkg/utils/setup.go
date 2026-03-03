/**
* Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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

package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type oAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	Expires      int    `json:"expires_in"`
}

type ServerConfigs struct {
	ServerUrl    string `json:"SERVER_URL"`
	ClientId     string `json:"CLIENT_ID"`
	ClientSecret string `json:"CLIENT_SECRET"`
	TenantDomain string `json:"TENANT_DOMAIN"`
	Organization string `json:"ORGANIZATION"`
	Token        string `json:"TOKEN"`
}

type ToolConfigs struct {
	AllowDelete        bool                   `json:"ALLOW_DELETE"`
	Exclude            []string               `json:"EXCLUDE"`
	IncludeOnly        []string               `json:"INCLUDE_ONLY"`
	ExcludeSecrets     bool                   `json:"EXCLUDE_SECRETS"`
	ApplicationConfigs map[string]interface{} `json:"APPLICATIONS"`
	IdpConfigs         map[string]interface{} `json:"IDENTITY_PROVIDERS"`
	ClaimConfigs       map[string]interface{} `json:"CLAIMS"`
	UserStoreConfigs   map[string]interface{} `json:"USERSTORES"`
	OidcScopeConfigs   map[string]interface{} `json:"OIDC_SCOPES"`
}

type KeywordConfigs struct {
	KeywordMappings    map[string]interface{} `json:"KEYWORD_MAPPINGS"`
	ApplicationConfigs map[string]interface{} `json:"APPLICATIONS"`
	IdpConfigs         map[string]interface{} `json:"IDENTITY_PROVIDERS"`
	ClaimConfigs       map[string]interface{} `json:"CLAIMS"`
	UserStoreConfigs   map[string]interface{} `json:"USERSTORES"`
	OidcScopeConfigs   map[string]interface{} `json:"OIDC_SCOPES"`
}

var SERVER_CONFIGS ServerConfigs
var TOOL_CONFIGS ToolConfigs
var KEYWORD_CONFIGS KeywordConfigs

func LoadConfigs(envConfigPath string) (baseDir string) {

	baseDir, toolConfigFile, keywordConfigPath := loadServerConfigs(envConfigPath)
	TOOL_CONFIGS = loadToolConfigsFromFile(toolConfigFile)
	KEYWORD_CONFIGS = loadKeywordConfigsFromFile(keywordConfigPath)
	return baseDir
}

func loadServerConfigs(envConfigPath string) (baseDir string, toolConfigPath string, keywordConfigPath string) {

	if envConfigPath == "" {
		log.Println("Loading configs from environment variables.")
		toolConfigPath, keywordConfigPath = loadConfigsFromEnvVar()
		baseDir = filepath.Dir(filepath.Dir(filepath.Dir(toolConfigPath)))
	} else {
		log.Println("Loading configs from config files.")
		baseDir = filepath.Dir(filepath.Dir(envConfigPath))
		serverConfigFile := filepath.Join(envConfigPath, SERVER_CONFIG_FILE)
		toolConfigPath = filepath.Join(envConfigPath, TOOL_CONFIG_FILE)
		keywordConfigPath = filepath.Join(envConfigPath, KEYWORD_CONFIG_FILE)

		SERVER_CONFIGS = loadServerConfigsFromFile(serverConfigFile)
	}
	sanitizeServerConfigs()

	// Get access token.
	SERVER_CONFIGS.Token = getAccessToken(SERVER_CONFIGS)
	log.Println("Access Token recieved succesfully.")
	return baseDir, toolConfigPath, keywordConfigPath
}

func loadConfigsFromEnvVar() (toolConfigPath string, keywordConfigPath string) {

	// Load server configs from environment variables.
	SERVER_CONFIGS.ServerUrl = os.Getenv(SERVER_URL_CONFIG)
	SERVER_CONFIGS.ClientId = os.Getenv(CLIENT_ID_CONFIG)
	SERVER_CONFIGS.ClientSecret = os.Getenv(CLIENT_SECRET_CONFIG)
	SERVER_CONFIGS.TenantDomain = os.Getenv(TENANT_DOMAIN_CONFIG)
	SERVER_CONFIGS.Organization = os.Getenv(ORGANIZATION_CONFIG)

	// Load tool config file path from environment variables.
	toolConfigPath = os.Getenv(TOOL_CONFIG_PATH)
	keywordConfigPath = os.Getenv(KEYWORD_CONFIG_PATH)
	return toolConfigPath, keywordConfigPath
}

func loadServerConfigsFromFile(configFilePath string) (serverConfigs ServerConfigs) {

	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Replace placeholder keys with environment variable values
	configFile = ReplacePlaceholders(configFile)

	reader := bytes.NewReader(configFile)
	jsonParser := json.NewDecoder(reader)
	err = jsonParser.Decode(&serverConfigs)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Server configs loaded succesfully from the config file.")
	return serverConfigs
}

func loadToolConfigsFromFile(configFilePath string) (toolConfigs ToolConfigs) {

	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln("Error when reading the tool config file.", err.Error())
	}

	if len(configFile) == 0 {
		return toolConfigs
	}

	TOOL_CONFIGS.ExcludeSecrets = true
	err = json.Unmarshal(configFile, &toolConfigs)
	if err != nil {
		log.Fatalln("Tool configs are not in the correct format. Please check the config file.", err)
	}

	log.Println("Tool configs loaded successfully from the config file.")
	return toolConfigs
}

func loadKeywordConfigsFromFile(configFilePath string) (keywordConfigs KeywordConfigs) {

	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln("Error when reading the keyword config file.", err.Error())
	}

	if len(configFile) == 0 {
		return keywordConfigs
	}

	// Replace placeholder keys with environment variable values
	configFile = ReplacePlaceholders(configFile)

	err = json.Unmarshal(configFile, &keywordConfigs)
	if err != nil {
		log.Fatalln("Keyword configs are not in the correct format. Please check the config file.", err)
	}

	log.Println("Keyword configs loaded successfully from the config file.")
	return keywordConfigs
}

func getAccessToken(config ServerConfigs) string {

	var err error
	var response oAuthResponse

	if config.ServerUrl == "" {
		log.Fatalln("Server URL is not defined in the config file.")
	}
	authUrl := config.ServerUrl + "/t/" + config.TenantDomain + "/oauth2/token"

	body := url.Values{}
	body.Set("grant_type", "client_credentials")
	body.Set("scope", SCOPE)

	req, err := http.NewRequest("POST", authUrl, strings.NewReader(body.Encode()))
	if err != nil {
		log.Fatalln(err)
	}
	req.SetBasicAuth(config.ClientId, config.ClientSecret)
	req.Header.Set("Content-Type", MEDIA_TYPE_FORM)
	defer req.Body.Close()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalln("Error in getting access token, response: " + string(respBody))
	}

	err2 := json.Unmarshal(respBody, &response)
	if err2 != nil {
		log.Fatalln(err2)
	}
	if IsSubOrganization() {
		log.Println("Getting access token for Organization: " + SERVER_CONFIGS.Organization)
		return switchAccessToken(config, response.AccessToken)
	}
	return response.AccessToken
}

func switchAccessToken(config ServerConfigs, accessToken string) string {

	var err error
	var response oAuthResponse

	authUrl := config.ServerUrl + "/t/" + config.TenantDomain + "/oauth2/token"

	body := url.Values{}
	body.Set("grant_type", "organization_switch")
	body.Set("scope", SCOPE)
	body.Set("token", accessToken)
	body.Set("switching_organization", config.Organization)

	req, err := http.NewRequest("POST", authUrl, strings.NewReader(body.Encode()))
	if err != nil {
		log.Fatalln(err)
	}
	req.SetBasicAuth(config.ClientId, config.ClientSecret)
	req.Header.Set("Content-Type", MEDIA_TYPE_FORM)
	defer req.Body.Close()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalln("Error in switching access token, response: " + string(respBody))
	}

	err2 := json.Unmarshal(respBody, &response)
	if err2 != nil {
		log.Fatalln(err2)
	}
	return response.AccessToken
}

func sanitizeServerConfigs() {

	SERVER_CONFIGS.ServerUrl = strings.TrimSuffix(SERVER_CONFIGS.ServerUrl, "/")

	// Set tenant domain if not defined in the config file.
	if SERVER_CONFIGS.TenantDomain == "" {
		log.Println("Tenant domain not defined. Defaulting to: carbon.super")
		SERVER_CONFIGS.TenantDomain = DEFAULT_TENANT_DOMAIN
	}
}

func IsSubOrganization() bool {

	return SERVER_CONFIGS.Organization != ""
}
