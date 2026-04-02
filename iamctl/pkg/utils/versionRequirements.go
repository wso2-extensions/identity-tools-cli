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

// Minimum WSO2 Identity Server version requirements for each resource type.
// If a resource type is not present in this map, there is no minimum version requirement.
const (
	MIN_VERSION_APPLICATIONS = "6.1.0"
)

var EntityMinVersionRequirements = map[ResourceType]string{
	APPLICATIONS: MIN_VERSION_APPLICATIONS,
}

// Maximum supported WSO2 Identity Server version for each resource type.
// If a resource type is not present in this map, there is no upper version limit.
const ()

var EntityMaxSupportedVersion = map[ResourceType]string{}

// Minimum WSO2 Identity Server version requirements for resource-specific APIs
const (
	MIN_VERSION_USERSTORE_EXPORT_API = "6.1.0"
	MIN_VERSION_CLAIMS_EXPORT_API    = "6.1.0"
	MIN_VERSION_IDP_EXPORT_API       = "6.1.0"
)

var ExportAPIMinVersionRequirements = map[ResourceType]string{
	USERSTORES:         MIN_VERSION_USERSTORE_EXPORT_API,
	CLAIMS:             MIN_VERSION_CLAIMS_EXPORT_API,
	IDENTITY_PROVIDERS: MIN_VERSION_IDP_EXPORT_API,
}
