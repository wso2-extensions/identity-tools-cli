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
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

const EXPORT = "export"
const IMPORT = "import"
const UPDATE = "update"
const DELETE = "delete"
const LIST = "list"
const GET = "get"
const POST = "post"
const PUT = "put"
const PATCH = "patch"

type sendConfig struct {
	contentType string
	pathSuffix  string
}

type SendOption func(*sendConfig)

func PrepareJSONRequestBody(data []byte, format Format, resourceType ResourceType, excludeFields ...string) ([]byte, error) {

	dataMap, err := DeserializeToMap(data, format, resourceType, excludeFields...)
	if err != nil {
		return nil, err
	}

	jsonBody, err := Serialize(dataMap, FormatJSON, resourceType)
	if err != nil {
		return nil, fmt.Errorf("error serializing to JSON for request body: %w", err)
	}
	return jsonBody, nil
}

func PrepareMultipartFormBody(data []byte, format Format, resourceType ResourceType, excludeFields ...string) ([]byte, string, error) {

	dataMap, err := DeserializeToMap(data, format, resourceType, excludeFields...)
	if err != nil {
		return nil, "", err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for key, value := range dataMap {
		if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
			return nil, "", fmt.Errorf("error writing field %s: %w", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("error closing multipart writer: %w", err)
	}

	return body.Bytes(), writer.FormDataContentType(), nil
}

func RemoveResponseFields(response interface{}, fieldsToRemove ...string) (interface{}, error) {

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("response is not in expected format")
	}
	for _, field := range fieldsToRemove {
		delete(responseMap, field)
	}
	return responseMap, nil
}

func SendExportRequest(resourceId, fileType string, resourceType ResourceType, excludeSecrets bool) (resp *http.Response, err error) {

	reqUrl := buildRequestUrl(EXPORT, resourceType, resourceId)
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

func SendImportRequest(importFilePath, fileData string, resourceType ResourceType) error {

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

func SendUpdateRequest(resourceId, importFilePath, fileData string, resourceType ResourceType) error {

	reqUrl := buildRequestUrl(UPDATE, resourceType, resourceId)
	formattedReqUrl := addQueryParams(reqUrl, resourceType, UPDATE)

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
	} else if statusCode == 400 && resourceType == CLAIMS {
		return handleClaimImportErrorResponse(resp)
	} else if error, ok := ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing resource: %s", resp.Status)
}

func SendDeleteRequest(resourceId string, resourceType ResourceType) error {

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

func GetResourceData(resourceType ResourceType, resourceId string) (interface{}, error) {

	body, err := SendGetRequest(resourceType, resourceId)
	if err != nil {
		return nil, err
	}

	data, err := Deserialize(body, FormatJSON, resourceType)
	if err != nil {
		return nil, fmt.Errorf("error deserializing JSON response: %w", err)
	}
	return data, nil
}

func SendGetRequest(resourceType ResourceType, resourceId string) ([]byte, error) {

	reqUrl := buildRequestUrl(GET, resourceType, resourceId)
	formattedReqUrl := addQueryParams(reqUrl, resourceType, GET)
	request, err := http.NewRequest("GET", formattedReqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	request.Header.Set("Accept", MEDIA_TYPE_JSON)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if errMsg, ok := ErrorCodes[resp.StatusCode]; ok {
			return nil, fmt.Errorf("error response for the GET request: %s", errMsg)
		}
		return nil, fmt.Errorf("unexpected error when sending GET request: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func WithContentType(ct string) SendOption {
	return func(c *sendConfig) { c.contentType = ct }
}

func WithPathSuffix(suffix string) SendOption {
	return func(c *sendConfig) { c.pathSuffix = suffix }
}

func applySendOptions(opts []SendOption) *sendConfig {
	cfg := &sendConfig{contentType: MEDIA_TYPE_JSON}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func SendPostRequest(resourceType ResourceType, requestBody []byte, opts ...SendOption) (*http.Response, error) {

	cfg := applySendOptions(opts)
	reqUrl := buildRequestUrl(POST, resourceType, cfg.pathSuffix)

	request, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating POST request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	request.Header.Set("Content-Type", cfg.contentType)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		resp.Body.Close()
		if errMsg, ok := ErrorCodes[resp.StatusCode]; ok {
			return nil, fmt.Errorf("error response for the POST request: %s", errMsg)
		}
		return nil, fmt.Errorf("unexpected error when sending POST request: %s", resp.Status)
	}

	return resp, nil
}

func SendPutRequest(resourceType ResourceType, resourceId string, requestBody []byte, opts ...SendOption) (*http.Response, error) {

	cfg := applySendOptions(opts)
	reqUrl := buildRequestUrl(PUT, resourceType, resourceId)

	request, err := http.NewRequest("PUT", reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating PUT request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	request.Header.Set("Content-Type", cfg.contentType)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending PUT request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if errMsg, ok := ErrorCodes[resp.StatusCode]; ok {
			return nil, fmt.Errorf("error response for the PUT request: %s", errMsg)
		}
		return nil, fmt.Errorf("unexpected error when sending PUT request: %s", resp.Status)
	}

	return resp, nil
}

func SendPatchRequest(resourceType ResourceType, resourceId string, requestBody []byte) (*http.Response, error) {

	reqUrl := buildRequestUrl(PATCH, resourceType, resourceId)
	request, err := http.NewRequest("PATCH", reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating PATCH request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	request.Header.Set("Content-Type", MEDIA_TYPE_JSON)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending PATCH request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if errMsg, ok := ErrorCodes[resp.StatusCode]; ok {
			return nil, fmt.Errorf("error response for the PATCH request: %s", errMsg)
		}
		return nil, fmt.Errorf("unexpected error when sending PATCH request: %s", resp.Status)
	}

	return resp, nil
}

func SendGetListRequest(resourceType ResourceType, resourceLimit int) (*http.Response, error) {

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

func getResourcePath(resourceType ResourceType) string {

	switch resourceType {
	case APPLICATIONS:
		return "applications"
	case IDENTITY_PROVIDERS:
		return "identity-providers"
	case USERSTORES:
		return "userstores"
	case CLAIMS:
		return "claim-dialects"
	case OIDC_SCOPES:
		return "oidc/scopes"
	case CHALLENGE_QUESTIONS:
		return "challenges"
	case EMAIL_TEMPLATES:
		return "email/template-types"
	case SCRIPT_LIBRARIES:
		return "script-libraries"
	case GOVERNANCE_CONNECTORS:
		return "identity-governance"
	}
	return ""
}

func getResourceBaseUrl(resourceType ResourceType) string {

	basePath := "/t/" + SERVER_CONFIGS.TenantDomain
	if IsSubOrganization() {
		basePath += "/o"
	}
	if resourceType == ROLES {
		basePath += "/scim2/Roles/"
	} else {
		basePath += "/api/server/v1/" + getResourcePath(resourceType) + "/"
	}
	return SERVER_CONFIGS.ServerUrl + basePath
}

func buildRequestUrl(requestType string, resourceType ResourceType, resourceId string) (reqUrl string) {

	switch requestType {
	case EXPORT:
		if resourceType == APPLICATIONS {
			reqUrl = getResourceBaseUrl(resourceType) + resourceId + "/exportFile"
		} else {
			reqUrl = getResourceBaseUrl(resourceType) + resourceId + "/" + EXPORT
		}
	case IMPORT:
		reqUrl = getResourceBaseUrl(resourceType) + IMPORT
	case UPDATE:
		if resourceType == APPLICATIONS || resourceType == CLAIMS {
			reqUrl = getResourceBaseUrl(resourceType) + IMPORT
		} else {
			reqUrl = getResourceBaseUrl(resourceType) + resourceId + "/" + IMPORT
		}
	case LIST:
		reqUrl = getResourceBaseUrl(resourceType)
	case GET:
		reqUrl = getResourceBaseUrl(resourceType) + resourceId
	case POST:
		reqUrl = getResourceBaseUrl(resourceType) + resourceId
	case PUT:
		reqUrl = getResourceBaseUrl(resourceType) + resourceId
	case PATCH:
		reqUrl = getResourceBaseUrl(resourceType) + resourceId
	case DELETE:
		reqUrl = getResourceBaseUrl(resourceType) + resourceId
	}
	return reqUrl
}

func addQueryParams(reqURL string, resourceType ResourceType, operation string) string {

	url, err := url.Parse(reqURL)
	if err != nil {
		log.Printf("Failed to parse URL: %s. Unable to add query parameters.", err)
		return reqURL
	}

	queryParams := url.Query()

	switch resourceType {
	case CLAIMS:
		if operation == UPDATE && TOOL_CONFIGS.AllowDelete {
			queryParams.Set("preserveClaims", "true")
		}
	case ROLES:
		if operation == GET {
			queryParams.Set("excludedAttributes", "meta,users,groups")
		}
	}

	url.RawQuery = queryParams.Encode()
	return url.String()
}
