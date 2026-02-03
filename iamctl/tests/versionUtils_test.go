/*
 * Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
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
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package tests

import (
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
		hasError bool
	}{
		// Valid comparisons
		{"v1 greater major", "7.0.0", "6.1.0", 1, false},
		{"v1 less major", "6.1.0", "7.0.0", -1, false},
		{"equal versions", "7.0.0", "7.0.0", 0, false},
		{"v1 greater minor", "7.1.0", "7.0.9", 1, false},
		{"v1 less minor", "7.0.0", "7.1.0", -1, false},
		{"v1 greater patch", "7.0.10", "7.0.9", 1, false},
		{"v1 less patch", "7.0.9", "7.0.10", -1, false},
		{"double digit major", "10.0.0", "9.0.0", 1, false},
		{"double digit minor", "7.10.0", "7.9.0", 1, false},
		{"double digit patch", "7.0.20", "7.0.19", 1, false},

		// Invalid formats
		{"missing patch", "7.0", "7.0.0", 0, true},
		{"too many components", "7.0.0.1", "7.0.0", 0, true},
		{"non-integer major", "abc.0.0", "7.0.0", 0, true},
		{"non-integer minor", "7.abc.0", "7.0.0", 0, true},
		{"non-integer patch", "7.0.abc", "7.0.0", 0, true},
		{"negative major", "-7.0.0", "7.0.0", 0, true},
		{"negative minor", "7.-1.0", "7.0.0", 0, true},
		{"negative patch", "7.0.-1", "7.0.0", 0, true},
		{"empty version", "", "7.0.0", 0, true},
		{"whitespace only", "   ", "7.0.0", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.CompareVersions(tt.v1, tt.v2)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for versions %s vs %s, got nil", tt.v1, tt.v2)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for versions %s vs %s: %v", tt.v1, tt.v2, err)
				}
				if result != tt.expected {
					t.Errorf("compareVersions(%s, %s) = %d; expected %d", tt.v1, tt.v2, result, tt.expected)
				}
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected [3]int
		hasError bool
	}{
		{"valid version", "7.0.0", [3]int{7, 0, 0}, false},
		{"double digit major", "10.5.3", [3]int{10, 5, 3}, false},
		{"version with spaces", " 7.0.0 ", [3]int{7, 0, 0}, false},
		{"component spaces", "7. 0 .0", [3]int{7, 0, 0}, false},

		{"missing component", "7.0", [3]int{0, 0, 0}, true},
		{"extra component", "7.0.0.1", [3]int{0, 0, 0}, true},
		{"non-integer", "7.x.0", [3]int{0, 0, 0}, true},
		{"negative", "7.-1.0", [3]int{0, 0, 0}, true},
		{"empty", "", [3]int{0, 0, 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ParseVersion(tt.version)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for version %s, got nil", tt.version)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for version %s: %v", tt.version, err)
				}
				if result != tt.expected {
					t.Errorf("parseVersion(%s) = %v; expected %v", tt.version, result, tt.expected)
				}
			}
		})
	}
}

func TestIsEntitySupportedInVersion(t *testing.T) {
	// Save original version config
	originalVersion := utils.SERVER_CONFIGS.Version
	defer func() {
		utils.SERVER_CONFIGS.Version = originalVersion
	}()

	tests := []struct {
		name         string
		version      string
		resourceType utils.ResourceType
		expected     bool
	}{
		{"no version", "", utils.APPLICATIONS, true},
		{"sufficient version", "7.0.0", utils.APPLICATIONS, true},
		{"exact version", "5.9.0", utils.USERSTORES, true},
		{"insufficient version", "5.0.0", utils.USERSTORES, false},
		{"invalid version", "invalid", utils.CLAIMS, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.SERVER_CONFIGS.Version = tt.version
			result := utils.IsEntitySupportedInVersion(tt.resourceType)

			if result != tt.expected {
				t.Errorf("IsEntitySupportedInVersion(%s) with version %s = %v; expected %v",
					tt.resourceType, tt.version, result, tt.expected)
			}
		})
	}
}
