/*
 * Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cmd

import (
	"bytes"
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

type ServerInfo struct {
	Server      string `json:"server"`
	PackageName string `json:"packageName"`
	Application string `json:"application"`
}

type ServerConfigs struct {
	ServerUrl    string `json:"SERVER_URL"`
	ClientId     string `json:"CLIENT_ID"`
	ClientSecret string `json:"CLIENT_SECRET"`
	TenantDomain string `json:"TENANT_DOMAIN"`
	Username     string `json:"USERNAME"`
	Password     string `json:"PASSWORD"`
	Token        string `json:"TOKEN"`
}

func createFileIfNotExist(filepath string) {

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// File path contains the file name also.
		dirPathArray := strings.Split(filepath, "/")
		dirPath := ""
		for i := 0; i < len(dirPathArray)-1; i++ {
			dirPath += dirPathArray[i] + "/"
		}
		os.MkdirAll(dirPath, 0700)
		// create file if not exists
		var file, err = os.Create(filepath)
		if isError(err) {
			fmt.Println("Error while creating file :", filepath)
			return
		}
		defer file.Close()
	}
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}
	return (err != nil)
}

func writeFile(filepath string, writestring string) {

	var file, err = os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error while opening the file :", filepath)
	}
	if file != nil {
		defer file.Close()
		_, err = file.WriteString(writestring)
		if err != nil {
			fmt.Println("Error while writing the file :", filepath)
		}
	}
}

func readServerDetails() ServerDetails {

	var serverConfig ServerDetails
	var data myJSON
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error while reading the server configurations")
	}
	err = json.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("Error while reading the server configurations")
	}
	for i := 0; i < len(data.Array); i++ {
		serverConfig = data.Array[i]
	}
	return serverConfig
}

func callArtifactServiceApi(technology string, application string) *http.Response {

	serverDetails := readServerDetails()
	token := serverDetails.AccessToken
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	artifactServiceUrl := serverDetails.Server + "/artifact-service/service/artifact/" + technology

	toJson := ServerInfo{
		Server:      serverDetails.Server,
		PackageName: packageName,
		Application: application,
	}
	jsonData, err := json.Marshal(toJson)
	if err != nil {
		fmt.Println("Error while creating request body to artifact-service")
		return nil
	}

	var req, errReq = http.NewRequest(http.MethodPost, artifactServiceUrl, bytes.NewBuffer(jsonData))
	if errReq != nil {
		fmt.Println("Error while creating request to artifact-service")
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error while getting response from artifact-service")
		return nil
	}
	return resp
}

func writeToken(config ServerConfigs) bool {

	_, err2 := url.ParseRequestURI(config.ServerUrl)
	if err2 != nil {
		log.Fatalln(err2)
		return false
	}
	ur, err2 := url.Parse(config.ServerUrl)
	if err2 != nil {
		log.Fatalln(err2)
		return false
	} else {
		IAMURL = ur.Scheme + "://" + ur.Host
	}

	AUTHURL = IAMURL + "/oauth2/token"

	accessToken, refreshToken = sendTokenRequest(config)
	if accessToken != "" {
		writeFiles(IAMURL, accessToken, refreshToken)
	}

	return true
}

func sendTokenRequest(config ServerConfigs) (string, string) {

	var err error
	var accessToken string
	var refreshToken string
	var list oAuthResponse

	// Build response body to POST :=
	body := url.Values{}
	body.Set("grant_type", "password")
	body.Set("username", config.Username)
	body.Set("password", config.Password)
	body.Set("scope", SCOPE)

	req, err := http.NewRequest("POST", AUTHURL, strings.NewReader(body.Encode()))
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
		setSampleSP()
		return accessToken, refreshToken
	}

	err2 := json.Unmarshal(body1, &list)
	if err2 != nil {
		log.Fatalln(err2)
	}

	accessToken = list.AccessToken
	refreshToken = list.RefreshToken

	return accessToken, refreshToken
}

func loadServerConfigsFromFile(configFilePath string) (config ServerConfigs) {

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

	writeToken(config)
	config.Token = readFile()

	return config
}
