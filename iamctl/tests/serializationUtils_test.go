package tests

import (
	"reflect"
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func TestFixArrayFields(t *testing.T) {

	testCases := []struct {
		description    string
		data           interface{}
		resourceType   utils.ResourceType
		expectedResult interface{}
	}{
		{
			description:  "Single string converted to array",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": []interface{}{"roles"},
				"name":   "test",
			},
		},
		{
			description:  "Empty string converted to empty array",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": "",
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": []interface{}{},
				"name":   "test",
			},
		},
		{
			description:  "Array with multiple elements unchanged",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": []interface{}{"roles", "email"},
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": []interface{}{"roles", "email"},
				"name":   "test",
			},
		},
		{
			description:  "Empty array unchanged",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": []interface{}{},
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": []interface{}{},
				"name":   "test",
			},
		},
		{
			description:  "Missing field remains missing",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"name": "test",
			},
			expectedResult: map[string]interface{}{
				"name": "test",
			},
		},
		{
			description:  "Resource types with no array field paths defined unchanged",
			resourceType: utils.APPLICATIONS,
			data: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
		},
		{
			description:  "Numeric value converted to array",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": 123,
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": []interface{}{123},
				"name":   "test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := utils.FixArrayFields(tc.data, tc.resourceType)
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("Unexpected result for %s: expected %+v, but got %+v", tc.description, tc.expectedResult, result)
			}
		})
	}
}

func TestRemoveXMLRootTag(t *testing.T) {

	testCases := []struct {
		description    string
		xmlMap         map[string]interface{}
		resourceType   utils.ResourceType
		expectedResult interface{}
		expectError    bool
	}{
		{
			description:  "Root tag present - returns value under root",
			resourceType: utils.OIDC_SCOPES,
			xmlMap: map[string]interface{}{
				"Scope": map[string]interface{}{
					"claims": "roles",
					"name":   "test",
				},
			},
			expectedResult: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
			expectError: false,
		},
		{
			description:  "Root tag missing - returns error",
			resourceType: utils.OIDC_SCOPES,
			xmlMap: map[string]interface{}{
				"WrongTag": map[string]interface{}{
					"claims": "roles",
					"name":   "test",
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			description:  "No root tag defined - returns map as-is",
			resourceType: utils.ResourceType("unknown"),
			xmlMap: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
			expectError: false,
		},
		{
			description:  "Root tag with nested structure",
			resourceType: utils.OIDC_SCOPES,
			xmlMap: map[string]interface{}{
				"Scope": map[string]interface{}{
					"claims": []interface{}{"roles", "email"},
					"nested": map[string]interface{}{
						"key": "value",
					},
				},
			},
			expectedResult: map[string]interface{}{
				"claims": []interface{}{"roles", "email"},
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := utils.RemoveXMLRootTag(tc.xmlMap, tc.resourceType)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.description, err)
				}
				if !reflect.DeepEqual(result, tc.expectedResult) {
					t.Errorf("Unexpected result for %s: expected %+v, but got %+v", tc.description, tc.expectedResult, result)
				}
			}
		})
	}
}

func TestAddXMLRootTag(t *testing.T) {

	testCases := []struct {
		description    string
		data           map[string]interface{}
		resourceType   utils.ResourceType
		expectedResult map[string]interface{}
	}{
		{
			description:  "Root tag defined - wraps data",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"Scope": map[string]interface{}{
					"claims": "roles",
					"name":   "test",
				},
			},
		},
		{
			description:  "No root tag defined - returns data as-is",
			resourceType: utils.ResourceType("unknown"),
			data: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"claims": "roles",
				"name":   "test",
			},
		},
		{
			description:  "Root tag with array data",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": []interface{}{"roles", "email"},
				"name":   "test",
			},
			expectedResult: map[string]interface{}{
				"Scope": map[string]interface{}{
					"claims": []interface{}{"roles", "email"},
					"name":   "test",
				},
			},
		},
		{
			description:  "Root tag with nested structure",
			resourceType: utils.OIDC_SCOPES,
			data: map[string]interface{}{
				"claims": "roles",
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			expectedResult: map[string]interface{}{
				"Scope": map[string]interface{}{
					"claims": "roles",
					"nested": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
		{
			description:  "Empty data with root tag",
			resourceType: utils.OIDC_SCOPES,
			data:         map[string]interface{}{},
			expectedResult: map[string]interface{}{
				"Scope": map[string]interface{}{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := utils.AddXMLRootTag(tc.data, tc.resourceType)
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("Unexpected result for %s: expected %+v, but got %+v", tc.description, tc.expectedResult, result)
			}
		})
	}
}
