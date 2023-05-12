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

package utils

import (
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
	Token        string `json:"TOKEN"`
}

type ToolConfigs struct {
	KeywordMappings    map[string]interface{} `json:"KEYWORD_MAPPINGS"`
	AllowDelete        bool                   `json:"ALLOW_DELETE"`
	ApplicationConfigs map[string]interface{} `json:"APPLICATIONS"`
	IdpConfigs         map[string]interface{} `json:"IDENTITY_PROVIDERS"`
	UserStoreConfigs   map[string]interface{} `json:"USERSTORES"`
}

type Application struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Self        string `json:"self"`
}

type AppList struct {
	TotalResults int           `json:"totalResults"`
	StartIndex   int           `json:"startIndex"`
	Count        int           `json:"count"`
	Applications []Application `json:"applications"`
	Links        []string      `json:"links"`
}

var SERVER_CONFIGS ServerConfigs
var TOOL_CONFIGS ToolConfigs

func LoadConfigs(envConfigPath string) (baseDir string) {

	baseDir, toolConfigFile := loadServerConfigs(envConfigPath)
	TOOL_CONFIGS = loadToolConfigsFromFile(toolConfigFile)
	return baseDir
}

func loadServerConfigs(envConfigPath string) (baseDir string, toolConfigPath string) {

	if envConfigPath == "" {
		log.Println("Loading configs from environment variables.")
		toolConfigPath = loadConfigsFromEnvVar()
		baseDir = filepath.Dir(filepath.Dir(filepath.Dir(toolConfigPath)))
	} else {
		log.Println("Loading configs from config files.")
		baseDir = filepath.Dir(filepath.Dir(envConfigPath))
		serverConfigFile := filepath.Join(envConfigPath, SERVER_CONFIG_FILE)
		toolConfigPath = filepath.Join(envConfigPath, TOOL_CONFIG_FILE)

		SERVER_CONFIGS = loadServerConfigsFromFile(serverConfigFile)
	}
	sanitizeServerConfigs()

	// Get access token.
	SERVER_CONFIGS.Token = getAccessToken(SERVER_CONFIGS)
	log.Println("Access Token recieved succesfully.")
	return baseDir, toolConfigPath
}

func loadConfigsFromEnvVar() string {

	// Load server configs from environment variables.
	SERVER_CONFIGS.ServerUrl = os.Getenv(SERVER_URL_CONFIG)
	SERVER_CONFIGS.ClientId = os.Getenv(CLIENT_ID_CONFIG)
	SERVER_CONFIGS.ClientSecret = os.Getenv(CLIENT_SECRET_CONFIG)
	SERVER_CONFIGS.TenantDomain = os.Getenv(TENANT_DOMAIN_CONFIG)

	// Load tool config file path from environment variables.
	toolConfigPath := os.Getenv("TOOL_CONFIG_PATH")
	return toolConfigPath
}
func loadServerConfigsFromFile(configFilePath string) (serverConfigs ServerConfigs) {

	configFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
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

	err = json.Unmarshal(configFile, &toolConfigs)
	if err != nil {
		log.Fatalln("Tool configs are not in the correct format. Please check the config file.", err)
	}

	log.Println("Tool configs loaded successfully from the config file.")
	return toolConfigs
}

func getAccessToken(config ServerConfigs) string {

	var err error
	var response oAuthResponse

	if config.ServerUrl == "" {
		log.Fatalln("Server URL is not defined in the config file.")
	}
	authUrl := config.ServerUrl + "/oauth2/token"

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
