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
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"

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
	setServer()

	var ADDAPPURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications/import"

	token := utils.ReadFile()

	file, err := os.Open(importFilePath)
	if err != nil {
		log.Fatal(err)
	}
	filename := filepath.Base(importFilePath)
	fileExtension := filepath.Ext(filename)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
    defer writer.Close()

	mime.AddExtensionType(".yml", "text/yaml")
	mime.AddExtensionType(".xml", "application/xml")

	mimeType := mime.TypeByExtension(fileExtension)

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal(err)
	}

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
	}

	switch resp.StatusCode {
	case 401:
		log.Println("Unauthorized access.\nPlease check your Username and password.")
	case 400:
		log.Println("Provided parameters are not in correct format.")
	case 403:
		log.Println("Forbidden request.")
	case 409:
		log.Println("An application with the same name already exists.")
	case 500:
		log.Println("Internal server error.")
	case 201:
		log.Println("Application imported successfully.")
	}

	return importedSp
}
