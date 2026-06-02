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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (rt ResourceType) String() string {
	return string(rt)
}

func IsResourceExcluded(resourceName string, resourceConfigs map[string]interface{}) bool {

	// Include only the resources added to INCLUDE_ONLY config. Note: INCLUDE_ONLY config overrides the EXCLUDE config.
	includeOnlyResources, ok := resourceConfigs[INCLUDE_ONLY_CONFIG].([]interface{})
	if ok {
		for _, resource := range includeOnlyResources {
			if resource.(string) == resourceName {
				return false
			}
		}
		PrintLog(LogLevelInfo, UtilsResourceWrapper, "", fmt.Sprintf("Excluded resource: %s", resourceName))
		return true
	} else {
		// Exclude resources added to EXCLUDE config.
		resourcesToExclude, ok := resourceConfigs[EXCLUDE_CONFIG].([]interface{})
		if ok {
			for _, resource := range resourcesToExclude {
				if resource.(string) == resourceName {
					PrintLog(LogLevelInfo, UtilsResourceWrapper, "", fmt.Sprintf("Excluded resource: %s", resourceName))
					return true
				}
			}
		}
		return false
	}
}

func IsResourceTypeExcluded(resourceType ResourceType) bool {

	// Include only the resource types added to INCLUDE_ONLY config. Note: INCLUDE_ONLY config overrides the EXCLUDE config.
	if len(TOOL_CONFIGS.IncludeOnly) > 0 {
		for _, resource := range TOOL_CONFIGS.IncludeOnly {
			if resource == resourceType.String() {
				return false
			}
		}
		PrintLog(LogLevelInfo, resourceType, "", "Skipping excluded resource type")
		return true
	} else if len(TOOL_CONFIGS.Exclude) > 0 {
		// Exclude resource types added to EXCLUDE config.
		for _, resource := range TOOL_CONFIGS.Exclude {
			if resource == resourceType.String() {
				PrintLog(LogLevelInfo, resourceType, "", "Skipping excluded resource type")
				return true
			}
		}
	}
	return false
}

func IsEntitySupportedInOrg(resourceType ResourceType) bool {

	if !IsSubOrganization() {
		return true
	}
	if entitySupportedInSubOrg[resourceType] {
		return true
	}
	PrintLog(LogLevelInfo, resourceType, "", "Not supported for sub organizations")
	return false
}

func ShouldSkip(resourceType ResourceType) bool {

	if !IsEntitySupportedInVersion(resourceType) {
		UpdateSkipSummary(resourceType, "Not supported in server version")
		return true
	}
	if !IsEntitySupportedInOrg(resourceType) {
		UpdateSkipSummary(resourceType, "Not supported in sub-organizations")
		return true
	}
	if IsResourceTypeExcluded(resourceType) {
		UpdateSkipSummary(resourceType, "Excluded via tool configs")
		return true
	}
	return false
}

func ResolveAdvancedKeywordMapping(resourceName string, resourceConfigs map[string]interface{}) map[string]interface{} {

	defaultKeywordMapping := KEYWORD_CONFIGS.KeywordMappings

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

	// Check if secrets are excluded for all resources. Note: global config will be overridden by resource level config.
	return TOOL_CONFIGS.ExcludeSecrets
}

func RemoveDeletedLocalDirectories(parentDir string, deployedDirNames []string) {

	deployedNames := make(map[string]struct{})
	for _, name := range deployedDirNames {
		deployedNames[name] = struct{}{}
	}

	localEntries, err := ioutil.ReadDir(parentDir)
	if err != nil {
		PrintLog(LogLevelError, UtilsResourceWrapper, "", fmt.Sprintf("Error loading directory: %s", err))
		return
	}

	for _, entry := range localEntries {
		if !entry.IsDir() {
			continue
		}
		if _, exists := deployedNames[entry.Name()]; !exists {
			dirPath := filepath.Join(parentDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				PrintLog(LogLevelError, UtilsResourceWrapper, "", fmt.Sprintf("Error when removing the directory %s: %s", entry.Name(), err))
			} else {
				PrintLog(LogLevelInfo, UtilsResourceWrapper, "", fmt.Sprintf("Removed the directory: %s", entry.Name()))
			}
		}
	}
}

func RemoveDeletedLocalResources(filePath string, deployedResourceNames []string) {

	// Remove local files of resources that do not exist in the remote during export.
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		PrintLog(LogLevelError, UtilsResourceWrapper, "", fmt.Sprintf("Error loading local files: %s", err))
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if !Contains(deployedResourceNames, GetFileInfo(fileName).ResourceName) {
			err := os.Remove(filepath.Join(filePath, fileName))
			if err != nil {
				PrintLog(LogLevelError, UtilsResourceWrapper, "", fmt.Sprintf("Error when removing the file: %s %s", fileName, err))
			} else {
				PrintLog(LogLevelInfo, UtilsResourceWrapper, "", fmt.Sprintf("Removed the file: %s", fileName))
			}
		}
	}
}

func RemoveSecretMasks(modifiedFileData string) string {

	modifiedFileData = strings.ReplaceAll(modifiedFileData, SENSITIVE_FIELD_MASK, "null")
	modifiedFileData = strings.ReplaceAll(modifiedFileData, `"`+SENSITIVE_FIELD_MASK_WITHOUT_QUOTES+`"`, "null")
	return modifiedFileData
}
