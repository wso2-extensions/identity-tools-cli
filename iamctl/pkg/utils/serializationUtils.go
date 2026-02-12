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
	"regexp"
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

func Serialize(data interface{}, format Format, resourceType string) ([]byte, error) {

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
		if xmlRootTag := GetXMLRootTag(resourceType); xmlRootTag != "" {
			xmlMap = map[string]interface{}{xmlRootTag: xmlMap}
		}

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

func Deserialize(data []byte, format Format, resourceType string) (interface{}, error) {

	switch format {
	case FormatYAML:
		var result interface{}
		err := yaml.Unmarshal(data, &result)
		if err != nil {
			return nil, fmt.Errorf("error when parsing data to YAML: %w", err)
		}
		return result, nil
	case FormatJSON:
		var result interface{}
		err := json.Unmarshal(data, &result)
		if err != nil {
			return nil, fmt.Errorf("error when parsing data to JSON: %w", err)
		}
		return result, nil
	case FormatXML:
		xmlMap, err := XMLToMap(data)
		if err != nil {
			return nil, fmt.Errorf("error when parsing data to XML: %w", err)
		}
		if xmlRootTag := GetXMLRootTag(resourceType); xmlRootTag != "" {
			if rootValue, ok := xmlMap[xmlRootTag]; ok {
				return rootValue, nil
			}
			return nil, fmt.Errorf("expected root element <%s> not found in XML", xmlRootTag)
		}

		return xmlMap, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
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

func XMLToMap(data []byte) (map[string]interface{}, error) {

	mxj.SetAttrPrefix("-")
	m, err := mxj.NewMapXml(data)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func GetXMLRootTag(resourceType string) string {

	xmlRootTags := map[string]string{}
	return xmlRootTags[resourceType]
}
