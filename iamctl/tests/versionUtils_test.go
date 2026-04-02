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
		{"2-digit equal to 3-digit", "7.2", "7.2.0", 0, false},
		{"2-digit vs 3-digit greater", "7.2", "7.1.5", 1, false},
		{"2-digit vs 3-digit less", "7.1", "7.2.0", -1, false},
		{"both 2-digit", "7.0", "7.0", 0, false},

		// Invalid formats
		{"missing minor", "7", "7.0.0", 0, true},
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
		{"2-digit version", "7.2", [3]int{7, 2, 0}, false},
		{"2-digit with zeros", "5.0", [3]int{5, 0, 0}, false},
		{"2-digit with spaces", " 7.2 ", [3]int{7, 2, 0}, false},

		{"missing minor", "7", [3]int{0, 0, 0}, true},
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
	originalVersion := utils.SERVER_CONFIGS.ServerVersion
	originalMinVersions := utils.EntityMinVersionRequirements
	originalMaxVersions := utils.EntityMaxSupportedVersion
	defer func() {
		utils.SERVER_CONFIGS.ServerVersion = originalVersion
		utils.EntityMinVersionRequirements = originalMinVersions
		utils.EntityMaxSupportedVersion = originalMaxVersions
	}()

	const testResource = utils.APPLICATIONS

	tests := []struct {
		name       string
		version    string
		minVersion string
		maxVersion string
		expected   bool
	}{
		// No server version configured / no version requirements
		{"no version configured", "", "6.0.0", "7.0.0", true},
		{"no version requirement", "5.0.0", "", "", true},

		// Min version checks
		{"version above min", "7.0.0", "6.0.0", "", true},
		{"version equals min", "6.0.0", "6.0.0", "", true},
		{"version below min", "5.9.0", "6.0.0", "", false},

		// Max version checks
		{"version below max", "6.0.0", "", "7.0.0", true},
		{"version equals max", "7.0.0", "", "7.0.0", true},
		{"version above max", "7.1.0", "", "7.0.0", false},

		// Both min and max bounds
		{"version within range", "6.5.0", "6.0.0", "7.0.0", true},
		{"version equals min of range", "6.0.0", "6.0.0", "7.0.0", true},
		{"version equals max of range", "7.0.0", "6.0.0", "7.0.0", true},
		{"version below range", "5.9.0", "6.0.0", "7.0.0", false},
		{"version above range", "7.1.0", "6.0.0", "7.0.0", false},

		// Invalid version format
		{"invalid version format", "invalid", "6.0.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.SERVER_CONFIGS.ServerVersion = tt.version

			if tt.minVersion != "" {
				utils.EntityMinVersionRequirements = map[utils.ResourceType]string{testResource: tt.minVersion}
			} else {
				utils.EntityMinVersionRequirements = map[utils.ResourceType]string{}
			}

			if tt.maxVersion != "" {
				utils.EntityMaxSupportedVersion = map[utils.ResourceType]string{testResource: tt.maxVersion}
			} else {
				utils.EntityMaxSupportedVersion = map[utils.ResourceType]string{}
			}

			result := utils.IsEntitySupportedInVersion(testResource)

			if result != tt.expected {
				t.Errorf("IsEntitySupportedInVersion(%s) with version=%s min=%s max=%s = %v; expected %v",
					testResource, tt.version, tt.minVersion, tt.maxVersion, result, tt.expected)
			}
		})
	}
}
