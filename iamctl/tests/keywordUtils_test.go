package tests

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func TestReplaceKeywords(t *testing.T) {

	tests := []struct {
		description    string
		fileContent    string
		keywordMapping map[string]any
		expectedResult string
	}{
		{
			description: "Replace single keyword",
			fileContent: "description: This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]any{
				"ENV": "dev",
			},
			expectedResult: "description: This is a sample application in the dev environment.",
		},
		{
			description: "Replace multiple keywords",
			fileContent: "description: This {{APP}} is a sample {{APP}} in the {{ENV}} environment.",
			keywordMapping: map[string]any{
				"ENV": "dev",
				"APP": "application",
			},
			expectedResult: "description: This application is a sample application in the dev environment.",
		},
		{
			// If the {{}} syntax is used for other purposes, it should not be replaced.
			description: "Ignore keyword markers without a mapping",
			fileContent: "description: This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]any{
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
		expectedResult map[string]any
	}{
		{
			description:  "Test with advanced keyword mapping",
			resourceName: "App1",
			keywordConfig: utils.KeywordConfigs{
				KeywordMappings: map[string]any{
					"CALLBACK_DOMAIN": "dev.env",
				},
				ApplicationConfigs: map[string]any{
					"App1": map[string]any{
						"KEYWORD_MAPPINGS": map[string]any{
							"CALLBACK_DOMAIN": "dev-app1.env",
						},
					},
				},
			},
			expectedResult: map[string]any{
				"CALLBACK_DOMAIN": "dev-app1.env",
			},
		},
		{
			description:  "Test only with default keyword mapping",
			resourceName: "App2",
			keywordConfig: utils.KeywordConfigs{
				KeywordMappings: map[string]any{
					"CALLBACK_DOMAIN": "dev.env",
				},
				ApplicationConfigs: map[string]any{
					"App1": map[string]any{
						"KEYWORD_MAPPINGS": map[string]any{
							"CALLBACK_DOMAIN": "dev-app1.env",
						},
					},
				},
			},
			expectedResult: map[string]any{
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
		keywordMapping map[string]any
		expectedResult bool
	}{
		{
			description: "Test with a single keyword",
			data:        "This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]any{
				"ENV": "dev",
			},
			expectedResult: true,
		},
		{
			description: "Test without a keyword",
			data:        "This is a sample application in the dev environment.",
			keywordMapping: map[string]any{
				"ENV": "dev",
			},
			expectedResult: false,
		},
		{
			description: "Test with a keyword, but without a mapping",
			data:        "This is a sample application in the {{ENV}} environment.",
			keywordMapping: map[string]any{
				"APP": "Application",
			},
			expectedResult: false,
		},
		{
			description: "Test with an empty string",
			data:        "",
			keywordMapping: map[string]any{
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
	fileData := map[any]any{
		"description": "A sample string with a {{KEYWORD}}.",
		"stringArray": []any{"Sample string with a {{KEYWORD}}", "Sample string without a keyword"},
		"nestedObject": map[any]any{
			"subKey1": map[any]any{
				"subSubKey1": "A sample string without a keyword.",
			},
			"subKey2": map[any]any{
				"subSubKey1": "A sample string with a {{KEYWORD}}.",
			},
		},
		"properties": []any{
			map[any]any{
				"name":    "element1",
				"subKey1": "Sample string with a {{KEYWORD}}",
				"subKey2": "Sample string without a keyword",
			},
			map[any]any{
				"name":    "element2",
				"subKey1": "Sample string without a keyword",
			},
		},
		// Array with the identifier "name" but is not defined in the arrayIdentifiers map.
		"array1": []any{
			map[any]any{
				"name":    "element1",
				"subKey1": "Sample string with a keyword",
				"subKey2": "Sample string without a {{KEYWORD}}",
			},
		},
		// Array with a different identifier which is not defined in the arrayIdentifiers map.
		"array2": []any{
			map[any]any{
				"id":      "id1",
				"subKey1": "Sample string with a {{KEYWORD}}",
				"subKey2": "Sample string without a keyword",
			},
		},
		"key1": []any{},
		"claimMappings": []any{
			map[any]any{
				"defaultValue": "some string",
				"localClaim": map[any]any{
					"claimId":  1,
					"claimUri": "http://wso2.org/claims/identity/accountLocked",
				},
				"remoteClaim": map[any]any{
					"claimId":  0,
					"claimUri": "http://wso2.org/claims/identity/{{KEYWORD}}",
				},
			},
		},
	}
	keywordMapping := map[string]any{
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

	result := utils.GetKeywordLocations(fileData, []string{}, keywordMapping, utils.APPLICATIONS)

	sort.Strings(result)
	sort.Strings(expectedResult)

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Unexpected result: expected %v, but got %v", expectedResult, result)
	}
}

func TestGetValue(t *testing.T) {

	data := map[any]any{
		"description": "A sample description",
		"stringArray": []any{"Sample string with a {{KEYWORD}}", "Sample string without a keyword"},
		"nestedObject": map[any]any{
			"subKey1": map[any]any{
				"subSubKey1": "A sample string without a keyword.",
			},
			"subKey2": map[any]any{
				"subSubKey1": "A sample string with a {{KEYWORD}}.",
			},
		},
		"properties": []any{
			map[any]any{
				"name":    "element1",
				"subKey1": "Sample string with a {{KEYWORD}}",
				"subKey2": "Sample string without a keyword",
			},
			map[any]any{
				"name":    "element2",
				"subKey1": "Sample string without a keyword",
			},
		},
		"key1": []any{},
		"claimMappings": []any{
			map[any]any{
				"defaultValue": "some string",
				"localClaim": map[any]any{
					"claimId": 1,
					"key": map[any]any{
						"claimUri": "http://wso2.org/claims/identity/accountLocked",
					},
				},
				"remoteClaim": map[any]any{
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
		data           any
		path           string
		replacement    string
		expectedResult any
	}{
		{
			data: map[any]any{
				"description": "A sample description",
			},
			path:        "description",
			replacement: "A sample description of the dev environment",
			expectedResult: map[any]any{
				"description": "A sample description of the dev environment",
			},
		},
		{
			data: map[any]any{
				"stringArray": []any{"Sample string with a {{KEYWORD}}", "Sample string without a keyword"},
			},
			path:        "stringArray",
			replacement: "Sample string with a replaced keyword,Sample string without a keyword",
			expectedResult: map[any]any{
				"stringArray": "Sample string with a replaced keyword,Sample string without a keyword",
			},
		},
		{
			data: map[any]any{
				"nestedObject": map[any]any{
					"subKey1": map[any]any{
						"subSubKey1": "A sample string without a keyword.",
					},
					"subKey2": map[any]any{
						"subSubKey1": "A sample string with a {{KEYWORD}}.",
					},
				},
			},
			path:        "nestedObject.subKey2.subSubKey1",
			replacement: "Sample string with the replaced keyword",
			expectedResult: map[any]any{
				"nestedObject": map[any]any{
					"subKey1": map[any]any{
						"subSubKey1": "A sample string without a keyword.",
					},
					"subKey2": map[any]any{
						"subSubKey1": "Sample string with the replaced keyword",
					},
				},
			},
		},
		{
			data: map[any]any{
				"properties": []any{
					map[any]any{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string without a keyword",
					},
					map[any]any{
						"name":    "element2",
						"subKey1": "Sample string without a keyword",
					},
				},
			},
			path:        "properties.[name=element1].subKey2",
			replacement: "Sample string with the added {{KEYWORD}}",
			expectedResult: map[any]any{
				"properties": []any{
					map[any]any{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string with the added {{KEYWORD}}",
					},
					map[any]any{
						"name":    "element2",
						"subKey1": "Sample string without a keyword",
					},
				},
			},
		},
		{
			data: map[any]any{
				"properties": []any{
					map[any]any{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string without a keyword",
					},
				},
			},
			path:        "properties.[name=element3].subKey1",
			replacement: "Sample string in array element which does not exist",
			expectedResult: map[any]any{
				"properties": []any{
					map[any]any{
						"name":    "element1",
						"subKey1": "Sample string with a {{KEYWORD}}",
						"subKey2": "Sample string without a keyword",
					},
				},
			},
		},
		{
			data: map[any]any{
				"key1": []any{},
			},
			path:        "key1.[name=element1].subKey2",
			replacement: "Sample string in object which does not exist",
			expectedResult: map[any]any{
				"key1": []any{},
			},
		},
		{
			data: map[any]any{
				"claimMappings": []any{
					map[any]any{
						"defaultValue": "some string",
						"localClaim": map[any]any{
							"claimId":  1,
							"claimUri": "http://wso2.org/claims/identity/accountLocked",
						},
						"remoteClaim": map[any]any{
							"claimId":  0,
							"claimUri": "http://wso2.org/claims/identity/{{KEYWORD}}",
						},
					},
				},
			},
			path:        "claimMappings.[localClaim.claimUri=http://wso2.org/claims/identity/accountLocked].remoteClaim.claimUri",
			replacement: "http://wso2.org/claims/identity/replacedKeyword",
			expectedResult: map[any]any{
				"claimMappings": []any{
					map[any]any{
						"defaultValue": "some string",
						"localClaim": map[any]any{
							"claimId":  1,
							"claimUri": "http://wso2.org/claims/identity/accountLocked",
						},
						"remoteClaim": map[any]any{
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
	arrayMap := []any{
		map[any]any{
			"name":     "element1",
			"property": "property value 1",
		},
		map[any]any{
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

func TestAddKeywordsYAML(t *testing.T) {

	exportedFileData := map[string]any{
		"key1": "A sample string in the dev environment with the keyword value",
		"key2": "A different string in the dev environment",
		"key3": "A string with a the placeholder {{syntax}} that should not be changed",
		"nestedObject": map[string]any{
			"subKey1": map[string]any{
				"subSubKey1": "A sample string without a keyword.",
				"subSubKey2": "A sample string with the keyword value.",
			},
		},
		"properties": []any{
			map[string]any{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the keyword value",
			},
			map[string]any{
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

	expectedExportedFileData := map[string]any{
		"key1": "A sample string in the {{ENV}} environment with the {{KEYWORD}}",
		"key2": "A different string in the dev environment",
		"key3": "A string with a the placeholder {{syntax}} that should not be changed",
		"nestedObject": map[string]any{
			"subKey1": map[string]any{
				"subSubKey1": "A sample string without a keyword.",
				"subSubKey2": "A sample string with the {{KEYWORD}}.",
			},
		},
		"properties": []any{
			map[string]any{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the {{KEYWORD}}",
			},
			map[string]any{
				"name":    "element2",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string 2",
			},
		},
	}

	keywordMapping := map[string]any{
		"ENV":     "dev",
		"KEYWORD": "keyword value",
	}
	result, err := utils.AddKeywords(exportedFileData, ".yaml", localFileData, keywordMapping, utils.APPLICATIONS)
	if err != nil {
		log.Println("Error when adding keywords: ", err)
	}

	if !reflect.DeepEqual(result, expectedExportedFileData) {
		t.Errorf("Expected %+v, but got %+v", expectedExportedFileData, result)
	}
}

func TestAddKeywordsJSON(t *testing.T) {

	exportedFileData := map[string]any{
		"key1": "A sample string in the dev environment with the keyword value",
		"key2": "A different string in the dev environment",
		"key3": "A string with a the placeholder {{syntax}} that should not be changed",
		"nestedObject": map[string]any{
			"subKey1": map[string]any{
				"subSubKey1": "A sample string without a keyword.",
				"subSubKey2": "A sample string with the keyword value.",
			},
		},
		"properties": []any{
			map[string]any{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the keyword value",
			},
			map[string]any{
				"name":    "element2",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string 2",
			},
		},
	}

	localFileData := []byte(`{
		"key1": "A sample string in the {{ENV}} environment with the {{KEYWORD}}",
		"key2": "A string with a {{KEYWORD}} that is changed",
		"key3": "A string with a the placeholder {{syntax}} that should not be changed",
		"nestedObject": {
			"subKey1": {
				"subSubKey1": "A sample string without a keyword.",
				"subSubKey2": "A sample string with the {{KEYWORD}}."
			}
		},
		"properties": [
			{
				"name": "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the {{KEYWORD}}"
			},
			{
				"name": "element2",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string 2"
			}
		]
	}`)
	expectedExportedFileData := map[string]any{
		"key1": "A sample string in the {{ENV}} environment with the {{KEYWORD}}",
		"key2": "A different string in the dev environment",
		"key3": "A string with a the placeholder {{syntax}} that should not be changed",
		"nestedObject": map[string]any{
			"subKey1": map[string]any{
				"subSubKey1": "A sample string without a keyword.",
				"subSubKey2": "A sample string with the {{KEYWORD}}.",
			},
		},
		"properties": []any{
			map[string]any{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the {{KEYWORD}}",
			},
			map[string]any{
				"name":    "element2",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string 2",
			},
		},
	}

	keywordMapping := map[string]any{
		"ENV":     "dev",
		"KEYWORD": "keyword value",
	}
	result, err := utils.AddKeywords(exportedFileData, ".json", localFileData, keywordMapping, utils.APPLICATIONS)
	if err != nil {
		log.Println("Error when adding keywords: ", err)
	}

	if !reflect.DeepEqual(result, expectedExportedFileData) {
		t.Errorf("Expected %+v, but got %+v", expectedExportedFileData, result)
	}
}

func TestAddKeywordsXML(t *testing.T) {
	exportedFileData := map[string]any{
		"root": map[string]any{
			"key1": "A sample string in the dev environment with the keyword value",
			"key2": "A different string in the dev environment",
			"key3": "A string with a the placeholder {{syntax}} that should not be changed",
			"nestedObject": map[string]any{
				"subKey1": map[string]any{
					"subSubKey1": "A sample string without a keyword.",
					"subSubKey2": "A sample string with the keyword value.",
				},
			},
			"properties": map[string]any{ // XML Lists keep the wrapper in MXJ
				"property": []any{
					map[string]any{
						"name":    "element1",
						"subKey1": "A sample string 1",
						"subKey2": "A sample string with the keyword value",
					},
					map[string]any{
						"name":    "element2",
						"subKey1": "A sample string 1",
						"subKey2": "A sample string 2",
					},
				},
			},
		},
	}

	localFileData := []byte("<root>" +
		"<key1>A sample string in the {{ENV}} environment with the {{KEYWORD}}</key1>" +
		"<key2>A sample string with a {{KEYWORD}} that is changed</key2>" +
		"<key3>A string with a the placeholder {{syntax}} that should not be changed</key3>" +
		"<nestedObject>" +
		"<subKey1>" +
		"<subSubKey1>A sample string without a keyword.</subSubKey1>" +
		"<subSubKey2>A sample string with the {{KEYWORD}}.</subSubKey2>" +
		"</subKey1>" +
		"</nestedObject>" +
		"<properties>" +
		"<property>" +
		"<name>element1</name>" +
		"<subKey1>A sample string 1</subKey1>" +
		"<subKey2>A sample string with the {{KEYWORD}}</subKey2>" +
		"</property>" +
		"<property>" +
		"<name>element2</name>" +
		"<subKey1>A sample string 1</subKey1>" +
		"<subKey2>A sample string 2</subKey2>" +
		"</property>" +
		"</properties>" +
		"</root>")

	expectedExportedFileData := map[string]any{
		"root": map[string]any{
			"key1": "A sample string in the {{ENV}} environment with the {{KEYWORD}}",
			"key2": "A different string in the dev environment",
			"key3": "A string with a the placeholder {{syntax}} that should not be changed",
			"nestedObject": map[string]any{
				"subKey1": map[string]any{
					"subSubKey1": "A sample string without a keyword.",
					"subSubKey2": "A sample string with the {{KEYWORD}}.",
				},
			},
			"properties": map[string]any{
				"property": []any{
					map[string]any{
						"name":    "element1",
						"subKey1": "A sample string 1",
						"subKey2": "A sample string with the {{KEYWORD}}",
					},
					map[string]any{
						"name":    "element2",
						"subKey1": "A sample string 1",
						"subKey2": "A sample string 2",
					},
				},
			},
		},
	}

	keywordMapping := map[string]any{
		"ENV":     "dev",
		"KEYWORD": "keyword value",
	}

	result, err := utils.AddKeywords(exportedFileData, ".xml", localFileData, keywordMapping, utils.APPLICATIONS)
	if err != nil {
		t.Fatalf("AddKeywords failed: %v", err)
	}

	if !reflect.DeepEqual(result, expectedExportedFileData) {
		t.Errorf("Mismatch found!\nExpected: %+v\nGot: %+v", expectedExportedFileData, result)
	}
}
func TestModifyFieldsWithKeywords(t *testing.T) {

	keywordLocations := []string{"key1", "nestedObject.subKey1.subSubKey2", "properties.[name=element1].subKey2"}
	keywordMap := map[string]any{
		"KEYWORD":  "keyword value",
		"keyword2": "replacement2",
	}
	localFileData := map[string]any{
		"key1": "A string with the {{KEYWORD}}",
		"nestedObject": map[string]any{
			"subKey1": map[string]any{
				"subSubKey1": "A string with a placeholder {{syntax}}",
				"subSubKey2": "A string with the {{KEYWORD}}",
			},
		},
		"properties": []any{
			map[string]any{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the {{KEYWORD}}",
			},
		},
	}
	exportedFileData := map[string]any{
		"key1": "A string with the keyword value",
		"nestedObject": map[string]any{
			"subKey1": map[string]any{
				"subSubKey1": "A string with a placeholder {{syntax}}",
				"subSubKey2": "A string with a change",
			},
		},
		"properties": []any{
			map[string]any{
				"name":    "element1",
				"subKey1": "A sample string 1",
				"subKey2": "A sample string with the keyword value",
			},
		},
	}
	expectedExportedFileData := map[string]any{
		"key1": "A string with the {{KEYWORD}}",
		"nestedObject": map[string]any{
			"subKey1": map[string]any{
				"subSubKey1": "A string with a placeholder {{syntax}}",
				"subSubKey2": "A string with a change",
			},
		},
		"properties": []any{
			map[string]any{
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

func TestFixXmlStructure(t *testing.T) {
	xmlData := []byte(`<root>
<key1>Value1</key1>
<value> SELECT * FROM table WHERE column <= 10; </value>
<key3 xsi="val">Text</key3>
</root>`)

	expectedContent := `<root>
<key1>Value1</key1>
<value><![CDATA[ SELECT * FROM table WHERE column <= 10; ]]></value>
<key3 xmlns:xsi="val">Text</key3>
</root>`

	fixedData := utils.FixXmlStructure(xmlData)

	// Normalize: Remove tabs, spaces, and newlines for a robust comparison
	normalize := func(s string) string {
		re := regexp.MustCompile(`\s+`)
		return strings.TrimSpace(re.ReplaceAllString(s, ""))
	}

	if normalize(string(fixedData)) != normalize(expectedContent) {
		t.Errorf("Mismatch found!\nExpected: %s\nGot: %s", expectedContent, string(fixedData))
	}
}

var XML_TO_MAP_TEST_CASES = map[string]map[string]any{
	"<root><key1>Value1</key1><key2>Value2</key2></root>": {
		"root": map[string]any{
			"key1": "Value1",
			"key2": "Value2",
		},
	},
	`<root>
		<parent>
			<child1>Value1</child1>
			<child2>Value2</child2>
		</parent>
	</root>`: {
		"root": map[string]any{
			"parent": map[string]any{
				"child1": "Value1",
				"child2": "Value2",
			},
		},
	},
}

func TestXMLToMap(t *testing.T) {
	for xmlInput, expectedMap := range XML_TO_MAP_TEST_CASES {
		resultMap, err := utils.XMLToMap([]byte(xmlInput))
		if err != nil {
			t.Errorf("Error converting XML to map: %v", err)
		}
		if !reflect.DeepEqual(resultMap, expectedMap) {
			t.Errorf("Mismatch found!\nExpected: %+v\nGot: %+v", expectedMap, resultMap)
		}
	}
}
