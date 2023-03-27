package tests

import (
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func TestIsResourceExcluded(t *testing.T) {
	testCases := []struct {
		name            string
		resourceName    string
		resourceConfigs map[string]interface{}
		expectedResult  bool
	}{
		{
			name:         "IncludeOnlyConfig: Resource not excluded",
			resourceName: "resource1",
			resourceConfigs: map[string]interface{}{
				"INCLUDE_ONLY": []interface{}{
					"resource1",
					"resource2",
				},
			},
			expectedResult: false,
		},
		{
			name:         "IncludeOnlyConfig: Resource excluded",
			resourceName: "resource1",
			resourceConfigs: map[string]interface{}{
				"INCLUDE_ONLY": []interface{}{
					"resource2",
					"resource3",
				},
			},
			expectedResult: true,
		},
		{
			name:         "ExcludeConfig: Resource excluded",
			resourceName: "resource1",
			resourceConfigs: map[string]interface{}{
				"EXCLUDE": []interface{}{
					"resource1",
					"resource2",
				},
			},
			expectedResult: true,
		},
		{
			name:         "ExcludeConfig: Resource not excluded",
			resourceName: "resource1",
			resourceConfigs: map[string]interface{}{
				"EXCLUDE": []interface{}{
					"resource2",
					"resource3",
				},
			},
			expectedResult: false,
		},
		{
			name:            "No Config: Resource not excluded",
			resourceName:    "resource1",
			resourceConfigs: map[string]interface{}{},
			expectedResult:  false,
		},
		{
			name:         "Both Configs: Resource not excluded",
			resourceName: "resource1",
			resourceConfigs: map[string]interface{}{
				"INCLUDE_ONLY": []interface{}{
					"resource1",
				},
				"EXCLUDE": []interface{}{
					"resource1",
					"resource2",
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.IsResourceExcluded(tc.resourceName, tc.resourceConfigs)
			if result != tc.expectedResult {
				t.Errorf("Expected result to be %v but got %v", tc.expectedResult, result)
			}
		})
	}
}
