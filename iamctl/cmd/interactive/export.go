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

package interactive

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "You can export a service provider",
	Long:  `You can export a service provider`,
	Run: func(cmd *cobra.Command, args []string) {
		appId, _ := cmd.Flags().GetString("serviceProviderID")
		exportLocation, _ := cmd.Flags().GetString("exportlocation")
		fileType, _ := cmd.Flags().GetString("fileType")

		if appId == "" || exportLocation == "" {
			setExportInfo()
		} else {
			exportApplication(appId, exportLocation, fileType)
		}
	},
}

func init() {

	createSPCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("serviceProviderID", "s", "", "set the service provide ID")
	exportCmd.Flags().StringP("exportlocation", "p", "", "set the export location")
	exportCmd.Flags().StringP("fileType", "t", "application/yaml", "set the file type")
}

var exportQuestions = []*survey.Question{
	{
		Name:     "exportlocation",
		Prompt:   &survey.Input{Message: "Enter export location : "},
		Validate: survey.Required,
	},
	{
		Name:     "serviceProviderID",
		Prompt:   &survey.Input{Message: "Enter service provider id to be exported :"},
		Validate: survey.Required,
	},
	{
		Name:     "fileType",
		Prompt:   &survey.Input{Message: "Enter file type i.e application/json or application/yaml :"},
		Validate: survey.Required,
	},
}

func setExportInfo() {

	exportAnswers := struct {
		Exportlocation    string `survey:"exportlocation"`
		ServiceProviderID string `survey:"serviceProviderID"`
		FileType          string `survey:"fileType"`
	}{}

	err := survey.Ask(exportQuestions, &exportAnswers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	exportApplication(exportAnswers.ServiceProviderID, exportAnswers.Exportlocation, exportAnswers.FileType)
}

func exportApplication(serviceProviderID string, exportlocation string, fileType string) {

	SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = utils.ReadSPConfig()
	setServer()

	var ADDAPPURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications"
	var err error

	token := utils.ReadFile()

	var reqUrl = ADDAPPURL + "/" + serviceProviderID + "/exportFile"
	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+token)
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

	var attachmentDetail = resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(attachmentDetail)

	if err != nil {
		panic(err)
	}

	var fileName = params["filename"]

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var exportedFilePath = exportlocation + "/" + fileName
	err = ioutil.WriteFile(exportedFilePath, body, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Successfully created the export file : " + exportedFilePath)
}
