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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type AppConfig struct {
	ApplicationName string `yaml:"applicationName"`
}

func getAppFileInfo(filePath string) (string, string, string) {

	filename := filepath.Base(filePath)
	fileExtension := filepath.Ext(filename)
	appName := strings.TrimSuffix(filename, fileExtension)

	return appName, filename, fileExtension
}

func isAppExcluded(appName string) bool {

	// Include only the applications added to INCLUDE_ONLY config
	includeOnlyApps, ok := utils.TOOL_CONFIGS.ApplicationConfigs["INCLUDE_ONLY"].([]interface{})
	if ok {
		for _, app := range includeOnlyApps {
			if app.(string) == appName {
				return false
			}
		}
		log.Println("Application " + appName + " is excluded.")
		return true
	} else {
		// Exclude applications added to EXCLUDE_APPLICATIONS config
		appsToExclude, ok := utils.TOOL_CONFIGS.ApplicationConfigs["EXCLUDE"].([]interface{})
		if ok {
			for _, app := range appsToExclude {
				if app.(string) == appName {
					log.Println("Application " + appName + " is excluded.")
					return true
				}
			}
		}
		return false
	}
}

func getDeployedAppNames() []string {

	apps := getAppList()
	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}
	return appNames
}

func getAppList() (spIdList []utils.Application) {

	var APPURL = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications"
	var list utils.List

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, _ := http.NewRequest("GET", APPURL, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	req.Header.Set("accept", "*/*")
	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
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
