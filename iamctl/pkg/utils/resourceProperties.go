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
	"log"
)

func IsResourceExcluded(resourceName string, resourceConfigs map[string]interface{}) bool {

	// Include only the resources added to INCLUDE_ONLY config. Note: INCLUDE_ONLY config overrides the EXCLUDE config.
	includeOnlyResources, ok := resourceConfigs["INCLUDE_ONLY"].([]interface{})
	if ok {
		for _, resource := range includeOnlyResources {
			if resource.(string) == resourceName {
				return false
			}
		}
		log.Println("Excluded resource: " + resourceName)
		return true
	} else {
		// Exclude resources added to EXCLUDE config.
		resourcesToExclude, ok := resourceConfigs["EXCLUDE"].([]interface{})
		if ok {
			for _, resource := range resourcesToExclude {
				if resource.(string) == resourceName {
					log.Println("Excluded resource: " + resourceName)
					return true
				}
			}
		}
		return false
	}
}