package tests

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func TestAddToResourceIdentifierMap(t *testing.T) {

	testCases := []struct {
		description  string
		resourceType utils.ResourceType
		metadata     map[utils.ResourceType]utils.ResourceIdentifierMeta
		resourceData interface{}
		operation    string
		expectedMap  map[string]string
	}{
		{
			description:  "Export: maps id to name",
			resourceType: utils.APPLICATIONS,
			metadata: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			resourceData: map[string]interface{}{
				"id":   "app-uuid-123",
				"name": "MyApp",
			},
			operation:   utils.EXPORT,
			expectedMap: map[string]string{"app-uuid-123": "MyApp"},
		},
		{
			description:  "Import: maps name to id",
			resourceType: utils.APPLICATIONS,
			metadata: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			resourceData: map[string]interface{}{
				"id":   "app-uuid-456",
				"name": "MyApp",
			},
			operation:   utils.IMPORT,
			expectedMap: map[string]string{"MyApp": "app-uuid-456"},
		},
		{
			description:  "Resource type not in metadata: no entry added",
			resourceType: utils.IDENTITY_PROVIDERS,
			metadata: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			resourceData: map[string]interface{}{
				"id":   "idp-uuid-123",
				"name": "MyIDP",
			},
			operation:   utils.EXPORT,
			expectedMap: nil,
		},
		{
			description:  "Missing id value: no entry added",
			resourceType: utils.APPLICATIONS,
			metadata: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			resourceData: map[string]interface{}{
				"name": "MyApp",
			},
			operation:   utils.EXPORT,
			expectedMap: nil,
		},
		{
			description:  "Missing name value: no entry added",
			resourceType: utils.APPLICATIONS,
			metadata: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			resourceData: map[string]interface{}{
				"id": "app-uuid-123",
			},
			operation:   utils.EXPORT,
			expectedMap: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			utils.RESOURCE_IDENTIFIER_METADATA = tc.metadata
			utils.ResetResourceIdentifierMap()

			utils.ExtractAndRegisterIdentifier(tc.resourceType, tc.resourceData, tc.operation)

			result := utils.GetResourceIdentifierMap(tc.resourceType)
			if !reflect.DeepEqual(result, tc.expectedMap) {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", tc.description, tc.expectedMap, result)
			}
		})
	}
}

func TestResolveAllItemsPaths(t *testing.T) {

	data := map[string]interface{}{
		"applicationId": "app-id-1",
		"properties": []interface{}{
			map[string]interface{}{"name": "prop1", "value": "val1"},
			map[string]interface{}{"name": "prop2", "value": "val2"},
		},
		"configs": []interface{}{
			map[string]interface{}{
				"type": "typeA",
				"items": []interface{}{
					map[string]interface{}{"key": "k1", "appId": "id1"},
					map[string]interface{}{"key": "k2", "appId": "id2"},
				},
			},
			map[string]interface{}{
				"type": "typeB",
				"items": []interface{}{
					map[string]interface{}{"key": "k3", "appId": "id3"},
				},
			},
		},
	}

	allItems := utils.ALL_ITEMS

	testCases := []struct {
		description   string
		path          string
		expectedPaths []string
		expectError   bool
	}{
		{
			description:   "Path with no wildcard: returned unchanged",
			path:          "applicationId",
			expectedPaths: []string{"applicationId"},
		},
		{
			description:   "Nested path with no wildcard: returned unchanged",
			path:          "properties.[name=prop1].value",
			expectedPaths: []string{"properties.[name=prop1].value"},
		},
		{
			description: "Wildcard on array with remaining path: expands to one path per element",
			path:        fmt.Sprintf("properties.[name=%s].value", allItems),
			expectedPaths: []string{
				"properties.[name=prop1].value",
				"properties.[name=prop2].value",
			},
		},
		{
			description: "Wildcard with no remaining path: expands to array element identifiers",
			path:        fmt.Sprintf("properties.[name=%s]", allItems),
			expectedPaths: []string{
				"properties.[name=prop1]",
				"properties.[name=prop2]",
			},
		},
		{
			description: "Nested wildcards: full cartesian expansion",
			path:        fmt.Sprintf("configs.[type=%s].items.[key=%s].appId", allItems, allItems),
			expectedPaths: []string{
				"configs.[type=typeA].items.[key=k1].appId",
				"configs.[type=typeA].items.[key=k2].appId",
				"configs.[type=typeB].items.[key=k3].appId",
			},
		},
		{
			description: "Wildcard on non-array field: error",
			path:        fmt.Sprintf("applicationId.[name=%s].value", allItems),
			expectError: true,
		},
		{
			description: "Array element missing wildcard key: error",
			path:        fmt.Sprintf("properties.[missingKey=%s].value", allItems),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := utils.ResolveAllItemsPaths(data, tc.path)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tc.description)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.description, err)
				return
			}

			sort.Strings(result)
			sort.Strings(tc.expectedPaths)
			if !reflect.DeepEqual(result, tc.expectedPaths) {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", tc.description, tc.expectedPaths, result)
			}
		})
	}
}

func TestReplaceArrayReferences(t *testing.T) {

	testCases := []struct {
		description   string
		arrayVal      []interface{}
		identifierMap map[string]string
		refPath       string
		expectedArray []interface{}
		expectError   bool
	}{
		{
			description:   "All elements replaced successfully",
			arrayVal:      []interface{}{"id-1", "id-2", "id-3"},
			identifierMap: map[string]string{"id-1": "App1", "id-2": "App2", "id-3": "App3"},
			refPath:       "applicationIds",
			expectedArray: []interface{}{"App1", "App2", "App3"},
		},
		{
			description:   "Empty string elements are kept as-is",
			arrayVal:      []interface{}{"id-1", "", "id-3"},
			identifierMap: map[string]string{"id-1": "App1", "id-3": "App3"},
			refPath:       "applicationIds",
			expectedArray: []interface{}{"App1", "", "App3"},
		},
		{
			description:   "Empty array: returns empty array",
			arrayVal:      []interface{}{},
			identifierMap: map[string]string{"id-1": "App1"},
			refPath:       "applicationIds",
			expectedArray: []interface{}{},
		},
		{
			description:   "Element not found in identifier map: error",
			arrayVal:      []interface{}{"id-1", "id-unknown"},
			identifierMap: map[string]string{"id-1": "App1"},
			refPath:       "applicationIds",
			expectError:   true,
		},
		{
			description:   "Non-string int element in array: error",
			arrayVal:      []interface{}{"id-1", 42},
			identifierMap: map[string]string{"id-1": "App1"},
			refPath:       "applicationIds",
			expectError:   true,
		},
		{
			description:   "Map element in array: error",
			arrayVal:      []interface{}{"id-1", map[string]interface{}{"key": "val"}},
			identifierMap: map[string]string{"id-1": "App1"},
			refPath:       "applicationIds",
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := utils.ReplaceArrayReferences(tc.arrayVal, tc.identifierMap, tc.refPath)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tc.description)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.description, err)
				return
			}
			if !reflect.DeepEqual(result, tc.expectedArray) {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", tc.description, tc.expectedArray, result)
			}
		})
	}
}

func TestReplaceValueAtPath(t *testing.T) {

	testCases := []struct {
		description   string
		resourceData  interface{}
		path          string
		identifierMap map[string]string
		expectedData  interface{}
		expectError   bool
	}{
		{
			description: "Single string value: replaced with mapped value",
			resourceData: map[string]interface{}{
				"applicationId": "app-uuid-1",
			},
			path:          "applicationId",
			identifierMap: map[string]string{"app-uuid-1": "MyApp"},
			expectedData: map[string]interface{}{
				"applicationId": "MyApp",
			},
		},
		{
			description: "String array: each element replaced individually",
			resourceData: map[string]interface{}{
				"applicationIds": []interface{}{"app-uuid-1", "app-uuid-2"},
			},
			path:          "applicationIds",
			identifierMap: map[string]string{"app-uuid-1": "App1", "app-uuid-2": "App2"},
			expectedData: map[string]interface{}{
				"applicationIds": []interface{}{"App1", "App2"},
			},
		},
		{
			description: "Nil value at path: no-op, no error",
			resourceData: map[string]interface{}{
				"otherField": "value",
			},
			path:          "applicationId",
			identifierMap: map[string]string{"app-uuid-1": "MyApp"},
			expectedData: map[string]interface{}{
				"otherField": "value",
			},
		},
		{
			description: "Empty string at path: no-op, no error",
			resourceData: map[string]interface{}{
				"applicationId": "",
			},
			path:          "applicationId",
			identifierMap: map[string]string{"app-uuid-1": "MyApp"},
			expectedData: map[string]interface{}{
				"applicationId": "",
			},
		},
		{
			description: "String value not in identifier map: error",
			resourceData: map[string]interface{}{
				"applicationId": "app-uuid-unknown",
			},
			path:          "applicationId",
			identifierMap: map[string]string{"app-uuid-1": "MyApp"},
			expectError:   true,
		},
		{
			description: "Non-string non-array value at path: error",
			resourceData: map[string]interface{}{
				"applicationId": 12345,
			},
			path:          "applicationId",
			identifierMap: map[string]string{},
			expectError:   true,
		},
		{
			description: "Nested path replaced correctly",
			resourceData: map[string]interface{}{
				"config": map[string]interface{}{
					"appId": "app-uuid-1",
				},
			},
			path:          "config.appId",
			identifierMap: map[string]string{"app-uuid-1": "MyApp"},
			expectedData: map[string]interface{}{
				"config": map[string]interface{}{
					"appId": "MyApp",
				},
			},
		},
		{
			description: "Non-string element in string array at path: error",
			resourceData: map[string]interface{}{
				"applicationIds": []interface{}{"app-uuid-1", 42},
			},
			path:          "applicationIds",
			identifierMap: map[string]string{"app-uuid-1": "App1"},
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := utils.ReplaceValueAtPath(tc.resourceData, tc.path, tc.identifierMap)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tc.description)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.description, err)
				return
			}
			if !reflect.DeepEqual(tc.resourceData, tc.expectedData) {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", tc.description, tc.expectedData, tc.resourceData)
			}
		})
	}
}

func TestReplaceReferences(t *testing.T) {

	allItems := utils.ALL_ITEMS

	testCases := []struct {
		description    string
		resourceType   utils.ResourceType
		referencedType utils.ResourceType
		refMetadata    map[utils.ResourceType][]utils.ResourceReferenceMeta
		identifierMeta map[utils.ResourceType]utils.ResourceIdentifierMeta
		identifierData []interface{}
		operation      string
		resourceData   interface{}
		expectedData   interface{}
		expectError    bool
	}{
		{
			description:  "Resource type with no reference metadata: returns data unchanged",
			resourceType: utils.IDENTITY_PROVIDERS,
			refMetadata:  map[utils.ResourceType][]utils.ResourceReferenceMeta{},
			resourceData: map[string]interface{}{
				"applicationId": "app-uuid-1",
			},
			expectedData: map[string]interface{}{
				"applicationId": "app-uuid-1",
			},
		},
		{
			description:    "Single string reference replaced correctly",
			resourceType:   utils.IDENTITY_PROVIDERS,
			referencedType: utils.APPLICATIONS,
			refMetadata: map[utils.ResourceType][]utils.ResourceReferenceMeta{
				utils.IDENTITY_PROVIDERS: {
					{ReferencedResourceType: utils.APPLICATIONS, ReferencePaths: []string{"applicationId"}},
				},
			},
			identifierMeta: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			identifierData: []interface{}{
				map[string]interface{}{"id": "app-uuid-1", "name": "MyApp"},
			},
			operation: utils.EXPORT,
			resourceData: map[string]interface{}{
				"applicationId": "app-uuid-1",
			},
			expectedData: map[string]interface{}{
				"applicationId": "MyApp",
			},
		},
		{
			description:    "String array reference: each element replaced",
			resourceType:   utils.IDENTITY_PROVIDERS,
			referencedType: utils.APPLICATIONS,
			refMetadata: map[utils.ResourceType][]utils.ResourceReferenceMeta{
				utils.IDENTITY_PROVIDERS: {
					{ReferencedResourceType: utils.APPLICATIONS, ReferencePaths: []string{"applicationIds"}},
				},
			},
			identifierMeta: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			identifierData: []interface{}{
				map[string]interface{}{"id": "app-uuid-1", "name": "App1"},
				map[string]interface{}{"id": "app-uuid-2", "name": "App2"},
			},
			operation: utils.EXPORT,
			resourceData: map[string]interface{}{
				"applicationIds": []interface{}{"app-uuid-1", "app-uuid-2"},
			},
			expectedData: map[string]interface{}{
				"applicationIds": []interface{}{"App1", "App2"},
			},
		},
		{
			description:    "ALL_ITEMS wildcard path: all array elements processed",
			resourceType:   utils.IDENTITY_PROVIDERS,
			referencedType: utils.APPLICATIONS,
			refMetadata: map[utils.ResourceType][]utils.ResourceReferenceMeta{
				utils.IDENTITY_PROVIDERS: {
					{
						ReferencedResourceType: utils.APPLICATIONS,
						ReferencePaths:         []string{fmt.Sprintf("properties.[name=%s].appId", allItems)},
					},
				},
			},
			identifierMeta: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			identifierData: []interface{}{
				map[string]interface{}{"id": "app-uuid-1", "name": "App1"},
				map[string]interface{}{"id": "app-uuid-2", "name": "App2"},
			},
			operation: utils.EXPORT,
			resourceData: map[string]interface{}{
				"properties": []interface{}{
					map[string]interface{}{"name": "prop1", "appId": "app-uuid-1"},
					map[string]interface{}{"name": "prop2", "appId": "app-uuid-2"},
				},
			},
			expectedData: map[string]interface{}{
				"properties": []interface{}{
					map[string]interface{}{"name": "prop1", "appId": "App1"},
					map[string]interface{}{"name": "prop2", "appId": "App2"},
				},
			},
		},
		{
			description:    "Referenced identifier not in map: error",
			resourceType:   utils.IDENTITY_PROVIDERS,
			referencedType: utils.APPLICATIONS,
			refMetadata: map[utils.ResourceType][]utils.ResourceReferenceMeta{
				utils.IDENTITY_PROVIDERS: {
					{ReferencedResourceType: utils.APPLICATIONS, ReferencePaths: []string{"applicationId"}},
				},
			},
			identifierMeta: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			identifierData: []interface{}{
				map[string]interface{}{"id": "app-uuid-1", "name": "App1"},
			},
			operation: utils.EXPORT,
			resourceData: map[string]interface{}{
				"applicationId": "app-uuid-unknown",
			},
			expectError: true,
		},
		{
			description:    "Non-string element in referenced array: error",
			resourceType:   utils.IDENTITY_PROVIDERS,
			referencedType: utils.APPLICATIONS,
			refMetadata: map[utils.ResourceType][]utils.ResourceReferenceMeta{
				utils.IDENTITY_PROVIDERS: {
					{ReferencedResourceType: utils.APPLICATIONS, ReferencePaths: []string{"applicationIds"}},
				},
			},
			identifierMeta: map[utils.ResourceType]utils.ResourceIdentifierMeta{
				utils.APPLICATIONS: {IdentifierPath: "id", UniqueValuePath: "name"},
			},
			identifierData: []interface{}{
				map[string]interface{}{"id": "app-uuid-1", "name": "App1"},
			},
			operation: utils.EXPORT,
			resourceData: map[string]interface{}{
				"applicationIds": []interface{}{"app-uuid-1", 42},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			utils.RESOURCE_REFERENCE_METADATA = tc.refMetadata
			utils.RESOURCE_IDENTIFIER_METADATA = tc.identifierMeta
			utils.ResetResourceIdentifierMap()

			for _, data := range tc.identifierData {
				utils.ExtractAndRegisterIdentifier(tc.referencedType, data, tc.operation)
			}

			result, err := utils.ReplaceReferences(tc.resourceType, tc.resourceData)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tc.description)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.description, err)
				return
			}
			if !reflect.DeepEqual(result, tc.expectedData) {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", tc.description, tc.expectedData, result)
			}
		})
	}
}
