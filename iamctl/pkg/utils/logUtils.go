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
	"fmt"
	"log"
	"strings"
)

type Summary struct {
	SuccessfulOperations int
	FailedOperations     int
	TotalRequests        int
}

type ResourceSummary struct {
	ResourceType                string
	SuccessfulExport            int
	SuccessfulImport            int
	SuccessfulUpdate            int
	Failed                      int
	Deleted                     int
	SecretGeneratedApplications []string
	FailedResources             []string
}

var (
	SummaryData       Summary
	ResourceSummaries map[string]ResourceSummary
)

var CURRENT_LOG_LEVEL LogLevel = LogLevelInfo

func resolveLogLevel(levelStr string) LogLevel {

	switch strings.ToUpper(strings.TrimSpace(levelStr)) {
	case "DEBUG":
		return LogLevelDebug
	case "WARN":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

func levelPrefix(level LogLevel) string {

	switch level {
	case LogLevelDebug:
		return "DEBUG:"
	case LogLevelWarn:
		return "WARN:"
	case LogLevelError:
		return "ERROR:"
	default:
		return "INFO:"
	}
}

func PrintLog(level LogLevel, resourceType ResourceType, resourceName string, msg string) {

	if level < CURRENT_LOG_LEVEL {
		return
	}
	prefix := levelPrefix(level)

	if resourceType == NoResource {
		log.Printf("%s %s", prefix, msg)
	} else if resourceName == "" {
		log.Printf("%s %s - %s", prefix, resourceType, msg)
	} else {
		log.Printf("%s %s - %s - %s", prefix, resourceType, resourceName, msg)
	}
}

func PrintSummary(Operation string) {

	InitializeResourceSummary()

	fmt.Println("========================================")
	fmt.Println("Total Summary:")
	fmt.Println("========================================")
	fmt.Printf("Total Requests: %d\n", SummaryData.TotalRequests)
	fmt.Printf("Successful Operations: %d\n", SummaryData.SuccessfulOperations)
	fmt.Printf("Failed Operations: %d\n", SummaryData.FailedOperations)

	if Operation == IMPORT {
		PrintImportSummary()
	} else if Operation == EXPORT {
		PrintExportSummary()
	}
}

func PrintExportSummary() {

	for _, summary := range ResourceSummaries {
		fmt.Println("----------------------------------------")
		fmt.Printf("%s\n", summary.ResourceType)
		fmt.Println("----------------------------------------")
		fmt.Printf("Successful Exports: %d\n", summary.SuccessfulExport)

		if summary.Failed > 0 {
			PrintFailedResources(summary)
		}
	}
	fmt.Println("----------------------------------------")
}

func PrintImportSummary() {

	for _, summary := range ResourceSummaries {
		fmt.Println("----------------------------------------")
		fmt.Printf("%s\n", summary.ResourceType)
		fmt.Println("----------------------------------------")
		fmt.Printf("Successful Imports: %d\n", summary.SuccessfulImport)
		fmt.Printf("Successful Updates: %d\n", summary.SuccessfulUpdate)
		fmt.Printf("Deleted: %d\n", summary.Deleted)
		if summary.Failed > 0 {
			PrintFailedResources(summary)
		}
		if summary.ResourceType == APPLICATIONS.String() {
			printNewSecretApplications(summary)
		}
	}
	fmt.Println("----------------------------------------")
}

func PrintFailedResources(summary ResourceSummary) {

	fmt.Println("....................")
	fmt.Printf("Failures:  %d\n", summary.Failed)
	fmt.Println("....................")
	for i, resourceName := range summary.FailedResources {
		if i != len(summary.FailedResources)-1 {
			fmt.Printf("%s, ", resourceName)
		} else {
			fmt.Print(resourceName)
		}
	}

	fmt.Println()
}

func printNewSecretApplications(summary ResourceSummary) {

	if len(summary.SecretGeneratedApplications) > 0 {
		fmt.Println("....................")
		fmt.Printf("New Client Secrets generated for: ")
		for i, appName := range summary.SecretGeneratedApplications {
			if i != len(summary.SecretGeneratedApplications)-1 {
				fmt.Printf("%s, ", appName)
			} else {
				fmt.Print(appName)
			}
		}
		fmt.Println()
	}
}

func AddNewSecretIndicatorToSummary(appName string) {

	InitializeResourceSummary()

	summary, ok := ResourceSummaries[APPLICATIONS.String()]
	if !ok {
		summary = ResourceSummary{
			ResourceType: APPLICATIONS.String(),
		}
	}
	summary.SecretGeneratedApplications = append(summary.SecretGeneratedApplications, appName)
	ResourceSummaries[APPLICATIONS.String()] = summary
}

func UpdateSuccessSummary(resourceType ResourceType, operation string) {

	InitializeResourceSummary()

	SummaryData.TotalRequests++
	SummaryData.SuccessfulOperations++

	summary, ok := ResourceSummaries[resourceType.String()]
	if !ok {
		summary = ResourceSummary{
			ResourceType: resourceType.String(),
		}
	}
	switch operation {
	case EXPORT:
		summary.SuccessfulExport++
	case IMPORT:
		summary.SuccessfulImport++
	case UPDATE:
		summary.SuccessfulUpdate++
	case DELETE:
		summary.Deleted++
	}
	ResourceSummaries[resourceType.String()] = summary
}

func UpdateFailureSummary(resourceType ResourceType, resourceName string) {

	InitializeResourceSummary()

	SummaryData.TotalRequests++
	SummaryData.FailedOperations++

	summary, ok := ResourceSummaries[resourceType.String()]
	if !ok {
		summary = ResourceSummary{
			ResourceType: resourceType.String(),
		}
	}
	summary.Failed++
	summary.FailedResources = append(summary.FailedResources, resourceName)
	ResourceSummaries[resourceType.String()] = summary
}

func InitializeResourceSummary() {

	if ResourceSummaries == nil {
		ResourceSummaries = make(map[string]ResourceSummary)
	}
}
