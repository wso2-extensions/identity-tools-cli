package tests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func TestLoadToolConfigsFromFile(t *testing.T) {

	tmpfile, err := ioutil.TempFile("", "config-dev.json")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some test JSON data to the file
	testData := `{
		"KEYWORD_MAPPINGS" : {
			"keyword1" : "value1",
			"keyword2" : "value2"
		}
	}`
	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatalf("Failed to write test data to temporary file: %v", err)
	}

	// Close the file before calling the function to ensure it's flushed to disk
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}
	toolConfigs := utils.LoadToolConfigsFromFile(tmpfile.Name())

	expectedConfigs := utils.ToolConfigs{
		KeywordMappings: map[string]interface{}{
			"keyword1": "value1",
			"keyword2": "value2",
		},
		ApplicationConfigs: map[string]interface{}(nil),
	}
	if !areEqual(toolConfigs, expectedConfigs) {
		t.Errorf("Server configs did not match expected values:\nGot: %#v\nExpected: %#v", toolConfigs, expectedConfigs)
	}

}

func areEqual(a, b utils.ToolConfigs) bool {
	// Convert to JSON and compare as strings to handle nested maps
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

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
