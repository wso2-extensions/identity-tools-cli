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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func IsResourceExcluded(resourceName string, resourceConfigs map[string]interface{}) bool {

	// Include only the resources added to INCLUDE_ONLY config. Note: INCLUDE_ONLY config overrides the EXCLUDE config.
	includeOnlyResources, ok := resourceConfigs[INCLUDE_ONLY_CONFIG].([]interface{})
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
		resourcesToExclude, ok := resourceConfigs[EXCLUDE_CONFIG].([]interface{})
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

func ResolveAdvancedKeywordMapping(resourceName string, resourceConfigs map[string]interface{}) map[string]interface{} {

	defaultKeywordMapping := TOOL_CONFIGS.KeywordMappings

	// Check if resource specific configs exist for the given resource and if not return the default keyword mappings.
	if resourceSpecificConfigs, ok := resourceConfigs[resourceName]; ok {
		// Check if advanced keyword mappings exist for the given resource.
		if resourceKeywordMap, ok := resourceSpecificConfigs.(map[string]interface{})[KEYWORD_MAPPINGS_CONFIG].(map[string]interface{}); ok {

			mergedKeywordMap := make(map[string]interface{})
			for key, value := range defaultKeywordMapping {
				mergedKeywordMap[key] = value.(string)
			}
			// Override the default keyword mappings with the resource specific keyword mappings.
			for key, value := range resourceKeywordMap {
				mergedKeywordMap[key] = value.(string)
			}
			return mergedKeywordMap
		}
	}
	return defaultKeywordMapping
}

func AreSecretsExcluded(resourceConfigs map[string]interface{}) bool {

	// Check if secrets are excluded for the given resource type.
	if secretsExcluded, ok := resourceConfigs[EXCLUDE_SECRETS_CONFIG].(bool); ok {
		return secretsExcluded
	}
	return true
}

func RemoveDeletedLocalResources(filePath string, deployedResourceNames []string) {

	// Remove local files of resources that do not exist in the remote during export.
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		log.Println("Error loading local files: ", err)
		return
	}

	for _, file := range files {
		fileName := file.Name()
		if !Contains(deployedResourceNames, GetFileInfo(fileName).ResourceName) {
			err := os.Remove(filepath.Join(filePath, fileName))
			if err != nil {
				log.Println("Error when removing the file: ", fileName, err)
			} else {
				log.Println("Removed the file:", fileName)
			}
		}
	}
}
