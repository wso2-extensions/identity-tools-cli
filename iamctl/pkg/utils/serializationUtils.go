/**
* Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"fmt"
	"strings"

	"github.com/clbanning/mxj/v2"
	"gopkg.in/yaml.v3"
)

func FormatFromExtension(ext string) (Format, error) {

	switch strings.ToLower(ext) {
	case ".yml", ".yaml":
		return FormatYAML, nil
	case ".json":
		return FormatJSON, nil
	case ".xml":
		return FormatXML, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func (f Format) Extension() string {

	switch f {
	case FormatJSON:
		return ".json"
	case FormatXML:
		return ".xml"
	default:
		return ".yml"
	}
}

func FormatFromString(format string) Format {

	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	case "xml":
		return FormatXML
	default:
		return FormatYAML
	}
}

func Serialize(data interface{}, format Format, resourceType ResourceType) ([]byte, error) {

	switch format {
	case FormatYAML:
		return yaml.Marshal(data)
	case FormatJSON:
		return json.MarshalIndent(data, "", "  ")
	case FormatXML:
		xmlMap, ok := data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("XML serialization requires map[string]interface{}, got %T", data)
		}

		xmlMap = AddXMLRootTag(xmlMap, resourceType)
		mv := mxj.Map(xmlMap)
		mxj.SetAttrPrefix("-")

		xmlData, err := mv.XmlIndent("", "  ")
		if err != nil {
			return nil, err
		}
		return FixXmlStructure(xmlData), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func Deserialize(data []byte, format Format, resourceType ResourceType, target ...interface{}) (interface{}, error) {

	switch format {
	case FormatYAML:
		var result interface{}
		if len(target) > 0 {
			return target[0], yaml.Unmarshal(data, target[0])
		}
		return result, yaml.Unmarshal(data, &result)
	case FormatJSON:
		var result interface{}
		if len(target) > 0 {
			return target[0], json.Unmarshal(data, target[0])
		}
		return result, json.Unmarshal(data, &result)
	case FormatXML:
		xmlMap, err := XMLToMap(data)
		if err != nil {
			return nil, fmt.Errorf("error when parsing data to XML: %w", err)
		}

		xmlData, err := RemoveXMLRootTag(xmlMap, resourceType)
		if err != nil {
			return nil, err
		}

		result := FixArrayFields(xmlData, resourceType)
		if len(target) > 0 {
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("error when converting XML data to typed struct: %w", err)
			}
			return target[0], json.Unmarshal(jsonBytes, target[0])
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func FixXmlStructure(data []byte) []byte {

	xmlStr := string(data)

	xmlStr = strings.ReplaceAll(xmlStr, " xsi=", " xmlns:xsi=")
	if !strings.Contains(xmlStr, "xsi:type=") {
		xmlStr = strings.ReplaceAll(xmlStr, " type=", " xsi:type=")
	}
	if !strings.Contains(xmlStr, "xsi:nil=") {
		xmlStr = strings.ReplaceAll(xmlStr, " nil=", " xsi:nil=")
	}

	return []byte(xmlStr)
}

func XMLToMap(data []byte) (map[string]interface{}, error) {

	mxj.SetAttrPrefix("-")
	mxj.XMLEscapeChars(true)
	m, err := mxj.NewMapXml(data)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func GetXMLRootTag(resourceType ResourceType) string {

	xmlRootTags := map[ResourceType]string{
		OIDC_SCOPES:           XML_ROOT_OIDC_SCOPE,
		ROLES:                 XML_ROOT_ROLE,
		CHALLENGE_QUESTIONS:   XML_ROOT_CHALLENGE_QUESTION,
		EMAIL_TEMPLATES:       XML_ROOT_EMAIL_TEMPLATE,
		SCRIPT_LIBRARIES:      XML_ROOT_SCRIPT_LIBRARY,
		GOVERNANCE_CONNECTORS: XML_ROOT_GOVERNANCE_CONNECTOR,
		USERSTORES:            XML_ROOT_USERSTORE,
		CLAIMS:                XML_ROOT_CLAIM,
		IDENTITY_PROVIDERS:    XML_ROOT_IDENTITY_PROVIDER,
		APPLICATIONS:          XML_ROOT_APPLICATION,
	}
	return xmlRootTags[resourceType]
}

func AddXMLRootTag(data map[string]interface{}, resourceType ResourceType) map[string]interface{} {

	xmlRootTag := GetXMLRootTag(resourceType)
	if xmlRootTag != "" {
		return map[string]interface{}{xmlRootTag: data}
	}
	return data
}

func RemoveXMLRootTag(xmlMap map[string]interface{}, resourceType ResourceType) (interface{}, error) {

	xmlRootTag := GetXMLRootTag(resourceType)
	if xmlRootTag == "" {
		return xmlMap, nil
	}

	if rootValue, ok := xmlMap[xmlRootTag]; ok {
		return rootValue, nil
	}

	return nil, fmt.Errorf("expected root element <%s> not found in XML", xmlRootTag)
}

func GetArrayFieldPaths(resourceType ResourceType) []string {

	switch resourceType {
	case OIDC_SCOPES:
		return oidcScopeArrayFields
	case ROLES:
		return rolesArrayFields
	case CHALLENGE_QUESTIONS:
		return challengeQuestionsArrayFields
	case GOVERNANCE_CONNECTORS:
		return governanceConnectorArrayFields
	case USERSTORES:
		return userStoreArrayFields
	case CLAIMS:
		return claimArrayFields
	case IDENTITY_PROVIDERS:
		return idpArrayFields
	case APPLICATIONS:
		return appArrayFields
	default:
		return []string{}
	}
}

func FixArrayFields(data interface{}, resourceType ResourceType) interface{} {

	arrayPaths := GetArrayFieldPaths(resourceType)

	for _, path := range arrayPaths {
		value := getRawValue(data, path)

		if value != nil {
			if _, isArray := value.([]interface{}); !isArray {
				if strValue, isString := value.(string); isString && strValue == "" {
					data = ReplaceRawValue(data, path, []interface{}{})
				} else {
					data = ReplaceRawValue(data, path, []interface{}{value})
				}
			}
		}
	}
	return data
}

func DeserializeToMap(data []byte, format Format, resourceType ResourceType, excludeFields ...string) (map[string]interface{}, error) {

	parsed, err := Deserialize(data, format, resourceType)
	if err != nil {
		return nil, fmt.Errorf("error deserializing data: %w", err)
	}

	if interfaceMap, ok := parsed.(map[interface{}]interface{}); ok {
		parsed = ConvertToStringKeyMap(interfaceMap)
	}

	dataMap, ok := parsed.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not in expected map format")
	}

	for _, field := range excludeFields {
		delete(dataMap, field)
	}

	return dataMap, nil
}
