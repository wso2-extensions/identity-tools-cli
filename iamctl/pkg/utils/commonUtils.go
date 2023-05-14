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
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	ResourceName  string
	FileName      string
	FileExtension string
}

func GetFileInfo(filePath string) (fileInfo FileInfo) {

	fileInfo.FileName = filepath.Base(filePath)
	fileInfo.FileExtension = filepath.Ext(fileInfo.FileName)
	fileInfo.ResourceName = strings.TrimSuffix(fileInfo.FileName, fileInfo.FileExtension)

	return fileInfo
}

func getResourcePath(resourceType string) string {

	switch resourceType {
	case APPLICATIONS:
		return "applications"
	case IDENTITY_PROVIDERS:
		return "identity-providers"
	case USERSTORES:
		return "userstores"
	}
	return ""
}

func SendImportRequest(importFilePath, fileData, resourceType string) error {

	reqUrl := SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/" + getResourcePath(resourceType) + "/import"

	var buf bytes.Buffer
	var err error
	_, err = io.WriteString(&buf, fileData)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	mime.AddExtensionType(".yml", "application/yaml")
	mime.AddExtensionType(".xml", "application/xml")
	mime.AddExtensionType(".json", "application/json")

	fileInfo := GetFileInfo(importFilePath)
	mimeType := mime.TypeByExtension(fileInfo.FileExtension)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", fileInfo.FileName)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	_, err = io.Copy(part, &buf)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	request, err := http.NewRequest("POST", reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
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
		return fmt.Errorf("error when sending the import request: %s", err)
	}

	statusCode := resp.StatusCode
	if statusCode == 201 {
		return nil
	} else if error, ok := ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing resource: %s", resp.Status)
}

func SendUpdateRequest(resourceId, importFilePath, fileData, resourceType string) error {

	reqUrl := SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/" + getResourcePath(resourceType) + "/" + resourceId + "/import"

	var buf bytes.Buffer
	var err error
	_, err = io.WriteString(&buf, fileData)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	mime.AddExtensionType(".yml", "application/yaml")
	mime.AddExtensionType(".xml", "application/xml")
	mime.AddExtensionType(".json", "application/json")

	fileInfo := GetFileInfo(importFilePath)
	mimeType := mime.TypeByExtension(fileInfo.FileExtension)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", fileInfo.FileName)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	_, err = io.Copy(part, &buf)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	request, err := http.NewRequest("PUT", reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
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
		return fmt.Errorf("error when sending the import request: %s", err)
	}

	statusCode := resp.StatusCode

	if statusCode == 200 {
		return nil
	} else if error, ok := ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing resource: %s", resp.Status)
}
