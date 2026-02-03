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

package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
)

// Checks if a resource type is supported in the configured WSO2 IS version.
// Returns true if:
//   - Version is not configured (backward compatibility - version checking disabled)
//   - Configured version >= minimum required version for the resource type
//   - No version requirement is defined for the resource type
//   - Configured version format is invalid
func IsEntitySupportedInVersion(resourceType string) bool {

	if SERVER_CONFIGS.ServerVersion == "" {
		return true
	}
	minVersion, exists := configs.EntityVersionRequirements[resourceType]
	if !exists {
		return true
	}

	comparison, err := CompareVersions(SERVER_CONFIGS.ServerVersion, minVersion)
	if err != nil {
		log.Printf("Warning: Invalid version format. Configured: %s",
			SERVER_CONFIGS.ServerVersion)
		return true
	}

	isSupported := comparison >= 0
	if !isSupported {
		log.Printf("Skipping %s: Supported from IS version %s or higher",
			resourceType, minVersion)
	}

	return isSupported
}

// CompareVersions compares two semantic version strings
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {

	parts1, err := ParseVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version format for v1 (%s): %w", v1, err)
	}

	parts2, err := ParseVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version format for v2 (%s): %w", v2, err)
	}

	for i := 0; i < 3; i++ {
		if parts1[i] < parts2[i] {
			return -1, nil
		}
		if parts1[i] > parts2[i] {
			return 1, nil
		}
	}

	return 0, nil
}

func ParseVersion(version string) ([3]int, error) {

	var parts [3]int
	components := strings.Split(strings.TrimSpace(version), ".")
	if len(components) < 2 || len(components) > 3 {
		return parts, fmt.Errorf("version must have format MAJOR.MINOR or MAJOR.MINOR.PATCH, got: %s", version)
	}

	for i, component := range components {
		num, err := strconv.Atoi(strings.TrimSpace(component))
		if err != nil {
			return parts, fmt.Errorf("invalid version component '%s': must be an integer", component)
		}
		if num < 0 {
			return parts, fmt.Errorf("version component cannot be negative: %d", num)
		}
		parts[i] = num
	}

	return parts, nil
}
