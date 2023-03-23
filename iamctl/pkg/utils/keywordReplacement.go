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
	"log"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func ReplaceKeywords(fileContent string, keywordMapping map[string]interface{}) string {

	// Loop over the keyword mapping and replace each keyword in the file.
	for keyword, value := range keywordMapping {
		if value, ok := value.(string); ok {
			fileContent = strings.ReplaceAll(fileContent, "{{"+keyword+"}}", value)
		} else {
			log.Printf("Keyword value for %s is not a string", keyword)
		}
	}
	return fileContent
}

// Functions for keyword replacement during export

func AddKeywords(exportedData []byte, localFileData []byte, keywordMapping map[string]interface{}) []byte {

	var localYaml interface{}
	err := yaml.Unmarshal(localFileData, &localYaml)
	if err != nil {
		log.Println("Error: ", err)
	}
	// If the local file is empty or not available, return the exported data as it is.
	if err == nil && localYaml != nil {
		// Get keyword locations in local file
		keywordLocations := GetKeywordLocations(localYaml, []string{}, keywordMapping)
		// Load exported app data as a yaml object
		var exportedYaml interface{}
		err = yaml.Unmarshal(exportedData, &exportedYaml)
		if err != nil {
			log.Println("Error: ", err)
		}

		// Compare the fields with keywords in the exported file and the local file and modify the exported file
		exportedYaml = ModifyFieldsWithKeywords(exportedYaml, localYaml, keywordLocations, keywordMapping)

		exportedData, err = yaml.Marshal(exportedYaml)
		if err != nil {
			panic(err)
		}
	}
	return exportedData
}

// func LoadYAMLFile(filename string) (interface{}, error) {
// 	yamlFile, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		log.Println("Error in loading file: ", filename, err)
// 		return nil, err
// 	}
// 	fmt.Println("YAML file content: ", string(yamlFile))
// 	var data interface{}
// 	err = yaml.Unmarshal(yamlFile, &data)
// 	if err != nil {
// 		fmt.Println("Error when loading YAML content from file: ", filename, err)
// 		return nil, err
// 	}
// 	return data, nil
// }

func GetKeywordLocations(fileData interface{}, path []string, keywordMapping map[string]interface{}) []string {

	var keys []string
	switch v := fileData.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			newPath := append(path, fmt.Sprintf("%v", k))
			keys = append(keys, GetKeywordLocations(val, newPath, keywordMapping)...)
		}
	case []interface{}:
		for _, val := range v {
			arrayElementPath, err := resolvePathWithIdentifiers(path[len(path)-1], val, arrayIdentifiers)
			if err != nil {
				log.Println("Error: ", err)
				break
			} else {
				newPath := append(path, arrayElementPath)
				keys = append(keys, GetKeywordLocations(val, newPath, keywordMapping)...)
			}
		}
	case string:
		if ContainsKeywords(fileData.(string), keywordMapping) {
			thisPath := strings.Join(path, ".")
			keys = append(keys, thisPath)
		}
	}
	return keys
}

func resolvePathWithIdentifiers(arrayName string, element interface{}, identifiers map[string]string) (string, error) {

	elementMap, ok := element.(map[interface{}]interface{})
	if !ok {
		log.Printf("cannot convert %T to a map", element)
	}
	identifier := identifiers[arrayName]
	// If an identifier is not defined for the array, use the default identifier "name".
	if identifier == "" {
		identifier = "name"
	}
	identifierValue, ok := elementMap[identifier]
	if !ok {
		return "", fmt.Errorf("identifier not found for the array %s", arrayName)
	}
	return fmt.Sprintf("[%s=%s]", identifier, identifierValue), nil
}

func ContainsKeywords(data string, keywordMapping map[string]interface{}) bool {

	for keyword, _ := range keywordMapping {
		if strings.Contains(data, "{{"+keyword+"}}") {
			return true
		}
	}
	return false
}

func ModifyFieldsWithKeywords(exportedFileData interface{}, localFileData interface{}, keywordLocations []string, keywordMap map[string]interface{}) interface{} {

	for _, location := range keywordLocations {

		localValue := GetValue(localFileData, location)
		localReplacedValue := ReplaceKeywords(localValue, keywordMap)
		exportedValue := GetValue(exportedFileData, location)

		if exportedValue != localReplacedValue {
			log.Printf("Warning: Keywords at %s field in the local file will be replaced by exported content.", location)
			log.Println("Local Value with Keyword Replaced: ", localReplacedValue)
			log.Println("Exported Value: ", exportedValue)
		} else {
			log.Printf("Keyword added at %s field\n", location)
			ReplaceValue(exportedFileData, strings.Split(location, "."), localValue)
		}
	}
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
			index, err := GetArrayIndex(v, part)
			if err != nil {
				return ""
			}
			if len(v) > index {
				data = v[index]
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

func ReplaceValue(data interface{}, path []string, replacement string) interface{} {

	if len(path) == 1 {
		switch data.(type) {
		case map[interface{}]interface{}:
			data.(map[interface{}]interface{})[path[0]] = replacement
		case map[string]interface{}:
			data.(map[string]interface{})[path[0]] = replacement
		}
	} else {
		switch v := data.(type) {
		case map[interface{}]interface{}:
			currentKey := path[0]
			data.(map[interface{}]interface{})[currentKey] = ReplaceValue(v[currentKey], path[1:], replacement)
		case map[string]interface{}:
			currentKey := path[0]
			data.(map[string]interface{})[currentKey] = ReplaceValue(v[currentKey], path[1:], replacement)
		case []interface{}:
			currentKey := path[0]
			index, err := GetArrayIndex(v, currentKey)
			if err != nil {
				return data
			}
			if len(v) > index {
				data.([]interface{})[index] = ReplaceValue(v[index], path[1:], replacement)
			}
		default:
			return data
		}
	}
	return data
}

func GetArrayIndex(arrayMap []interface{}, elementIdentifier string) (int, error) {

	if strings.HasPrefix(elementIdentifier, "[") && strings.HasSuffix(elementIdentifier, "]") {
		identifier := elementIdentifier[1 : len(elementIdentifier)-1]
		parts := strings.SplitN(identifier, "=", 2)
		for k, v := range arrayMap {
			switch v := v.(type) {
			case map[interface{}]interface{}:
				if v[parts[0]] == parts[1] {
					return k, nil
				}
			case map[string]interface{}:
				if v[parts[0]] == parts[1] {
					return k, nil
				}
			}
		}
	} else {
		log.Println("Element identifier is not in the expected format")
	}
	return -1, errors.New("element not found")
}
