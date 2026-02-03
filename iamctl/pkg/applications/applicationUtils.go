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
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

type Application struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AppList struct {
	AppCount     int           `json:"totalResults"`
	Applications []Application `json:"applications"`
}

type AppConfig struct {
	ApplicationName string `yaml:"applicationName"`
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
	resp, err := utils.SendGetListRequest(configs.APPLICATIONS, totalAppCount)
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
	resp, err := utils.SendGetListRequest(configs.APPLICATIONS, -1)
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve available app list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return -1, fmt.Errorf("error when reading the retrived app list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return -1, fmt.Errorf("error when unmarshalling the retrived app list. %w", err)
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

func isToolMgtApp(file os.FileInfo, importFilePath string) (bool, error) {

	appFilePath := filepath.Join(importFilePath, file.Name())
	fileData, err := ioutil.ReadFile(appFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %s", err.Error())
	}

	config, err := unmarshalAuthConfig(fileData)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	for _, requestConfig := range config.InboundAuthenticationConfig.InboundAuthenticationRequestConfigs {
		if requestConfig.InboundAuthKey == utils.SERVER_CONFIGS.ClientId {
			appName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			log.Printf("Info: Tool Management App: %s is excluded from deletion.\n", appName)
			return true, nil
		}
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
