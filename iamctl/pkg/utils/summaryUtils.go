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
)

type Summary struct {
	SuccessfulOperations int
	FailedOperations     int
	TotalRequests        int
}

type ResourceSummary struct {
	ResourceType          string
	SuccessfulExport      int
	FailedExport          int
	SuccessfulImport      int
	FailedImport          int
	SuccessfulUpdate      int
	FailedUpdate          int
	NewSecretApplications []string
}

var (
	SummaryData       Summary
	ResourceSummaries map[string]ResourceSummary
)

func PrintSummary() {

	fmt.Println("----------------------------------------")
	fmt.Println("Total Summary:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Total Requests: %d\n", SummaryData.TotalRequests)
	fmt.Printf("Successful Operations: %d\n", SummaryData.SuccessfulOperations)
	fmt.Printf("Failed Operations: %d\n", SummaryData.FailedOperations)
}

func PrintExportSummary() {

	for _, summary := range ResourceSummaries {
		fmt.Println("----------------------------------------")
		fmt.Printf("%s:\n", summary.ResourceType)
		fmt.Println("----------------------------------------")
		fmt.Printf("Successful Exports: %d\n", summary.SuccessfulExport)
		fmt.Printf("Failed Exports: %d\n", summary.FailedExport)
	}
	fmt.Println("----------------------------------------")
}

func PrintImportSummary() {

	for _, summary := range ResourceSummaries {
		fmt.Println("----------------------------------------")
		fmt.Printf("%s:\n", summary.ResourceType)
		fmt.Println("----------------------------------------")
		fmt.Printf("Successful Imports: %d\n", summary.SuccessfulImport)
		fmt.Printf("Failed Imports: %d\n", summary.FailedImport)
		if summary.ResourceType == APPLICATIONS {
			printNewSecretApplications(summary)
		}
		fmt.Printf("Successful Updates: %d\n", summary.SuccessfulUpdate)
		fmt.Printf("Failed Updates: %d\n", summary.FailedUpdate)
	}
	fmt.Println("----------------------------------------")
}

func printNewSecretApplications(summary ResourceSummary) {

	if len(summary.NewSecretApplications) > 0 {
		fmt.Print("New Client Secrets generated for: ")
		for i, appName := range summary.NewSecretApplications {
			if i != len(summary.NewSecretApplications)-1 {
				fmt.Printf("%s, ", appName)
			} else {
				fmt.Print(appName)
			}
		}
		fmt.Println()
	}
}

func AddNewSecretApplication(appName string) {

	summary, ok := ResourceSummaries[APPLICATIONS]
	if !ok {
		summary = ResourceSummary{
			ResourceType: APPLICATIONS,
		}
	}
	summary.NewSecretApplications = append(summary.NewSecretApplications, appName)
	ResourceSummaries[APPLICATIONS] = summary
}

func UpdateSummary(success bool, resourceType string, operation string) {

	SummaryData.TotalRequests++

	if _, ok := ResourceSummaries[resourceType]; !ok {
		ResourceSummaries[resourceType] = ResourceSummary{
			ResourceType: resourceType,
		}
	}

	switch operation {
	case EXPORT:
		exportSummary := ResourceSummaries[resourceType]
		if success {
			exportSummary.SuccessfulExport++
		} else {
			exportSummary.FailedExport++
		}
		ResourceSummaries[resourceType] = exportSummary

	case IMPORT:
		importSummary := ResourceSummaries[resourceType]
		if success {
			importSummary.SuccessfulImport++
		} else {
			importSummary.FailedImport++
		}
		ResourceSummaries[resourceType] = importSummary

	case UPDATE:
		updateSummary := ResourceSummaries[resourceType]
		if success {
			updateSummary.SuccessfulUpdate++
		} else {
			updateSummary.FailedUpdate++
		}
		ResourceSummaries[resourceType] = updateSummary
	}
}
