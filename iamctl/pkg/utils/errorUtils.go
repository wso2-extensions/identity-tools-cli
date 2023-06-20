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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Code             string            `json:"code"`
	Message          string            `json:"message"`
	Description      string            `json:"description"`
	TraceID          string            `json:"traceId"`
	FailedOperations []FailedOperation `json:"failedOperations"`
}

type FailedOperation struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	ClaimURI    string `json:"claimURI,omitempty"`
}

func handleClaimImportErrorResponse(resp *http.Response) error {

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err.Error())
	}

	errorResponse := ErrorResponse{}
	err = json.Unmarshal(responseBody, &errorResponse)
	if err != nil {
		return fmt.Errorf("failed to parse error response: %s", err.Error())
	}

	errorMessages := collectFailedOperations(errorResponse.FailedOperations)

	return fmt.Errorf("error response for the import request: %s\n%s", errorResponse.Message, strings.Join(errorMessages, "\n"))
}

func collectFailedOperations(failedOperations []FailedOperation) []string {

	errorMessages := make([]string, 0, len(failedOperations))

	for _, failedOp := range failedOperations {
		message := strings.TrimSpace(failedOp.Message)
		message = strings.TrimSuffix(message, ".")
		errorMessage := fmt.Sprintf("%s for %s", message, failedOp.ClaimURI)
		errorMessages = append(errorMessages, errorMessage)
	}

	return errorMessages
}
