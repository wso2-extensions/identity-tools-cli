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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	queryParams map[string]string
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

func ParseResponseBody(resp *http.Response, target ...interface{}) (interface{}, error) {

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if len(target) > 0 {
		if err := json.Unmarshal(body, target[0]); err != nil {
			return nil, fmt.Errorf("error unmarshalling response body: %w", err)
		}
		return target[0], nil
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}
	return result, nil
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
	}

	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	debugBody, _ := ioutil.ReadAll(resp.Body)
	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", statusCode, string(debugBody)))
	if error, ok := ErrorCodes[statusCode]; ok {
		return resp, fmt.Errorf("error while exporting resource: %s", error)
	}
	return resp, fmt.Errorf("unexpected error while exporting the resource with status code: %s", strconv.FormatInt(int64(statusCode), 10))
}

func SendImportRequest(importFilePath, fileData string, resourceType ResourceType) (*http.Response, error) {

	reqUrl := buildRequestUrl(IMPORT, resourceType, "")

	var buf bytes.Buffer
	var err error
	_, err = io.WriteString(&buf, fileData)
	if err != nil {
		return nil, fmt.Errorf("error when creating the import request: %s", err)
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
		return nil, fmt.Errorf("error when creating the import request: %s", err)
	}

	_, err = io.Copy(part, &buf)
	if err != nil {
		return nil, fmt.Errorf("error when creating the import request: %s", err)
	}

	var capturedImportBody bytes.Buffer
	var importBodyReader io.Reader = body
	if TOOL_CONFIGS.Logs.LogRequestPayloads {
		importBodyReader = io.TeeReader(body, &capturedImportBody)
	}

	request, err := http.NewRequest("POST", reqUrl, importBodyReader)
	if err != nil {
		return nil, fmt.Errorf("error when creating the import request: %s", err)
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	defer request.Body.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error when sending the import request: %s", err)
	}

	statusCode := resp.StatusCode
	if statusCode == 201 {
		return resp, nil
	}
	debugBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", request.Method, request.URL.String()))
	if TOOL_CONFIGS.Logs.LogRequestPayloads {
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Request body: %s", capturedImportBody.String()))
	}
	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", statusCode, string(debugBody)))
	if error, ok := ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error response for the import request: %s", error)
	}
	return nil, fmt.Errorf("unexpected error when importing resource: %s", resp.Status)
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

	var capturedUpdateBody bytes.Buffer
	var updateBodyReader io.Reader = body
	if TOOL_CONFIGS.Logs.LogRequestPayloads {
		updateBodyReader = io.TeeReader(body, &capturedUpdateBody)
	}

	request, err := http.NewRequest("PUT", formattedReqUrl, updateBodyReader)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	defer request.Body.Close()

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
	}
	debugBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", request.Method, request.URL.String()))
	if TOOL_CONFIGS.Logs.LogRequestPayloads {
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Request body: %s", capturedUpdateBody.String()))
	}
	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", statusCode, string(debugBody)))
	if error, ok := ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing resource: %s", resp.Status)
}

func SendDeleteRequest(resourceId string, resourceType ResourceType, opts ...SendOption) error {

	cfg := applySendOptions(opts)
	reqUrl := buildRequestUrl(DELETE, resourceType, resourceId)
	request, err := http.NewRequest("DELETE", reqUrl, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("error when creating the delete request: %s", err)
	}
	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)

	query := request.URL.Query()
	for k, v := range cfg.queryParams {
		query.Add(k, v)
	}
	request.URL.RawQuery = query.Encode()
	defer request.Body.Close()

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
		return nil
	}
	debugBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", request.Method, request.URL.String()))
	PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", statusCode, string(debugBody)))
	if error, ok := ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the delete request: %s", error)
	}
	return fmt.Errorf("unexpected error when deleting resource: %s", resp.Status)
}

func GetResourceData(resourceType ResourceType, resourceId string, opts ...SendOption) (interface{}, error) {

	body, err := SendGetRequest(resourceType, resourceId, opts...)
	if err != nil {
		return nil, err
	}

	data, err := Deserialize(body, FormatJSON, resourceType)
	if err != nil {
		return nil, fmt.Errorf("error deserializing JSON response: %w", err)
	}
	return data, nil
}

func WithContentType(ct string) SendOption {
	return func(c *sendConfig) { c.contentType = ct }
}

// WithPathSuffix appends a suffix to the request URL. Only used in POST requests
func WithPathSuffix(suffix string) SendOption {

	return func(c *sendConfig) { c.pathSuffix = suffix }
}

func WithQueryParams(params map[string]string) SendOption {

	return func(c *sendConfig) { c.queryParams = params }
}

func applySendOptions(opts []SendOption) *sendConfig {
	cfg := &sendConfig{contentType: MEDIA_TYPE_JSON}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func SendGetRequest(resourceType ResourceType, resourceId string, opts ...SendOption) ([]byte, error) {

	cfg := applySendOptions(opts)
	reqUrl := buildRequestUrl(GET, resourceType, resourceId)
	formattedReqUrl := addQueryParams(reqUrl, resourceType, GET)
	request, err := http.NewRequest("GET", formattedReqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	request.Header.Set("Accept", cfg.contentType)
	query := request.URL.Query()
	for k, v := range cfg.queryParams {
		query.Add(k, v)
	}
	request.URL.RawQuery = query.Encode()

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
		if !(request.Method == "GET" && resp.StatusCode == 404) {
			debugBody, _ := ioutil.ReadAll(resp.Body)
			PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", request.Method, request.URL.String()))
			PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", resp.StatusCode, string(debugBody)))
		}
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

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		debugBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", request.Method, request.URL.String()))
		if TOOL_CONFIGS.Logs.LogRequestPayloads {
			PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Request body: %s", string(requestBody)))
		}
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", resp.StatusCode, string(debugBody)))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		debugBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", request.Method, request.URL.String()))
		if TOOL_CONFIGS.Logs.LogRequestPayloads {
			PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Request body: %s", string(requestBody)))
		}
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", resp.StatusCode, string(debugBody)))
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
		debugBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", request.Method, request.URL.String()))
		if TOOL_CONFIGS.Logs.LogRequestPayloads {
			PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Request body: %s", string(requestBody)))
		}
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", resp.StatusCode, string(debugBody)))
		if errMsg, ok := ErrorCodes[resp.StatusCode]; ok {
			return nil, fmt.Errorf("error response for the PATCH request: %s", errMsg)
		}
		return nil, fmt.Errorf("unexpected error when sending PATCH request: %s", resp.Status)
	}

	return resp, nil
}

func SendGetListRequest(resourceType ResourceType, opts ...SendOption) ([]byte, error) {

	cfg := applySendOptions(opts)
	reqUrl := buildRequestUrl(LIST, resourceType, "")
	reqUrl = addQueryParams(reqUrl, resourceType, LIST)

	req, err := http.NewRequest("GET", reqUrl, bytes.NewBuffer(nil))
	if err != nil {
		return nil, fmt.Errorf("error creating GET list request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	req.Header.Set("Accept", MEDIA_TYPE_JSON)

	query := req.URL.Query()
	for k, v := range cfg.queryParams {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()
	defer req.Body.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending GET list request. %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		debugBody, _ := ioutil.ReadAll(resp.Body)
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", resp.StatusCode, string(debugBody)))
		if errMsg, ok := ErrorCodes[resp.StatusCode]; ok {
			return nil, fmt.Errorf("error response for the GET list request. Error: %s", errMsg)
		}
		return nil, fmt.Errorf("unexpected error when sending GET list request: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return body, nil
}

func SendPaginatedGetListRequest(resourceType ResourceType, totField, curCountField, offsetField, limitField, resultsField string, startIndex int, opts ...SendOption) ([]byte, error) {

	cfg := applySendOptions(opts)

	firstBody, err := SendGetListRequest(resourceType, WithQueryParams(cfg.queryParams))
	if err != nil {
		return nil, err
	}
	var firstResponse map[string]interface{}
	if err := json.Unmarshal(firstBody, &firstResponse); err != nil {
		return nil, fmt.Errorf("error parsing paginated list response: %w", err)
	}

	pageSize, err := extractIntField(firstResponse, curCountField)
	if err != nil {
		return nil, fmt.Errorf("error reading %s from response: %w", curCountField, err)
	}
	if pageSize == 0 {
		return json.Marshal([]interface{}{})
	}
	totalCount, err := extractIntField(firstResponse, totField)
	if err != nil {
		return nil, fmt.Errorf("error reading %s from response: %w", totField, err)
	}
	allResults, exists, err := extractResultsArray(firstResponse, resultsField)
	if err != nil {
		return nil, fmt.Errorf("error reading results array from response: %w", err)
	}
	if !exists && totalCount > 0 {
		return nil, fmt.Errorf("results field %q not found in response", resultsField)
	}

	if totalCount <= pageSize || pageSize == 0 {
		data, err := json.Marshal(allResults)
		if err != nil {
			return nil, fmt.Errorf("error marshalling initial results: %w", err)
		}
		return data, nil
	}

	for nextOffset := startIndex + pageSize; len(allResults) < totalCount; nextOffset += pageSize {
		pageParams := make(map[string]string)
		for k, v := range cfg.queryParams {
			pageParams[k] = v
		}
		pageParams[offsetField] = strconv.Itoa(nextOffset)
		pageParams[limitField] = strconv.Itoa(pageSize)

		pageBody, err := SendGetListRequest(resourceType, WithQueryParams(pageParams))
		if err != nil {
			return nil, err
		}
		var pageResponse map[string]interface{}
		if err := json.Unmarshal(pageBody, &pageResponse); err != nil {
			return nil, fmt.Errorf("error parsing paginated list response: %w", err)
		}

		pageResults, _, err := extractResultsArray(pageResponse, resultsField)
		if err != nil {
			return nil, fmt.Errorf("error reading results array from response: %w", err)
		}
		if len(pageResults) == 0 {
			break
		}
		allResults = append(allResults, pageResults...)
	}

	data, err := json.Marshal(allResults)
	if err != nil {
		return nil, fmt.Errorf("error marshalling combined results: %w", err)
	}
	return data, nil
}

func SendCustomRequest(method, reqURL string, body []byte, contentType string) (*http.Response, error) {

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %w", method, err)
	}
	req.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending %s request: %w", method, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		debugBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewReader(debugBody))
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
		if TOOL_CONFIGS.Logs.LogRequestPayloads {
			PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Request body: %s", string(body)))
		}
		PrintLog(LogLevelDebug, NoResource, "", fmt.Sprintf("Response [%d]: %s", resp.StatusCode, string(debugBody)))
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
		if NotificationTemplatesApiExists {
			return "notification/email/template-types"
		}
		return "email/template-types"
	case SCRIPT_LIBRARIES:
		return "script-libraries"
	case GOVERNANCE_CONNECTORS:
		return "identity-governance"
	case CERTIFICATES:
		return "keystores/certs"
	case WORKFLOWS:
		return "workflows"
	case WORKFLOW_ASSOCIATIONS:
		return "workflow-associations"
	case API_RESOURCES:
		return "api-resources"
	case VALIDATION_RULES:
		return "validation-rules"
	case ORGANIZATIONS:
		return "organizations"
	case EMAIL_PROVIDERS:
		return "notification-senders/email"
	case SMS_PROVIDERS:
		return "notification-senders/sms"
	case SMS_TEMPLATES:
		return "notification/sms/template-types"
	case ACTIONS:
		return "actions"
	case BRANDING_PREFERENCES:
		return "branding-preference"
	case CUSTOM_TEXTS:
		return "branding-preference/text"
	case FLOWS:
		return "flow"
	}
	return ""
}

func GetTenantBaseUrl() string {

	basePath := "/t/" + SERVER_CONFIGS.TenantDomain
	if IsSubOrganization() {
		basePath += "/o"
	}
	return SERVER_CONFIGS.ServerUrl + basePath
}

func getResourceBaseUrl(resourceType ResourceType) string {

	base := GetTenantBaseUrl()
	switch resourceType {
	case ROLES:
		if RolesV2ApiExists {
			return base + "/scim2/v2/Roles/"
		}
		return base + "/scim2/Roles/"
	case EMAIL_PROVIDERS, SMS_PROVIDERS:
		return base + "/api/server/v2/" + getResourcePath(resourceType) + "/"
	default:
		return base + "/api/server/v1/" + getResourcePath(resourceType) + "/"
	}
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
	return strings.TrimSuffix(reqUrl, "/")
}

func addQueryParams(reqURL string, resourceType ResourceType, operation string) string {

	url, err := url.Parse(reqURL)
	if err != nil {
		PrintLog(LogLevelError, NoResource, "", fmt.Sprintf("Failed to parse URL: %s. Unable to add query parameters.", err))
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
			queryParams.Set("excludedAttributes", "meta,users,groups,associatedApplications")
		}
	case CERTIFICATES:
		if operation == GET {
			queryParams.Set("encode-cert", "true")
		}
	}

	url.RawQuery = queryParams.Encode()
	return url.String()
}

func IsResourceNotFound(err error) bool {

	return err != nil && strings.Contains(err.Error(), ErrorCodes[404])
}

func extractIntField(m map[string]interface{}, field string) (int, error) {

	val, ok := m[field]
	if !ok {
		return 0, fmt.Errorf("field %q not found in response", field)
	}
	f, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for field %q", field)
	}
	return int(f), nil
}

func extractResultsArray(m map[string]interface{}, field string) (results []interface{}, exists bool, err error) {

	val, ok := m[field]
	if !ok {
		return []interface{}{}, false, nil
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil, true, fmt.Errorf("unexpected format for results field %q", field)
	}
	return arr, true, nil
}
