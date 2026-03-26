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
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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

func ProcessExportedData(exportedData interface{}, localFilePath string, format Format, keywordMapping map[string]interface{}, resourceType ResourceType) (interface{}, error) {

	localFileContent, err := ioutil.ReadFile(localFilePath)
	if err != nil {
		log.Printf("Local file not found at %s. Creating new file.", localFilePath)
		return exportedData, nil
	}

	// Replace ESVs in the exported file according to the keyword placeholders added in the local file.
	modifiedData, err := AddLocalKeywords(exportedData, format, localFileContent, keywordMapping, resourceType)
	if err != nil {
		log.Println("Error processing keywords. Using exported content.", err)
		return exportedData, nil
	}

	return modifiedData, nil
}

func ProcessExportedContent(exportedFileName string, exportedFileContent []byte, keywordMapping map[string]interface{}, resourceType ResourceType) ([]byte, error) {

	format, err := FormatFromExtension(filepath.Ext(exportedFileName))
	if err != nil {
		return nil, fmt.Errorf("unsupported file format: %w", err)
	}
	if format == FormatYAML {
		// To preserve type tags in YAML files, replace the type tags with a placeholder.
		exportedFileContent = ReplaceTypeTags(exportedFileContent)
	}

	// Unmarshall and marshall exported content in all cases to preserve a consistent format in the output.
	exportedData, err := Deserialize(exportedFileContent, format, resourceType)
	if err != nil {
		return nil, fmt.Errorf("error when deserializing exported content: %w", err)
	}

	// Process exported content
	modifiedData, err := ProcessExportedData(exportedData, exportedFileName, format, keywordMapping, resourceType)
	if err != nil {
		log.Println("Error when processing with keywords. Using original exported content. ", err)
		modifiedData = exportedData
	}

	modifiedContent, err := Serialize(modifiedData, format, resourceType)
	if err != nil {
		return nil, fmt.Errorf("error when creating exported data with keywords: %w", err)
	}

	// Re-add YAML type tags for YAML files
	if format == FormatYAML {
		modifiedContent = AddTypeTags(modifiedContent)
	}
	return modifiedContent, nil
}

// Adds local keywords for exported content specifically in YAML format.
// Use AddLocalKeywords for granular control over different formats.
func AddKeywords(exportedYaml interface{}, localFileContent []byte, keywordMapping map[string]interface{}, resourceType ResourceType) (interface{}, error) {
	return AddLocalKeywords(exportedYaml, FormatYAML, localFileContent, keywordMapping, resourceType)
}

func AddLocalKeywords(exportedData interface{}, format Format, localFileContent []byte, keywordMapping map[string]interface{}, resourceType ResourceType) (interface{}, error) {

	localData, err := Deserialize(localFileContent, format, resourceType)
	if err != nil || localData == nil {
		return exportedData, fmt.Errorf("empty or invalid local file data: %w", err)
	}

	if preprocessFunc, exists := DataPreprocessFuncs[resourceType]; exists {
		localData, err = preprocessFunc(localData)
		if err != nil {
			return exportedData, fmt.Errorf("error occurred while preprocessing local data: %w", err)
		}
	}

	// Get keyword locations in local file.
	keywordLocations := GetKeywordLocations(localData, []string{}, keywordMapping, resourceType)

	// Compare the fields with keywords in the exported file and the local file and modify the exported file.
	exportedData = ModifyFieldsWithKeywords(exportedData, localData, keywordLocations, keywordMapping)

	return exportedData, nil
}

func GetKeywordLocations(fileData interface{}, path []string, keywordMapping map[string]interface{}, resourceType ResourceType) []string {

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

func GetArrayIdentifiers(resourceType ResourceType) map[string]string {

	switch resourceType {
	case APPLICATIONS:
		if ExportAPIExists(APPLICATIONS) {
			return appExportAPIArrayIdentifiers
		}
		return appGetAPIArrayIdentifiers
	case IDENTITY_PROVIDERS:
		if ExportAPIExists(IDENTITY_PROVIDERS) {
			return idpExportAPIArrayIdentifiers
		}
		return idpGetAPIArrayIdentifiers
	case USERSTORES:
		return userStoreArrayIdentifiers
	case CLAIMS:
		return claimArrayIdentifiers
	case CHALLENGE_QUESTIONS:
		return challengeQuestionsArrayIdentifiers
	default:
		return make(map[string]string)
	}
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

	value := getRawValue(data, key)

	if value == nil {
		return ""
	}
	if reflect.TypeOf(value).Kind() == reflect.Int {
		return strconv.Itoa(value.(int))
	}
	if finalArray, ok := value.([]interface{}); ok {
		strArray := make([]string, len(finalArray))
		for i, v := range finalArray {
			strArray[i] = fmt.Sprintf("%v", v)
		}
		return strings.Join(strArray, ",")
	}
	return fmt.Sprintf("%v", value)
}

func getRawValue(data interface{}, pathString string) interface{} {

	parts := GetPathKeys(pathString)
	for _, part := range parts {
		switch v := data.(type) {
		case map[interface{}]interface{}:
			data = v[part]
		case map[string]interface{}:
			data = v[part]
		case []interface{}:
			index, err := GetArrayIndex(v, part)
			if err != nil {
				return nil
			}
			if len(v) > index {
				data = v[index]
			} else {
				return nil
			}
		default:
			return nil
		}
	}
	return data
}

func ReplaceValue(data interface{}, pathString string, replacement string) interface{} {
	return ReplaceRawValue(data, pathString, replacement)
}

func ReplaceRawValue(data interface{}, pathString string, replacement interface{}) interface{} {

	path := GetPathKeys(pathString)
	if len(path) == 1 {
		switch v := data.(type) {
		case map[interface{}]interface{}:
			v[path[0]] = replacement
		case map[string]interface{}:
			v[path[0]] = replacement
		}
	} else {
		switch v := data.(type) {
		case map[interface{}]interface{}:
			currentKey := path[0]
			v[currentKey] = ReplaceRawValue(v[currentKey], strings.Join(path[1:], "."), replacement)
		case map[string]interface{}:
			currentKey := path[0]
			v[currentKey] = ReplaceRawValue(v[currentKey], strings.Join(path[1:], "."), replacement)
		case []interface{}:
			currentKey := path[0]
			index, err := GetArrayIndex(v, currentKey)
			if err != nil {
				log.Printf("Error: when resolving array index for element %s.", currentKey)
				return data
			}
			if len(v) > index {
				v[index] = ReplaceRawValue(v[index], strings.Join(path[1:], "."), replacement)
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

func extendPath(base, segment string) string {
	if base == "" {
		return segment
	}
	return base + "." + segment
}

var DataPreprocessFuncs = map[ResourceType]func(interface{}) (interface{}, error){}
