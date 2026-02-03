/**
* Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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
	"net/url"
	"strconv"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
)

const EXPORT = "export"
const IMPORT = "import"
const UPDATE = "update"
const DELETE = "delete"
const LIST = "list"

func SendExportRequest(resourceId, fileType, resourceType string, excludeSecrets bool) (resp *http.Response, err error) {

	reqUrl := buildRequestUrl(EXPORT, resourceType, resourceId)
	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		return resp, fmt.Errorf("error while creating the export request: %s", err)
	}
	req.Header.Set("Content-Type", MEDIA_TYPE_FORM)
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)

	query := req.URL.Query()
	if resourceType == configs.APPLICATIONS {
		query.Add("exportSecrets", strconv.FormatBool(!excludeSecrets))
		req.URL.RawQuery = query.Encode()
	} else if resourceType == configs.IDENTITY_PROVIDERS {
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

	reqUrl := buildRequestUrl(IMPORT, resourceType, "")

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

	reqUrl := buildRequestUrl(UPDATE, resourceType, resourceId)
	formattedReqUrl := addQueryParams(reqUrl, resourceType)

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

	request, err := http.NewRequest("PUT", formattedReqUrl, body)
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
	} else if statusCode == 400 && resourceType == configs.CLAIMS {
		return handleClaimImportErrorResponse(resp)
	} else if error, ok := ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing resource: %s", resp.Status)
}

func SendDeleteRequest(resourceId string, resourceType string) error {

	reqUrl := buildRequestUrl(DELETE, resourceType, resourceId)
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

func SendGetListRequest(resourceType string, resourceLimit int) (*http.Response, error) {

	var reqUrl = buildRequestUrl(LIST, resourceType, "")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, _ := http.NewRequest("GET", reqUrl, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	req.Header.Set("accept", "*/*")

	if resourceLimit != -1 {
		query := req.URL.Query()
		query.Add("limit", strconv.Itoa(resourceLimit))
		req.URL.RawQuery = query.Encode()
	}
	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve available userstore list. %w", err)
	}
	return resp, nil
}

func getResourcePath(resourceType string) string {

	switch resourceType {
	case configs.APPLICATIONS:
		return "applications"
	case configs.IDENTITY_PROVIDERS:
		return "identity-providers"
	case configs.USERSTORES:
		return "userstores"
	case configs.CLAIMS:
		return "claim-dialects"
	}
	return ""
}

func getResourceBaseUrl(resourceType string) string {

	basePath := "/t/" + SERVER_CONFIGS.TenantDomain
	if IsSubOrganization() {
		basePath += "/o"
	}
	basePath += "/api/server/v1/" + getResourcePath(resourceType) + "/"
	return SERVER_CONFIGS.ServerUrl + basePath
}

func buildRequestUrl(requestType, resourceType, resourceId string) (reqUrl string) {

	switch requestType {
	case EXPORT:
		if resourceType == configs.APPLICATIONS {
			reqUrl = getResourceBaseUrl(resourceType) + resourceId + "/exportFile"
		} else {
			reqUrl = getResourceBaseUrl(resourceType) + resourceId + "/" + EXPORT
		}
	case IMPORT:
		reqUrl = getResourceBaseUrl(resourceType) + IMPORT
	case UPDATE:
		if resourceType == configs.APPLICATIONS || resourceType == configs.CLAIMS {
			reqUrl = getResourceBaseUrl(resourceType) + IMPORT
		} else {
			reqUrl = getResourceBaseUrl(resourceType) + resourceId + "/" + IMPORT
		}
	case LIST:
		reqUrl = getResourceBaseUrl(resourceType)
	case DELETE:
		reqUrl = getResourceBaseUrl(resourceType) + resourceId
	}
	return reqUrl
}

func addQueryParams(reqURL, resourceType string) string {

	url, err := url.Parse(reqURL)
	if err != nil {
		log.Printf("Failed to parse URL: %s. Unable to add query parameters.", err)
		return reqURL
	}

	queryParams := url.Query()

	switch resourceType {
	case configs.CLAIMS:
		if resourceType == configs.CLAIMS && TOOL_CONFIGS.AllowDelete {
			queryParams.Set("preserveClaims", "true")
		}
	}

	url.RawQuery = queryParams.Encode()
	return url.String()
}
