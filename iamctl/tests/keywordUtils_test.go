package tests

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func TestReplaceKeywords(t *testing.T) {

	tests := []struct {
		description    string
		fileContent    string
		keywordMapping map[string]interface{}
		expectedResult string
	}{
		{
			description: "Replace single keyword",
			fileContent: "description: This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]interface{}{
				"ENV": "dev",
			},
			expectedResult: "description: This is a sample application in the dev environment.",
		},
		{
			description: "Replace multiple keywords",
			fileContent: "description: This {{APP}} is a sample {{APP}} in the {{ENV}} environment.",
			keywordMapping: map[string]interface{}{
				"ENV": "dev",
				"APP": "application",
			},
			expectedResult: "description: This application is a sample application in the dev environment.",
		},
		{
			// If the {{}} syntax is used for other purposes, it should not be replaced.
			description: "Ignore keyword markers without a mapping",
			fileContent: "description: This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]interface{}{
				"APP": "application",
			},
			expectedResult: "description: This is a sample application in the {{ENV}} environment.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			result := utils.ReplaceKeywords(tc.fileContent, tc.keywordMapping)

			if result != tc.expectedResult {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", tc.description, tc.expectedResult, result)
			}
		})
	}
}

func TestResolveAdvancedKeywordMapping(t *testing.T) {

	testCases := []struct {
		description    string
		resourceName   string
		keywordConfig  utils.KeywordConfigs
		expectedResult map[string]interface{}
	}{
		{
			description:  "Test with advanced keyword mapping",
			resourceName: "App1",
			keywordConfig: utils.KeywordConfigs{
				KeywordMappings: map[string]interface{}{
					"CALLBACK_DOMAIN": "dev.env",
				},
				ApplicationConfigs: map[string]interface{}{
					"App1": map[string]interface{}{
						"KEYWORD_MAPPINGS": map[string]interface{}{
							"CALLBACK_DOMAIN": "dev-app1.env",
						},
					},
				},
			},
			expectedResult: map[string]interface{}{
				"CALLBACK_DOMAIN": "dev-app1.env",
			},
		},
		{
			description:  "Test only with default keyword mapping",
			resourceName: "App2",
			keywordConfig: utils.KeywordConfigs{
				KeywordMappings: map[string]interface{}{
					"CALLBACK_DOMAIN": "dev.env",
				},
				ApplicationConfigs: map[string]interface{}{
					"App1": map[string]interface{}{
						"KEYWORD_MAPPINGS": map[string]interface{}{
							"CALLBACK_DOMAIN": "dev-app1.env",
						},
					},
				},
			},
			expectedResult: map[string]interface{}{
				"CALLBACK_DOMAIN": "dev.env",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			utils.KEYWORD_CONFIGS = tc.keywordConfig
			result := utils.ResolveAdvancedKeywordMapping(tc.resourceName, tc.keywordConfig.ApplicationConfigs)
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", tc.description, tc.expectedResult, result)
			}
		})
	}

}

func TestContainsKeywords(t *testing.T) {
	tests := []struct {
		description    string
		data           string
		keywordMapping map[string]interface{}
		expectedResult bool
	}{
		{
			description: "Test with a single keyword",
			data:        "This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]interface{}{
				"ENV": "dev",
			},
			expectedResult: true,
		},
		{
			description: "Test without a keyword",
			data:        "This is a sample application in the dev environment.",
			keywordMapping: map[string]interface{}{
				"ENV": "dev",
			},
			expectedResult: false,
		},
		{
			description: "Test with a keyword, but without a mapping",
			data:        "This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]interface{}{
				"APP": "Application",
			},
			expectedResult: false,
		},
		{
			description: "Test with an empty string",
			data:        "",
			keywordMapping: map[string]interface{}{
				"ENV": "dev",
			},
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := utils.ContainsKeywords(test.data, test.keywordMapping)
			if result != test.expectedResult {
				t.Errorf("Unexpected result for %s: expected %v, but got %v", test.description, test.expectedResult, result)
			}
		})
	}
}

func TestGetKeywordLocations(t *testing.T) {
	fileData := map[interface{}]interface{}{
		"description": "A sample string with a {{KEYWORD}}.",
		"stringArray": []interface{}{"Sample string with a {{KEYWORD}}", "Sample string without a keyword"},
		"nestedObject": map[interface{}]interface{}{
			"subKey1": map[interface{}]interface{}{
				"subSubKey1": "A sample string without a keyword.",
			},
			"subKey2": map[interface{}]interface{}{
				"subSubKey1": "A sample string with a {{KEYWORD}}.",
			},
		},
		"properties": []interface{}{
			map[interface{}]interface{}{
				"name":    "element1",
				"subKey1": "Sample string with a {{KEYWORD}}",
				"subKey2": "Sample string without a keyword",
			},
			map[interface{}]interface{}{
				"name":    "element2",
				"subKey1": "Sample string without a keyword",
			},
		},
		// Array with the identifier "name" but is not defined in the arrayIdentifiers map.
		"array1": []interface{}{
			map[interface{}]interface{}{
				"name":    "element1",
				"subKey1": "Sample string with a keyword",
				"subKey2": "Sample string without a {{KEYWORD}}",
			},
		},
		// Array with a different identifier which is not defined in the arrayIdentifiers map.
		"array2": []interface{}{
			map[interface{}]interface{}{
				"id":      "id1",
				"subKey1": "Sample string with a {{KEYWORD}}",
				"subKey2": "Sample string without a keyword",
			},
		},
		"key1": []interface{}{},
		"claimMappings": []interface{}{
			map[interface{}]interface{}{
				"defaultValue": "some string",
				"localClaim": map[interface{}]interface{}{
					"claimId":  1,
					"claimUri": "http://wso2.org/claims/identity/accountLocked",
				},
				"remoteClaim": map[interface{}]interface{}{
					"claimId":  0,
					"claimUri": "http://wso2.org/claims/identity/{{KEYWORD}}",
				},
			},
		},
	}
	keywordMapping := map[string]interface{}{
		"KEYWORD": "value",
	}
	expectedResult := []string{
		"description",
		"stringArray",
		"nestedObject.subKey2.subSubKey1",
		"properties.[name=element1].subKey1",
		"array1.[name=element1].subKey2",
		"claimMappings.[localClaim.claimUri=http://wso2.org/claims/identity/accountLocked].remoteClaim.claimUri",
	}

	result := utils.GetKeywordLocations(fileData, []string{}, keywordMapping, configs.APPLICATIONS)

	sort.Strings(result)
	sort.Strings(expectedResult)

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Unexpected result: expected %v, but got %v", expectedResult, result)
	}
}

func TestGetValue(t *testing.T) {

	data := map[interface{}]interface{}{
		"description": "A sample description",
		"stringArray": []interface{}{"Sample string with a {{KEYWORD}}", "Sample string without a keyword"},
		"nestedObject": map[interface{}]interface{}{
			"subKey1": map[interface{}]interface{}{
				"subSubKey1": "A sample string without a keyword.",
			},
			"subKey2": map[interface{}]interface{}{
				"subSubKey1": "A sample string with a {{KEYWORD}}.",
			},
		},
		"properties": []interface{}{
			map[interface{}]interface{}{
				"name":    "element1",
				"subKey1": "Sample string with a {{KEYWORD}}",
				"subKey2": "Sample string without a keyword",
			},
			map[interface{}]interface{}{
				"name":    "element2",
				"subKey1": "Sample string without a keyword",
			},
		},
		"key1": []interface{}{},
		"claimMappings": []interface{}{
			map[interface{}]interface{}{
				"defaultValue": "some string",
				"localClaim": map[interface{}]interface{}{
					"claimId": 1,
					"key": map[interface{}]interface{}{
						"claimUri": "http://wso2.org/claims/identity/accountLocked",
					},
				},
				"remoteClaim": map[interface{}]interface{}{
					"claimId":  0,
					"claimUri": "http://wso2.org/claims/identity/{{KEYWORD}}",
				},
			},
		},
	}
	testCases := []struct {
		path           string
		expectedResult string
	}{
		{
			path:           "description",
			expectedResult: "A sample description",
		},
		{
			path:           "stringArray",
			expectedResult: "Sample string with a {{KEYWORD}},Sample string without a keyword",
		},
		{
			path:           "nestedObject.subKey2.subSubKey1",
			expectedResult: "A sample string with a {{KEYWORD}}.",
		},
		{
			path:           "properties.[name=element1].subKey2",
			expectedResult: "Sample string without a keyword",
		},
		{
			path:           "properties.[name=element3].subKey1",
			expectedResult: "",
		},
		{
			path:           "key1.[name=element1].subKey2",
			expectedResult: "",
		},
		{
			path:           "claimMappings.[localClaim.key.claimUri=http://wso2.org/claims/identity/accountLocked].remoteClaim.claimUri",
			expectedResult: "http://wso2.org/claims/identity/{{KEYWORD}}",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			result := utils.GetValue(data, tc.path)
			if result != tc.expectedResult {
				t.Errorf("Unexpected result: expected %v, but got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestReplaceValue(t *testing.T) {

	testCases := []struct {
		data           interface{}
		path           string
		replacement    string
		expectedResult interface{}
	}{
		{
			data: map[interface{}]interface{}{
				"description": "A sample description",
			},
			path:        "description",
			replacement: "A sample description of the dev environment",
			expectedResult: map[interface{}]interface{}{
				"description": "A sample description of the dev environment",
			},
		},
		{
			data: map[interface{}]interface{}{
				"stringArray": []interface{}{"Sample string with a {{KEYWORD}}", "Sample string without a keyword"},
			},
			path:        "stringArray",
			replacement: "Sample string with a replaced keyword,Sample string without a keyword",
			expectedResult: map[interface{}]interface{}{
				"stringArray": "Sample string with a replaced keyword,Sample string without a keyword",
			},
		},
		{
			data: map[interface{}]interface{}{
				"nestedObject": map[interface{}]interface{}{
					"subKey1": map[interface{}]interface{}{
						"subSubKey1": "A sample string without a keyword.",
					},
					"subKey2": map[interface{}]interface{}{
						"subSubKey1": "A sample string with a {{KEYWORD}}.",
					},
				},
			},
			path:        "nestedObject.subKey2.subSubKey1",
			replacement: "Sample string with the replaced keyword",
			expectedResult: map[interface{}]interface{}{
				"nestedObject": map[interface{}]interface{}{
					"subKey1": map[interface{}]interface{}{
						"subSubKey1": "A sample string without a keyword.",
					},
					"subKey2": map[interface{}]interface{}{
						"subSubKey1": "Sample string with the replaced keyword",
					},
				},
			},
		},
		{
			data: map[interface{}]interface{}{
				"properties": []interface{}{
					map[interface{}]interface{}{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string without a keyword",
					},
					map[interface{}]interface{}{
						"name":    "element2",
						"subKey1": "Sample string without a keyword",
					},
				},
			},
			path:        "properties.[name=element1].subKey2",
			replacement: "Sample string with the added {{KEYWORD}}",
			expectedResult: map[interface{}]interface{}{
				"properties": []interface{}{
					map[interface{}]interface{}{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string with the added {{KEYWORD}}",
					},
					map[interface{}]interface{}{
						"name":    "element2",
						"subKey1": "Sample string without a keyword",
					},
				},
			},
		},
		{
			data: map[interface{}]interface{}{
				"properties": []interface{}{
					map[interface{}]interface{}{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string without a keyword",
					},
				},
			},
			path:        "properties.[name=element3].subKey1",
			replacement: "Sample string in array element which does not exist",
			expectedResult: map[interface{}]interface{}{
				"properties": []interface{}{
					map[interface{}]interface{}{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string without a keyword",
					},
				},
			},
		},
		{
			data: map[interface{}]interface{}{
				"key1": []interface{}{},
			},
			path:        "key1.[name=element1].subKey2",
			replacement: "Sample string in object which does not exist",
			expectedResult: map[interface{}]interface{}{
				"key1": []interface{}{},
			},
		},
		{
			data: map[interface{}]interface{}{
				"claimMappings": []interface{}{
					map[interface{}]interface{}{
						"defaultValue": "some string",
						"localClaim": map[interface{}]interface{}{
							"claimId":  1,
							"claimUri": "http://wso2.org/claims/identity/accountLocked",
						},
						"remoteClaim": map[interface{}]interface{}{
							"claimId":  0,
							"claimUri": "http://wso2.org/claims/identity/{{KEYWORD}}",
						},
					},
				},
			},
			path:        "claimMappings.[localClaim.claimUri=http://wso2.org/claims/identity/accountLocked].remoteClaim.claimUri",
			replacement: "http://wso2.org/claims/identity/replacedKeyword",
			expectedResult: map[interface{}]interface{}{
				"claimMappings": []interface{}{
					map[interface{}]interface{}{
						"defaultValue": "some string",
						"localClaim": map[interface{}]interface{}{
							"claimId":  1,
							"claimUri": "http://wso2.org/claims/identity/accountLocked",
						},
						"remoteClaim": map[interface{}]interface{}{
							"claimId":  0,
							"claimUri": "http://wso2.org/claims/identity/replacedKeyword",
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			result := utils.ReplaceValue(tc.data, tc.path, tc.replacement)
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("Unexpected result: expected %v, but got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestGetArrayIndex(t *testing.T) {
	arrayMap := []interface{}{
		map[interface{}]interface{}{
			"name":     "element1",
			"property": "property value 1",
		},
		map[interface{}]interface{}{
			"name":     "element2",
			"property": "property value 2",
		},
	}

	testCases := []struct {
		elementIdentifier string
		expectedIndex     int
	}{
		{
			elementIdentifier: "[name=element2]",
			expectedIndex:     1,
		},
		{
			elementIdentifier: "[name=element3]",
			expectedIndex:     -1,
		},
		{
			elementIdentifier: "[id=element1]",
			expectedIndex:     -1,
		},
		{
			elementIdentifier: "name=element1",
			expectedIndex:     -1,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			index, _ := utils.GetArrayIndex(arrayMap, tc.elementIdentifier)
			if index != tc.expectedIndex {
				t.Errorf("Unexpected result: expected %v, but got %v", tc.expectedIndex, index)
			}
		})
	}
}

func TestAddKeywords(t *testing.T) {

	exportedFileData := map[string]interface{}{
		"key1": "A sample string in the dev environment with the keyword value",
		"key2": "A different string in the dev environment",
		"key3": "A string with a the placeholder {{syntax}} that should not be changed",
		"nestedObject": map[string]interface{}{
			"subKey1": map[string]interface{}{
				"subSubKey1": "A sample string without a keyword.",
				"subSubKey2": "A sample string with the keyword value.",
			},
		},
		"properties": []interface{}{
			map[string]interface{}{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the keyword value",
			},
			map[string]interface{}{
				"name":    "element2",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string 2",
			},
		},
	}

	localFileData := []byte(`
        key1: A sample string in the {{ENV}} environment with the {{KEYWORD}}
        key2: A string with a {{KEYWORD}} that is changed
        key3: A string with a the placeholder {{syntax}} that should not be changed
        nestedObject:
          subKey1:
            subSubKey1: A sample string without a keyword.
            subSubKey2: A sample string with the {{KEYWORD}}.
        properties:
          - name: element1
            subKey1: A sample string 1
            subKey2: A sample string with the {{KEYWORD}}
          - name: element2
            subKey1: A sample string 1
            subKey2: A sample string 2
        `)

	expectedExportedFileData := map[string]interface{}{
		"key1": "A sample string in the {{ENV}} environment with the {{KEYWORD}}",
		"key2": "A different string in the dev environment",
		"key3": "A string with a the placeholder {{syntax}} that should not be changed",
		"nestedObject": map[string]interface{}{
			"subKey1": map[string]interface{}{
				"subSubKey1": "A sample string without a keyword.",
				"subSubKey2": "A sample string with the {{KEYWORD}}.",
			},
		},
		"properties": []interface{}{
			map[string]interface{}{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the {{KEYWORD}}",
			},
			map[string]interface{}{
				"name":    "element2",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string 2",
			},
		},
	}

	keywordMapping := map[string]interface{}{
		"ENV":     "dev",
		"KEYWORD": "keyword value",
	}
	result, err := utils.AddKeywords(exportedFileData, localFileData, keywordMapping, configs.APPLICATIONS)
	if err != nil {
		log.Println("Error when adding keywords: ", err)
	}

	if !reflect.DeepEqual(result, expectedExportedFileData) {
		t.Errorf("Expected %+v, but got %+v", expectedExportedFileData, result)
	}
}

func TestModifyFieldsWithKeywords(t *testing.T) {

	keywordLocations := []string{"key1", "nestedObject.subKey1.subSubKey2", "properties.[name=element1].subKey2"}
	keywordMap := map[string]interface{}{
		"KEYWORD":  "keyword value",
		"keyword2": "replacement2",
	}
	localFileData := map[string]interface{}{
		"key1": "A string with the {{KEYWORD}}",
		"nestedObject": map[string]interface{}{
			"subKey1": map[string]interface{}{
				"subSubKey1": "A string with a placeholder {{syntax}}",
				"subSubKey2": "A string with the {{KEYWORD}}",
			},
		},
		"properties": []interface{}{
			map[string]interface{}{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the {{KEYWORD}}",
			},
		},
	}
	exportedFileData := map[string]interface{}{
		"key1": "A string with the keyword value",
		"nestedObject": map[string]interface{}{
			"subKey1": map[string]interface{}{
				"subSubKey1": "A string with a placeholder {{syntax}}",
				"subSubKey2": "A string with a change",
			},
		},
		"properties": []interface{}{
			map[string]interface{}{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the keyword value",
			},
		},
	}
	expectedExportedFileData := map[string]interface{}{
		"key1": "A string with the {{KEYWORD}}",
		"nestedObject": map[string]interface{}{
			"subKey1": map[string]interface{}{
				"subSubKey1": "A string with a placeholder {{syntax}}",
				"subSubKey2": "A string with a change",
			},
		},
		"properties": []interface{}{
			map[string]interface{}{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the {{KEYWORD}}",
			},
		},
	}

	result := utils.ModifyFieldsWithKeywords(exportedFileData, localFileData, keywordLocations, keywordMap)

	if !reflect.DeepEqual(result, expectedExportedFileData) {
		t.Errorf("Expected %+v, but got %+v", expectedExportedFileData, result)
	}
}

func TestGetPathKeys(t *testing.T) {

	testCases := []struct {
		path     string
		expected []string
	}{
		{
			path:     "key1",
			expected: []string{"key1"},
		},
		{
			path:     "nestedObject.subKey1.subSubKey2",
			expected: []string{"nestedObject", "subKey1", "subSubKey2"},
		},
		{
			path:     "properties.[name=element1.element2].subKey2",
			expected: []string{"properties", "[name=element1.element2]", "subKey2"},
		},
		{
			path:     "federatedAuthenticatorConfigs.[name=GoogleOIDCAuthenticator].properties.[name=ClientSecret].value",
			expected: []string{"federatedAuthenticatorConfigs", "[name=GoogleOIDCAuthenticator]", "properties", "[name=ClientSecret]", "value"},
		},
	}

	for _, testCase := range testCases {
		result := utils.GetPathKeys(testCase.path)
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("Expected %+v, but got %+v", testCase.expected, result)
		}
	}
}
