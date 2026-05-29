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
	"time"
)

type Summary struct {
	SuccessfulOperations int
	FailedOperations     int
	TotalRequests        int
}

type ResourceTypeSummary struct {
	ResourceType                ResourceType
	SuccessfulExport            int
	SuccessfulImport            int
	SuccessfulUpdate            int
	FailedCount                 int
	DeletedCount                int
	SecretGeneratedApplications []string
	FailedResources             []string
	Skipped                     bool
	SkipReason                  string
	Failed                      bool
	Duration                    time.Duration
}

var (
	AggregatedSummary Summary
	ResTypeSummaryMap map[ResourceType]ResourceTypeSummary
	Warnings          []string
	ResTypeStartTimes = make(map[ResourceType]time.Time)
	StartTime         time.Time
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

func MarkResTypeStart(resourceType ResourceType) {

	ResTypeStartTimes[resourceType] = time.Now()
}

func MarkResTypeEnd(resourceType ResourceType) {

	startTime, ok := ResTypeStartTimes[resourceType]
	if !ok {
		return
	}
	InitializeResTypeSummaryMap()
	summary := getOrInitSummary(resourceType)
	summary.Duration = time.Since(startTime).Round(time.Millisecond)
	ResTypeSummaryMap[resourceType] = summary
}

func UpdateSkipSummary(resourceType ResourceType, reason string) {

	InitializeResTypeSummaryMap()
	summary := getOrInitSummary(resourceType)
	summary.Skipped = true
	summary.SkipReason = reason
	ResTypeSummaryMap[resourceType] = summary
}

func MarkResTypeFailure(resourceType ResourceType) {

	InitializeResTypeSummaryMap()
	summary := getOrInitSummary(resourceType)
	summary.Failed = true
	ResTypeSummaryMap[resourceType] = summary
}

func AddNewSecretIndicatorToSummary(appName string) {

	InitializeResTypeSummaryMap()
	summary := getOrInitSummary(APPLICATIONS)
	summary.SecretGeneratedApplications = append(summary.SecretGeneratedApplications, appName)
	ResTypeSummaryMap[APPLICATIONS] = summary
}

func UpdateSuccessSummary(resourceType ResourceType, operation string) {

	InitializeResTypeSummaryMap()

	AggregatedSummary.TotalRequests++
	AggregatedSummary.SuccessfulOperations++

	summary := getOrInitSummary(resourceType)
	switch operation {
	case EXPORT:
		summary.SuccessfulExport++
	case IMPORT:
		summary.SuccessfulImport++
	case UPDATE:
		summary.SuccessfulUpdate++
	case DELETE:
		summary.DeletedCount++
	}
	ResTypeSummaryMap[resourceType] = summary
}

func UpdateFailureSummary(resourceType ResourceType, resourceName string) {

	InitializeResTypeSummaryMap()

	AggregatedSummary.TotalRequests++
	AggregatedSummary.FailedOperations++

	summary := getOrInitSummary(resourceType)
	summary.FailedCount++
	summary.FailedResources = append(summary.FailedResources, resourceName)
	ResTypeSummaryMap[resourceType] = summary
}

func PrintLog(level LogLevel, resourceType ResourceType, resourceName string, msg string) {

	var body string
	if resourceType == NoResource {
		body = msg
	} else if resourceName == "" {
		body = fmt.Sprintf("%s - %s", resourceType, msg)
	} else {
		body = fmt.Sprintf("%s - %s - %s", resourceType, resourceName, msg)
	}

	if level == LogLevelWarn {
		Warnings = append(Warnings, body)
	}
	if level < CURRENT_LOG_LEVEL {
		return
	}
	log.Printf("%s %s", levelPrefix(level), body)
}

func PrintSummary(Operation string) {

	InitializeResTypeSummaryMap()

	type skippedEntry struct {
		name   string
		reason string
	}
	var successCount int
	var skippedTypes []skippedEntry
	var failedTypes []string

	for rt, summary := range ResTypeSummaryMap {
		if summary.Skipped {
			skippedTypes = append(skippedTypes, skippedEntry{rt.String(), summary.SkipReason})
		} else if summary.Failed || summary.FailedCount > 0 {
			failedTypes = append(failedTypes, rt.String())
		} else {
			successCount++
		}
	}

	fmt.Println("========================================")
	fmt.Println("Total Summary:")
	fmt.Println("========================================")
	fmt.Printf("Total Operations: %d\n", AggregatedSummary.TotalRequests)
	fmt.Printf("Successful Operations: %d\n", AggregatedSummary.SuccessfulOperations)
	fmt.Printf("Failed Operations: %d\n", AggregatedSummary.FailedOperations)
	fmt.Printf("Successful Resource Types: %d\n", successCount)
	fmt.Printf("Skipped Resource Types: %d\n", len(skippedTypes))
	fmt.Printf("Failed Resource Types: %d\n", len(failedTypes))
	if !StartTime.IsZero() {
		fmt.Printf("Total Execution Time: %s\n", time.Since(StartTime).Round(time.Millisecond))
	}
	if len(skippedTypes) > 0 {
		fmt.Println("========================================")
		fmt.Println("Skipped Resource Types")
		fmt.Println("========================================")
		for _, s := range skippedTypes {
			fmt.Printf("%s - %s\n", s.name, s.reason)
		}
	}
	if len(failedTypes) > 0 {
		fmt.Println("========================================")
		fmt.Println("Failed Resource Types")
		fmt.Println("========================================")
		for _, f := range failedTypes {
			fmt.Printf("%s\n", f)
		}
	}

	fmt.Println("========================================")
	fmt.Println("Per Resource Breakdown")
	fmt.Println("========================================")
	if Operation == IMPORT {
		PrintImportSummary()
	} else if Operation == EXPORT {
		PrintExportSummary()
	}

	if len(Warnings) > 0 {
		fmt.Println("========================================")
		fmt.Println("Warnings")
		fmt.Println("========================================")
		for _, w := range Warnings {
			fmt.Printf("%s\n", w)
		}
	}
	fmt.Println("========================================")
}

func PrintExportSummary() {

	first := true
	for _, summary := range ResTypeSummaryMap {
		if summary.SuccessfulExport+summary.FailedCount == 0 {
			continue
		}
		if !first {
			fmt.Println("----------------------------------------")
		}
		first = false
		fmt.Printf("%s\n", summary.ResourceType)
		fmt.Println("----------------------------------------")
		fmt.Printf("Successful Exports: %d\n", summary.SuccessfulExport)
		if summary.FailedCount > 0 {
			PrintFailedResources(summary)
		}
		if summary.Duration > 0 && summary.SuccessfulExport > 0 {
			fmt.Printf("Execution time: %s\n", summary.Duration.Round(time.Millisecond))
		}
	}
	fmt.Println("----------------------------------------")
}

func PrintImportSummary() {

	first := true
	for _, summary := range ResTypeSummaryMap {
		if summary.SuccessfulImport+summary.SuccessfulUpdate+summary.DeletedCount+summary.FailedCount == 0 {
			continue
		}
		if !first {
			fmt.Println("----------------------------------------")
		}
		first = false
		fmt.Printf("%s\n", summary.ResourceType)
		fmt.Println("----------------------------------------")
		fmt.Printf("Successful Imports: %d\n", summary.SuccessfulImport)
		fmt.Printf("Successful Updates: %d\n", summary.SuccessfulUpdate)
		fmt.Printf("Deleted: %d\n", summary.DeletedCount)
		if summary.FailedCount > 0 {
			PrintFailedResources(summary)
		}
		if summary.ResourceType == APPLICATIONS {
			printNewSecretApplications(summary)
		}
		if summary.Duration > 0 && summary.SuccessfulImport+summary.SuccessfulUpdate > 0 {
			fmt.Printf("Execution time: %s\n", summary.Duration.Round(time.Millisecond))
		}
	}
}

func PrintFailedResources(summary ResourceTypeSummary) {

	fmt.Println("....................")
	fmt.Printf("Failures:  %d\n", summary.FailedCount)
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

func printNewSecretApplications(summary ResourceTypeSummary) {

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

func InitializeResTypeSummaryMap() {

	if ResTypeSummaryMap == nil {
		ResTypeSummaryMap = make(map[ResourceType]ResourceTypeSummary)
	}
}

func getOrInitSummary(resourceType ResourceType) ResourceTypeSummary {

	summary, exists := ResTypeSummaryMap[resourceType]
	if !exists {
		return ResourceTypeSummary{ResourceType: resourceType}
	}
	return summary
}
