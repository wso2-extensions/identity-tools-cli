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
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all applications to the Applications folder.
	exportFilePath = exportFilePath + "/Applications/"
	os.MkdirAll(exportFilePath, 0700)

	apps := getAppList()
	for _, app := range apps {
		if !utils.IsResourceExcluded(app.Name, utils.TOOL_CONFIGS.ApplicationConfigs) {
			err := exportApp(app.Id, exportFilePath, format)
			if err != nil {
				log.Println("Error while exporting application: ", app.Name)
			} else {
				log.Println("Application exported successfully: ", app.Name)
			}
		}
	}
}

func exportApp(appId string, outputDirPath string, format string) error {

	var fileType = "application/yaml"
	if format == "json" {
		fileType = "application/json"
	} else if format == "xml" {
		fileType = "application/xml"
	}

	var APPURL = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications"
	var err error
	var reqUrl = APPURL + "/" + appId + "/exportFile"

	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		log.Println("Error: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
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
		log.Println("Error: ", err)
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		var attachmentDetail = resp.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(attachmentDetail)
		if err != nil {
			log.Println("Error while parsing the content disposition header", err)
			return err
		}

		var fileName = params["filename"]

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error: ", err)
			return err
		}
		exportedFile := outputDirPath + fileName

		// Add keywords to the exported file according to the keyword locations in the local file.
		appName, _, _ := getAppFileInfo(exportedFile)
		appKeywordMapping := getAppKeywordMapping(appName)
		modifiedFile := utils.AddKeywords(body, exportedFile, appKeywordMapping)

		err = ioutil.WriteFile(exportedFile, modifiedFile, 0644)
		if err != nil {
			log.Println("Error when writing the exported content to file: ", err)
			return err
		}
		return nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return errors.New(error)
	} else {
		return errors.New("Unexpected error while exporting the application")
	}
}
