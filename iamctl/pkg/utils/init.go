/**
* Copyright (c) 2022, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type oAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	Expires      int    `json:"expires_in"`
}

type EnvConfigs struct {
	ServerUrl          string                 `json:"SERVER_URL"`
	ClientId           string                 `json:"CLIENT_ID"`
	ClientSecret       string                 `json:"CLIENT_SECRET"`
	TenantDomain       string                 `json:"TENANT_DOMAIN"`
	Username           string                 `json:"USERNAME"`
	Password           string                 `json:"PASSWORD"`
	Token              string                 `json:"TOKEN"`
	KeywordMappings    map[string]interface{} `json:"KEYWORD_MAPPINGS"`
	ApplicationConfigs map[string]interface{} `json:"APPLICATIONS"`
}

type Application struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Self        string `json:"self"`
}

type List struct {
	TotalResults int           `json:"totalResults"`
	StartIndex   int           `json:"startIndex"`
	Count        int           `json:"count"`
	Applications []Application `json:"applications"`
	Links        []string      `json:"links"`
}

var SERVER_CONFIGS EnvConfigs

func LoadServerConfigsFromFile(configFilePath string) (config EnvConfigs) {

	var rootDir, _ = os.Getwd()
	var configPath = rootDir + "/config.json"
	if configFilePath != "" {
		configPath = configFilePath
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer configFile.Close()

	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	fmt.Println("Server configs loaded succesfully from the config file.")

	config.Token = getAccessToken(config)
	fmt.Println("Access Token recieved succesfully.")

	return config
}

func getAccessToken(config EnvConfigs) string {

	var err error
	var response oAuthResponse

	authUrl := config.ServerUrl + "/oauth2/token"

	// Build response body to POST :=
	body := url.Values{}
	body.Set("grant_type", "password")
	body.Set("username", config.Username)
	body.Set("password", config.Password)
	body.Set("scope", SCOPE)

	req, err := http.NewRequest("POST", authUrl, strings.NewReader(body.Encode()))
	if err != nil {
		log.Fatalln(err)
	}
	req.SetBasicAuth(config.ClientId, config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

	body1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode == 401 {
		type clientError struct {
			Description string `json:"error_description"`
			Error       string `json:"error"`
		}
		var err = new(clientError)

		err2 := json.Unmarshal(body1, &err)
		if err2 != nil {
			log.Fatalln(err2)
		}
		fmt.Println(err.Error + "\n" + err.Description)
		return ""
	}

	err2 := json.Unmarshal(body1, &response)
	if err2 != nil {
		log.Fatalln(err2)
	}

	return response.AccessToken
}
