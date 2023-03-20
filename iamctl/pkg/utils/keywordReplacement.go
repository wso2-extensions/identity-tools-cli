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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func ReplaceKeywords(fileData []byte, appName string) string {

	appKeywordMap := getAppKeywordMappings(appName)

	re := regexp.MustCompile(`\${([^}]+)}`)
	output := re.ReplaceAllStringFunc(string(fileData), func(match string) string {
		keyword := strings.Trim(match, "${}")
		if val, ok := appKeywordMap[keyword]; ok {
			if val, ok := val.(string); ok {
				return string(val)
			} else {
				log.Printf("Keyword %s is not a string", keyword)
			}
		} else {
			log.Printf("Removing the keyword %s from exported file, since it is not defined in the configs.", keyword)
		}
		return ""
	})
	return output
}

func getAppKeywordMappings(appName string) (keywordMappings map[string]interface{}) {

	keywordMappings = TOOL_CONFIGS.KeywordMappings
	if TOOL_CONFIGS.ApplicationConfigs != nil {
		if appConfigs, ok := TOOL_CONFIGS.ApplicationConfigs[appName]; ok {
			if appKeywordMappings, ok := appConfigs.(map[string]interface{})["KEYWORD_MAPPINGS"]; ok {
				mergedKeywordMap := make(map[string]interface{})
				for key, value := range keywordMappings {
					mergedKeywordMap[key] = value.(string)
				}
				for key, value := range appKeywordMappings.(map[string]interface{}) {
					mergedKeywordMap[key] = value.(string)
				}
				return mergedKeywordMap
			}
		}
	}
	return keywordMappings
}

// Functions for keyword replacement during export

func AddKeywords(exportedData []byte, localFilePath string, appName string) []byte {

	// Load local file data as a yaml object
	localFileData, err := loadYAMLFile(localFilePath)
	if err != nil {
		log.Printf("Local file %s is not available. Exported data will not be modified.", localFilePath)
	}

	// If the local file is empty or not available, return the exported data as it is.
	if localFileData != nil {
		// Get keyword locations in local file
		keywordLocations := getKeywordLocations(localFileData, []string{})

		// Load exported app data as a yaml object
		var exportedYaml interface{}
		err = yaml.Unmarshal(exportedData, &exportedYaml)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		appKeywordMap := getAppKeywordMappings(appName)

		// Compare the fields with keywords in the exported file and the local file and modify the exported file
		exportedYaml = modifyFieldsWithKeywords(localFileData, exportedYaml, keywordLocations, appKeywordMap)

		exportedData, err = yaml.Marshal(exportedYaml)
		if err != nil {
			panic(err)
		}
	}
	return exportedData
}

func loadYAMLFile(filename string) (interface{}, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var data interface{}
	err = yaml.Unmarshal(yamlFile, &data)

	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	return data, nil
}

func getKeywordLocations(fileData interface{}, path []string) map[string][]string {

	var keys = make(map[string][]string)
	switch v := fileData.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			newPath := append(path, fmt.Sprintf("%v", k))
			for path, key := range getKeywordLocations(val, newPath) {
				keys[path] = key
			}
		}
	case []interface{}:
		for _, val := range v {
			newPath := append(path, resolvePathWithIdentifiers(path[len(path)-1], val, arrayIdentifiers))
			for path, key := range getKeywordLocations(val, newPath) {
				keys[path] = key
			}
		}
	case string:
		containKeywords, keywords := getKeywords(fileData.(string))
		if containKeywords {
			thisPath := strings.Join(path, ".")
			for _, match := range keywords {
				keys[thisPath] = append(keys[thisPath], match[1])
			}
		}
	}
	return keys
}

func resolvePathWithIdentifiers(arrayName string, element interface{}, identifiers map[string]string) string {

	elementMap, ok := element.(map[interface{}]interface{})

	if !ok {
		fmt.Printf("cannot convert %T to map[string]string", element)
	}
	identifier := identifiers[arrayName]
	// If an identifier is not defined for the array, use the default identifier "name".
	if identifier == "" {
		identifier = "name"
	}
	identifierValue := elementMap[identifier]
	// TODO: Handle the case where the identifier value is a path
	// identifierValue := GetValue(elementMap, identifier)
	// TODO: Handle the case where the identifier value is empty
	return fmt.Sprintf("[%s=%s]", identifier, identifierValue)
}

func modifyFieldsWithKeywords(localFileData interface{}, exportedFileData interface{}, keywordLocations map[string][]string, keywordMap map[string]interface{}) interface{} {

	for location, keyword := range keywordLocations {

		localValue := GetValue(localFileData, location)
		localReplacedValue := replaceKeywords(localValue, keyword, keywordMap)
		exportedValue := GetValue(exportedFileData, location)

		if exportedValue != localReplacedValue {
			log.Printf("Warning: Keywords at %s field in the local file will be replaced by exported content.", location)
			log.Println("Local Value with Keyword Replaced: ", localReplacedValue)
			log.Println("Exported Value: ", exportedValue)
		} else {
			log.Printf("Keyword added at %s field\n", location)
			ReplaceValue(exportedFileData, location, localValue)
		}
	}
	// fmt.Println("Exported Value: ", exportedFileData)
	return exportedFileData
}

func GetValue(data interface{}, key string) string {

	parts := strings.Split(key, ".")
	for _, part := range parts {
		switch v := data.(type) {
		case map[interface{}]interface{}:
			data = v[part]
		case map[string]interface{}:
			data = v[part]
		case []interface{}:
			index, err := getArrayIndex(v, part)
			if err == nil {
				if len(v) > index {
					data = v[index]
				}
			}
		default:
			return ""
		}
	}
	if reflect.TypeOf(data).Kind() == reflect.Int {
		return strconv.Itoa(data.(int))
	}
	if data == nil {
		return ""
	}
	return data.(string)
}

func ReplaceValue(data interface{}, key string, replacement string) interface{} {

	parts := strings.Split(key, ".")
	for i, part := range parts {
		if i == len(parts)-1 {
			data.(map[interface{}]interface{})[part] = replacement
		} else {
			switch v := data.(type) {
			case map[interface{}]interface{}:
				data = v[part]
			case map[string]interface{}:
				data = v[part]
			case []interface{}:
				index, err := getArrayIndex(v, part)
				if err == nil {
					if len(v) > index {
						data = v[index]
					}
				}
			default:
				return nil
			}
		}
	}
	return data
}

func getKeywords(data string) (bool, [][]string) {

	re := regexp.MustCompile(`\${([^}]+)}`)
	matches := re.FindAllStringSubmatch(data, -1)
	return (len(matches) > 0), matches
}

func replaceKeywords(data string, keywords []string, keywordMapping map[string]interface{}) (replacedData string) {

	replacedData = data
	for _, keyword := range keywords {
		replacedData = strings.ReplaceAll(replacedData, "${"+keyword+"}", keywordMapping[keyword].(string))
	}
	return replacedData
}

func getArrayIndex(arrayMap []interface{}, elementIdentifier string) (int, error) {

	for k, v := range arrayMap {
		if strings.HasPrefix(elementIdentifier, "[") && strings.HasSuffix(elementIdentifier, "]") {
			identifier := elementIdentifier[1 : len(elementIdentifier)-1]
			parts := strings.SplitN(identifier, "=", 2)
			if v.(map[interface{}]interface{})[parts[0]] == parts[1] {
				return k, nil
			}
		} else {
			fmt.Println("String is not in the expected format")
		}
	}
	return -1, errors.New("Element not found")
}
