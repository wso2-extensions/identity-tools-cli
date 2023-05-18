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
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"
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

func SendExportRequest(resourceId, fileType, resourceType string, excludeSecrets bool) (resp *http.Response, err error) {

	var reqUrl string
	if resourceType == APPLICATIONS {
		reqUrl = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/" + getResourcePath(resourceType) + "/" + resourceId + "/exportFile"
	} else {
		reqUrl = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/" + getResourcePath(resourceType) + "/" + resourceId + "/export"
	}
	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		return resp, fmt.Errorf("error while creating the export request: %s", err)
	}
	req.Header.Set("Content-Type", MEDIA_TYPE_FORM)
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)

	query := req.URL.Query()
	if resourceType == APPLICATIONS {
		query.Add("exportSecrets", strconv.FormatBool(!excludeSecrets))
		req.URL.RawQuery = query.Encode()
	} else if resourceType == IDENTITY_PROVIDERS {
		query.Add("excludeSecrets", strconv.FormatBool(excludeSecrets))
		req.URL.RawQuery = query.Encode()
	}

	defer req.Body.Close()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err = httpClient.Do(req)
	if err != nil {
		return resp, fmt.Errorf("error while exporting resource: %s", err)
	}

	statusCode := resp.StatusCode
	if statusCode == 200 {
		return resp, nil
	} else if error, ok := ErrorCodes[statusCode]; ok {
		return resp, fmt.Errorf("error while exporting resource: %s", error)
	}
	return resp, fmt.Errorf("unexpected error while exporting the resource with status code: %s", strconv.FormatInt(int64(statusCode), 10))
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

	var reqUrl string
	if resourceType == APPLICATIONS {
		reqUrl = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/" + getResourcePath(resourceType) + "/import"
	} else {
		reqUrl = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/" + getResourcePath(resourceType) + "/" + resourceId + "/import"
	}

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

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func SendDeleteRequest(resourceId string, resourceType string) error {

	reqUrl := SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/" + resourceType + "/" + resourceId

	request, err := http.NewRequest("DELETE", reqUrl, bytes.NewBuffer(nil))
	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		return fmt.Errorf("error when creating the delete request: %s", err)
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
		return fmt.Errorf("error when sending the delete request: %s", err)
	}

	statusCode := resp.StatusCode
	if statusCode == 204 {
		log.Println("Resource deleted successfully.")
		return nil
	} else if error, ok := ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the delete request: %s", error)
	}
	return fmt.Errorf("unexpected error when deleting resource: %s", resp.Status)
}
