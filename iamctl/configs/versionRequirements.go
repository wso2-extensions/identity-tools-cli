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

package configs

// Minimum WSO2 Identity Server version requirements for each resource type.
// Update these values based on when each API was introduced in WSO2 IS.
const (
	MIN_VERSION_APPLICATIONS       = "5.9.0"
	MIN_VERSION_IDENTITY_PROVIDERS = "5.9.0"
	MIN_VERSION_CLAIMS             = "5.9.0"
	MIN_VERSION_USERSTORES         = "5.9.0"
)

var EntityVersionRequirements = map[string]string{
	APPLICATIONS:       MIN_VERSION_APPLICATIONS,
	IDENTITY_PROVIDERS: MIN_VERSION_IDENTITY_PROVIDERS,
	CLAIMS:             MIN_VERSION_CLAIMS,
	USERSTORES:         MIN_VERSION_USERSTORES,
}
