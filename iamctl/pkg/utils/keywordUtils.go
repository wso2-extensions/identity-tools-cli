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
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/clbanning/mxj/v2"
	"gopkg.in/yaml.v2"
)

func ReplaceKeywords(fileContent string, keywordMapping map[string]any) string {

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

func ProcessExportedContent(exportedFileName string, exportedFileContent []byte, keywordMapping map[string]any, resourceType string) ([]byte, error) {
	fileType := strings.ToLower(filepath.Ext(exportedFileName))
	// Only YAML uses !!org.wso2 type tags; avoid mangling XML/JSON
	if fileType == ".yaml" || fileType == ".yml" {
		// To preserve type tags in the exported YAML file, replace the type tags with a placeholder.
		exportedFileContent = ReplaceTypeTags(exportedFileContent)
	}
	var exportedData any

	switch fileType {
	case ".json":
		var exportedJson any
		err := json.Unmarshal(exportedFileContent, &exportedJson)
		if err != nil {
			return nil, fmt.Errorf("error when parsing exported data to JSON. %w", err)
		}
		exportedData = exportedJson
	case ".xml":
		// FIX: Parse XML to Map instead of treating as string
		exportedXmlMap, err := XMLToMap(exportedFileContent)
		if err != nil {
			return nil, fmt.Errorf("error when parsing exported data to XML. %w", err)
		}
		exportedData = exportedXmlMap
	default:
		var exportedYaml any
		err := yaml.Unmarshal(exportedFileContent, &exportedYaml)
		if err != nil {
			return nil, fmt.Errorf("error when parsing exported data to YAML. %w", err)
		}
		exportedData = exportedYaml
	}

	// Replace ESVs in the exported file according to the keyword placeholders added in the local file.
	var modifiedExportedData any
	localFileData, err := os.ReadFile(exportedFileName)
	if err != nil {
		log.Printf("Local file not found at %s. Creating new file.", exportedFileName)
		modifiedExportedData = exportedData
	} else {
		modifiedExportedData, err = AddKeywords(exportedData, fileType, localFileData, keywordMapping, resourceType)
		if err != nil {
			log.Println("Error when adding keywords to the exported file. Overriding local file with exported content. ", err)
			modifiedExportedData = exportedData
		}
	}

	var modifiedExportedContent []byte
	var marshalErr error

	switch fileType {
	case ".json":
		modifiedExportedContent, marshalErr = json.MarshalIndent(modifiedExportedData, "", "  ")
	case ".xml":
		xmlMap, ok := modifiedExportedData.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unexpected XML data type %T; expected map[string]any", modifiedExportedData)
		}
		mv := mxj.Map(xmlMap)
		mxj.SetAttrPrefix("-")

		xmlData, err := mv.XmlIndent("", "  ")
		if err != nil {
			marshalErr = err
		} else {
			// Post-process the XML to fix the xsi:type and namespace issues
			modifiedExportedContent = FixXmlStructure(xmlData)
		}
	default:
		modifiedExportedContent, marshalErr = yaml.Marshal(modifiedExportedData)
	}

	if marshalErr != nil {
		return nil, fmt.Errorf("error when creating exported data with keywords. %w", marshalErr)
	}

	// Re-add YAML type tags only for YAML files
	if fileType == ".yaml" || fileType == ".yml" {
		modifiedExportedContent = AddTypeTags(modifiedExportedContent)
	}
	return modifiedExportedContent, nil
}

func AddKeywords(exportedData any, fileType string, localFileData []byte, keywordMapping map[string]any, resourceType string) (any, error) {
	var localData any
	switch fileType {
	case ".json":
		var localJson any
		err := json.Unmarshal(localFileData, &localJson)
		if err != nil || localJson == nil {
			err1 := fmt.Errorf("empty or invalid local file data. %w", err)
			return exportedData, err1
		}
		localData = localJson
	case ".xml":
		localJsonMap, err := XMLToMap(localFileData)
		if err != nil || localJsonMap == nil {
			return exportedData, fmt.Errorf("invalid local file data: %w", err)
		}
		localData = localJsonMap
	default:
		var localYaml any
		err := yaml.Unmarshal(localFileData, &localYaml)
		if err != nil || localYaml == nil {
			err1 := fmt.Errorf("empty or invalid local file data. %w", err)
			return exportedData, err1
		}
		localData = localYaml
	}

	// Get keyword locations in local file.
	keywordLocations := GetKeywordLocations(localData, []string{}, keywordMapping, resourceType)
	// Compare the fields with keywords in the exported file and the local file and modify the exported file.
	exportedData = ModifyFieldsWithKeywords(exportedData, localData, keywordLocations, keywordMapping)

	return exportedData, nil
}

func GetKeywordLocations(fileData any, path []string, keywordMapping map[string]any, resourceType string) []string {
	var keys []string
	switch v := fileData.(type) {
	case map[any]any:
		for k, val := range v {
			newPath := append(path, fmt.Sprintf("%v", k))
			keys = append(keys, GetKeywordLocations(val, newPath, keywordMapping, resourceType)...)
		}
	case map[string]any:
		for k, val := range v {
			newPath := append(path, fmt.Sprintf("%v", k))
			keys = append(keys, GetKeywordLocations(val, newPath, keywordMapping, resourceType)...)
		}
	case []any:
		for _, val := range v {
			if _, ok := val.(string); ok {
				if ContainsKeywords(val.(string), keywordMapping) {
					thisPath := strings.Join(path, ".")
					keys = append(keys, thisPath)
				}
				break
			} else {
				parentName := ""
				if len(path) > 0 {
					parentName = path[len(path)-1]
				} else {
					parentName = resourceType
				}

				arrayIdentifiers := GetArrayIdentifiers(resourceType)
				arrayElementPath, err := resolvePathWithIdentifiers(parentName, val, arrayIdentifiers)
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
	case APPLICATIONS:
		return applicationArrayIdentifiers
	case IDENTITY_PROVIDERS:
		return idpArrayIdentifiers
	case USERSTORES:
		return userStoreArrayIdentifiers
	case CLAIMS:
		return claimArrayIdentifiers
	}
	return make(map[string]string)
}

func resolvePathWithIdentifiers(arrayName string, element any, identifiers map[string]string) (string, error) {

	var elementMap any
	elementMap, ok := element.(map[any]any)
	if !ok {
		elementMap, ok = element.(map[string]any)
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

func ContainsKeywords(data string, keywordMapping map[string]any) bool {

	for keyword := range keywordMapping {
		if strings.Contains(data, "{{"+keyword+"}}") {
			return true
		}
	}
	return false
}

func ModifyFieldsWithKeywords(exportedFileData any, localFileData any,
	keywordLocations []string, keywordMap map[string]any) any {

	if exportedStr, ok := exportedFileData.(string); ok {
		if localStr, ok := localFileData.(string); ok && ContainsKeywords(localStr, keywordMap) {
			return ReplaceKeywords(localStr, keywordMap)
		}
		return ReplaceKeywords(exportedStr, keywordMap)
	}

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

func GetValue(data any, key string) string {

	parts := GetPathKeys(key)
	for _, part := range parts {
		switch v := data.(type) {
		case map[any]any:
			data = v[part]
		case map[string]any:
			data = v[part]
		case []any:
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
	if finalArray, ok := data.([]any); ok {
		strArray := make([]string, len(finalArray))
		for i, v := range finalArray {
			strArray[i] = v.(string)
		}
		data = strings.Join(strArray, ",")
	}
	// Convert the final value to a string and return.
	dataAsString := fmt.Sprintf("%v", data)
	return dataAsString
}

func ReplaceValue(data any, pathString string, replacement string) any {

	path := GetPathKeys(pathString)
	if len(path) == 1 {
		switch data := data.(type) {
		case map[any]any:
			data[path[0]] = replacement
		case map[string]any:
			data[path[0]] = replacement
		}
	} else {
		switch v := data.(type) {
		case map[any]any:
			currentKey := path[0]
			data.(map[any]any)[currentKey] = ReplaceValue(v[currentKey], strings.Join(path[1:], "."), replacement)
		case map[string]any:
			currentKey := path[0]
			data.(map[string]any)[currentKey] = ReplaceValue(v[currentKey], strings.Join(path[1:], "."), replacement)
		case []any:
			currentKey := path[0]
			index, err := GetArrayIndex(v, currentKey)
			if err != nil {
				log.Printf("Error: when resolving array index for element %s.", currentKey)
				return data
			}
			if len(v) > index {
				data.([]any)[index] = ReplaceValue(v[index], strings.Join(path[1:], "."), replacement)
			}
		default:
			return data
		}
	}
	return data
}

func GetArrayIndex(arrayMap []any, elementIdentifier string) (int, error) {

	if strings.HasPrefix(elementIdentifier, "[") && strings.HasSuffix(elementIdentifier, "]") {
		identifier := elementIdentifier[1 : len(elementIdentifier)-1]
		parts := strings.SplitN(identifier, "=", 2)
		for k, v := range arrayMap {
			switch v := v.(type) {
			case map[any]any:
				if GetValue(v, parts[0]) == parts[1] {
					return k, nil
				}
			case map[string]any:
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

func XMLToMap(data []byte) (map[string]any, error) {
	mxj.SetAttrPrefix("-")
	m, err := mxj.NewMapXml(data)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func FixXmlStructure(data []byte) []byte {
	xmlStr := string(data)

	xmlStr = strings.ReplaceAll(xmlStr, " xsi=", " xmlns:xsi=")
	xmlStr = strings.ReplaceAll(xmlStr, " type=\"oAuthAppDO\"", " xsi:type=\"oAuthAppDO\"")
	xmlStr = strings.ReplaceAll(xmlStr, " nil=\"true\"", " xsi:nil=\"true\"")

	re := regexp.MustCompile(`(?s)<value>(.*?)</value>`)

	fixedXml := re.ReplaceAllStringFunc(xmlStr, func(match string) string {
		content := strings.TrimPrefix(match, "<value>")
		content = strings.TrimSuffix(content, "</value>")

		if strings.ContainsAny(content, "<>&") || strings.Contains(content, "&lt;") || strings.Contains(content, "&amp;") {
			content = strings.ReplaceAll(content, "&lt;", "<")
			content = strings.ReplaceAll(content, "&gt;", ">")
			content = strings.ReplaceAll(content, "&amp;", "&")
			content = strings.ReplaceAll(content, "&quot;", "\"")
			content = strings.ReplaceAll(content, "&apos;", "'")

			return "<value><![CDATA[" + content + "]]></value>"
		}
		return match
	})

	return []byte(fixedXml)
}
