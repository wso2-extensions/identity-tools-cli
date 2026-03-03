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
	"fmt"
	"log"
	"strings"
)

type ResourceIdentifierMap map[ResourceType]map[string]string

var resourceIdentifierMap = make(ResourceIdentifierMap)

func AddToResourceIdentifierMap(resourceType ResourceType, resourceData interface{}, operation string) error {

	resourceMeta, exists := RESOURCE_IDENTIFIER_METADATA[resourceType]
	if !exists {
		return nil
	}

	idValue := GetValue(resourceData, resourceMeta.IdentifierPath)
	uniqueValue := GetValue(resourceData, resourceMeta.UniqueValuePath)

	if idValue == "" || uniqueValue == "" {
		log.Println("Warning: Could not extract identifier or unique value. Skipping resource identifier map entry.")
		return nil
	}

	if resourceIdentifierMap[resourceType] == nil {
		resourceIdentifierMap[resourceType] = make(map[string]string)
	}

	switch operation {
	case EXPORT:
		resourceIdentifierMap[resourceType][idValue] = uniqueValue
	case IMPORT:
		resourceIdentifierMap[resourceType][uniqueValue] = idValue
	}

	return nil
}

func ReplaceReferences(resourceType ResourceType, resourceData interface{}) (interface{}, error) {

	references, exists := RESOURCE_REFERENCE_METADATA[resourceType]
	if !exists || len(references) == 0 {
		return resourceData, nil
	}

	for _, refData := range references {
		referencedType := refData.ReferencedResourceType
		identifierMap := resourceIdentifierMap[referencedType]

		for _, refPath := range refData.ReferencePaths {
			err := replaceReferenceValue(resourceData, refPath, identifierMap)
			if err != nil {
				return nil, fmt.Errorf("error replacing references for referenced type %s: %w", referencedType, err)
			}
		}
	}

	return resourceData, nil
}

func replaceReferenceValue(resourceData interface{}, refPath string, identifierMap map[string]string) error {

	concretePaths, err := ResolveAllItemsPaths(resourceData, refPath)
	if err != nil {
		return fmt.Errorf("error resolving paths for '%s': %w", refPath, err)
	}

	for _, path := range concretePaths {
		err := ReplaceValueAtPath(resourceData, path, identifierMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReplaceValueAtPath(resourceData interface{}, path string, identifierMap map[string]string) error {

	rawValue := getRawValue(resourceData, path)
	if rawValue == nil {
		return nil
	}

	switch v := rawValue.(type) {
	case string:
		if v == "" {
			return nil
		}
		newValue, exists := identifierMap[v]
		if !exists {
			return fmt.Errorf("referenced resource with identifier '%s' has not been exported", v)
		}
		ReplaceValue(resourceData, path, newValue)
	case []interface{}:
		newArray, err := ReplaceArrayReferences(v, identifierMap, path)
		if err != nil {
			return err
		}
		ReplaceRawValue(resourceData, path, newArray)
	default:
		return fmt.Errorf("unexpected value type %T at path '%s': expected string or string array", rawValue, path)
	}

	return nil
}

func ReplaceArrayReferences(arrayVal []interface{}, identifierMap map[string]string, refPath string) ([]interface{}, error) {

	newArray := make([]interface{}, len(arrayVal))

	for i, elem := range arrayVal {
		strElem, ok := elem.(string)
		if !ok {
			return nil, fmt.Errorf("expected string array at path '%s' but found an element of type %T", refPath, elem)
		}
		if strElem == "" {
			newArray[i] = strElem
			continue
		}

		newValue, exists := identifierMap[strElem]
		if !exists {
			return nil, fmt.Errorf("referenced resource with identifier '%s' has not been exported", strElem)
		}
		newArray[i] = newValue
	}

	return newArray, nil
}

func ResolveAllItemsPaths(data interface{}, pathString string) ([]string, error) {

	if !containsAllItemsWildcard(pathString) {
		return []string{pathString}, nil
	}

	return expandAllItemsPaths(data, GetPathKeys(pathString), "")
}

func expandAllItemsPaths(data interface{}, parts []string, prefix string) ([]string, error) {

	if len(parts) == 0 {
		return []string{prefix}, nil
	}
	part := parts[0]
	rest := parts[1:]

	if containsAllItemsWildcard(part) {
		// data at this point must be an array
		arr, ok := data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array for value at ALL_ITEMS wildcard path '%s'", prefix)
		}

		keyName := strings.SplitN(part[1:len(part)-1], "=", 2)[0]
		remainingPath := strings.Join(rest, ".")
		hasMoreWildcards := containsAllItemsWildcard(remainingPath)

		var allPaths []string
		for _, elem := range arr {
			keyVal := GetValue(elem, keyName)
			if keyVal == "" {
				return nil, fmt.Errorf("could not find key '%s' in array element at path '%s'", keyName, prefix)
			}
			concretePrefix := extendPath(prefix, fmt.Sprintf("[%s=%s]", keyName, keyVal))

			if hasMoreWildcards {
				subPaths, err := expandAllItemsPaths(elem, rest, concretePrefix)
				if err != nil {
					return nil, err
				}
				allPaths = append(allPaths, subPaths...)
			} else {
				if remainingPath == "" {
					allPaths = append(allPaths, concretePrefix)
				} else {
					allPaths = append(allPaths, concretePrefix+"."+remainingPath)
				}
			}
		}

		return allPaths, nil
	}

	var nextData interface{}
	switch v := data.(type) {
	case map[interface{}]interface{}:
		nextData = v[part]
	case map[string]interface{}:
		nextData = v[part]
	}

	return expandAllItemsPaths(nextData, rest, extendPath(prefix, part))
}

func containsAllItemsWildcard(pathString string) bool {

	return strings.Contains(pathString, "="+ALL_ITEMS+"]")
}

func ResetResourceIdentifierMap() {

	resourceIdentifierMap = make(ResourceIdentifierMap)
}

func GetResourceIdentifierMap(resourceType ResourceType) map[string]string {

	return resourceIdentifierMap[resourceType]
}
