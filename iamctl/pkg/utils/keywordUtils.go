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
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"gopkg.in/yaml.v2"
)

func ReplaceKeywords(fileContent string, keywordMapping map[string]interface{}) string {

	// Loop over the keyword mapping and replace each keyword in the file.
	for keyword, value := range keywordMapping {
		if value, ok := value.(string); ok {
			fileContent = strings.ReplaceAll(fileContent, "{{"+keyword+"}}", value)
		} else {
			log.Printf("Error: keyword value for %s is not a string", keyword)
		}
	}
	return fileContent
}

func ProcessExportedContent(exportedFileName string, exportedFileContent []byte, keywordMapping map[string]interface{}, resourceType string) ([]byte, error) {

	// To preserve type tags in the exported file, replace the type tags with a placeholder.
	exportedFileContent = ReplaceTypeTags(exportedFileContent)

	// Unmarshall and marshall exported content in all cases to preserve a consistent format in the output.
	var exportedYaml interface{}
	err := yaml.Unmarshal(exportedFileContent, &exportedYaml)
	if err != nil {
		err1 := fmt.Errorf("error when parsing exported data to YAML. %w", err)
		return nil, err1
	}

	// Replace ESVs in the exported file according to the keyword placeholders added in the local file.
	var modifiedExportedYaml interface{}
	localFileData, err := ioutil.ReadFile(exportedFileName)
	if err != nil {
		log.Printf("Local file not found at %s. Creating new file.", exportedFileName)
		modifiedExportedYaml = exportedYaml
	} else {
		modifiedExportedYaml, err = AddKeywords(exportedYaml, localFileData, keywordMapping, resourceType)
		if err != nil {
			log.Println("Error when adding keywords to the exported file. Overriding local file with exported content. ", err)
		}
	}

	modifiedExportedContent, err := yaml.Marshal(modifiedExportedYaml)
	if err != nil {
		err1 := fmt.Errorf("error when creating exported data with keywords. %w", err)
		return nil, err1
	}
	modifiedExportedContent = AddTypeTags(modifiedExportedContent)
	return modifiedExportedContent, nil
}

func AddKeywords(exportedYaml interface{}, localFileData []byte, keywordMapping map[string]interface{}, resourceType string) (interface{}, error) {

	var localYaml interface{}
	err := yaml.Unmarshal(localFileData, &localYaml)
	if err != nil || localYaml == nil {
		err1 := fmt.Errorf("empty or invalid local file data. %w", err)
		return exportedYaml, err1
	}

	// Get keyword locations in local file.
	keywordLocations := GetKeywordLocations(localYaml, []string{}, keywordMapping, resourceType)

	// Compare the fields with keywords in the exported file and the local file and modify the exported file.
	exportedYaml = ModifyFieldsWithKeywords(exportedYaml, localYaml, keywordLocations, keywordMapping)

	return exportedYaml, nil
}

func GetKeywordLocations(fileData interface{}, path []string, keywordMapping map[string]interface{}, resourceType string) []string {

	var keys []string
	switch v := fileData.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			newPath := append(path, fmt.Sprintf("%v", k))
			keys = append(keys, GetKeywordLocations(val, newPath, keywordMapping, resourceType)...)
		}
	case map[string]interface{}:
		for k, val := range v {
			newPath := append(path, fmt.Sprintf("%v", k))
			keys = append(keys, GetKeywordLocations(val, newPath, keywordMapping, resourceType)...)
		}
	case []interface{}:
		for _, val := range v {
			if _, ok := val.(string); ok {
				if ContainsKeywords(val.(string), keywordMapping) {
					thisPath := strings.Join(path, ".")
					keys = append(keys, thisPath)
				}
				break
			} else {
				arrayIdentifiers := GetArrayIdentifiers(resourceType)
				arrayElementPath, err := resolvePathWithIdentifiers(path[len(path)-1], val, arrayIdentifiers)
				if err != nil {
					log.Printf("Error: cannot resolve path for the field %s. %s.\n", strings.Join(path, "."), err)
					break
				}
				newPath := append(path, arrayElementPath)
				keys = append(keys, GetKeywordLocations(val, newPath, keywordMapping, resourceType)...)
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

func GetArrayIdentifiers(resourceType string) map[string]string {

	switch resourceType {
	case configs.APPLICATIONS:
		return applicationArrayIdentifiers
	case configs.IDENTITY_PROVIDERS:
		return idpArrayIdentifiers
	case configs.USERSTORES:
		return userStoreArrayIdentifiers
	case configs.CLAIMS:
		return claimArrayIdentifiers
	}
	return make(map[string]string)
}

func resolvePathWithIdentifiers(arrayName string, element interface{}, identifiers map[string]string) (string, error) {

	var elementMap interface{}
	elementMap, ok := element.(map[interface{}]interface{})
	if !ok {
		elementMap, ok = element.(map[string]interface{})
		if !ok {
			log.Printf("Error: cannot convert %T to a map", element)
		}
	}
	identifier := identifiers[arrayName]

	// If an identifier is not defined for the array, use the default identifier "name".
	if identifier == "" {
		identifier = "name"
	}
	identifierValue := GetValue(elementMap, identifier)
	if identifierValue == "" {
		return identifierValue, fmt.Errorf("identifier not found for array %s", arrayName)
	}
	return fmt.Sprintf("[%s=%s]", identifier, identifierValue), nil
}

func ContainsKeywords(data string, keywordMapping map[string]interface{}) bool {

	for keyword := range keywordMapping {
		if strings.Contains(data, "{{"+keyword+"}}") {
			return true
		}
	}
	return false
}

func ModifyFieldsWithKeywords(exportedFileData interface{}, localFileData interface{},
	keywordLocations []string, keywordMap map[string]interface{}) interface{} {

	for _, location := range keywordLocations {

		localValue := GetValue(localFileData, location)
		localReplacedValue := ReplaceKeywords(localValue, keywordMap)
		exportedValue := GetValue(exportedFileData, location)

		if exportedValue != localReplacedValue {
			if exportedValue == strings.ReplaceAll(SENSITIVE_FIELD_MASK, "'", "") {
				ReplaceValue(exportedFileData, location, localValue)
				log.Printf("Info: Keyword added at %s field\n", location)
			} else {
				log.Printf("Warning: Keywords at %s field in the local file will be replaced by exported content.", location)
				log.Println("Info: Local Value with Keyword Replaced: ", localReplacedValue)
				log.Println("Info: Exported Value: ", exportedValue)
			}
		} else {
			ReplaceValue(exportedFileData, location, localValue)
			log.Printf("Info: Keyword added at %s field\n", location)
		}
	}
	return exportedFileData
}

func GetValue(data interface{}, key string) string {

	parts := GetPathKeys(key)
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
	if data == nil {
		return ""
	}
	if reflect.TypeOf(data).Kind() == reflect.Int {
		return strconv.Itoa(data.(int))
	}
	if finalArray, ok := data.([]interface{}); ok {
		strArray := make([]string, len(finalArray))
		for i, v := range finalArray {
			strArray[i] = v.(string)
		}
		data = strings.Join(strArray, ",")
	}
	return data.(string)
}

func ReplaceValue(data interface{}, pathString string, replacement string) interface{} {

	path := GetPathKeys(pathString)
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
			data.(map[interface{}]interface{})[currentKey] = ReplaceValue(v[currentKey], strings.Join(path[1:], "."), replacement)
		case map[string]interface{}:
			currentKey := path[0]
			data.(map[string]interface{})[currentKey] = ReplaceValue(v[currentKey], strings.Join(path[1:], "."), replacement)
		case []interface{}:
			currentKey := path[0]
			index, err := GetArrayIndex(v, currentKey)
			if err != nil {
				log.Printf("Error: when resolving array index for element %s.", currentKey)
				return data
			}
			if len(v) > index {
				data.([]interface{})[index] = ReplaceValue(v[index], strings.Join(path[1:], "."), replacement)
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
				if GetValue(v, parts[0]) == parts[1] {
					return k, nil
				}
			case map[string]interface{}:
				if GetValue(v, parts[0]) == parts[1] {
					return k, nil
				}
			}
		}
	} else {
		return -1, fmt.Errorf("element identifier is not in the expected format")
	}
	return -1, errors.New("element not found")
}

func GetPathKeys(pathString string) []string {

	pathArray := strings.Split(pathString, ".")
	finalKeys := []string{}
	for i := 0; i < len(pathArray); i++ {
		v := pathArray[i]
		if !strings.HasPrefix(v, "[") {
			finalKeys = append(finalKeys, v)
		} else {
			key := v
			var j int
			for j = i; j < (len(pathArray) - 1); j++ {

				if strings.HasSuffix(pathArray[j], "]") {
					break
				}
				key = key + "." + pathArray[j+1]
				i += 1
			}
			finalKeys = append(finalKeys, key)
		}
	}
	return finalKeys
}

func ReplaceTypeTags(data []byte) []byte {

	re := regexp.MustCompile(`!!org\.wso2\.`)
	data = re.ReplaceAll(data, []byte("1typeTag: "))

	re = regexp.MustCompile(`inboundConfigurationProtocol: 1typeTag: `)
	return re.ReplaceAll(data, []byte("inboundConfigurationProtocol:\n      1typeTag: "))
}

func AddTypeTags(data []byte) []byte {

	re := regexp.MustCompile(`1typeTag: `)
	return re.ReplaceAll(data, []byte("!!org.wso2."))
}

func ReplacePlaceholders(configFile []byte) []byte {

	configStr := string(configFile)

	for _, value := range os.Environ() {
		pair := strings.SplitN(value, "=", 2)
		envVarName := fmt.Sprintf("${%s}", pair[0])
		envVarValue := pair[1]
		configStr = strings.ReplaceAll(configStr, envVarName, envVarValue)
	}
	return []byte(configStr)
}
