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
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "You can import a service provider",
	Long:  `You can import a service provider`,
	Run: func(cmd *cobra.Command, args []string) {
		importFilePath, errEXL := cmd.Flags().GetString("importFilePath")

		if errEXL != nil {
			log.Fatalln(errEXL)
		}
		importApplication(importFilePath)
	},
}

func init() {

	createSPCmd.AddCommand(importCmd)
	importCmd.Flags().StringP("importFilePath", "i", "", "set the export file name")
}

func importApplication(importFilePath string) bool {
	importedSp := false

	SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = utils.ReadSPConfig()

	start(SERVER, "admin", "admin")

	var ADDAPPURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications/import"
	var err error

	token := utils.ReadFile()

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	extraParams := map[string]string{
		"file": string(fileBytes),
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range extraParams {
		err := writer.WriteField(key, val)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer writer.Close()

	request, err := http.NewRequest("POST", ADDAPPURL, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
	defer request.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(bodyBytes))

		importedSp = true
	}

	return importedSp
}
