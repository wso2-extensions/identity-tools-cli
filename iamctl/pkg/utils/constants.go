/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package utils

// Resource type configs
const APPLICATIONS_CONFIG = "APPLICATIONS"
const IDP_CONFIG = "IDENTITY_PROVIDERS"
const USERSTORES_CONFIG = "USERSTORES"

// Tool configs
const KEYWORD_MAPPINGS_CONFIG = "KEYWORD_MAPPINGS"
const EXCLUDE_CONFIG = "EXCLUDE"
const INCLUDE_ONLY_CONFIG = "INCLUDE_ONLY"
const EXCLUDE_SECRETS_CONFIG = "EXCLUDE_SECRETS"
const ALLOW_DELETE_CONFIG = "ALLOW_DELETE"

// Server configs
const SERVER_URL_CONFIG = "SERVER_URL"
const CLIENT_ID_CONFIG = "CLIENT_ID"
const CLIENT_SECRET_CONFIG = "CLIENT_SECRET"
const TENANT_DOMAIN_CONFIG = "TENANT_DOMAIN"
const TOKEN_CONFIG = "TOKEN"

// Resource types
const APPLICATIONS = "Applications"
const IDENTITY_PROVIDERS = "IdentityProviders"
const USERSTORES = "UserStores"

// Config file names
const SERVER_CONFIG_FILE = "serverConfig.json"
const TOOL_CONFIG_FILE = "toolConfig.json"

// Media types
const MEDIA_TYPE_JSON = "application/json"
const MEDIA_TYPE_XML = "application/xml"
const MEDIA_TYPE_YAML = "application/yaml"
const MEDIA_TYPE_FORM = "application/x-www-form-urlencoded"

const DEFAULT_TENANT_DOMAIN = "carbon.super"
const SENSITIVE_FIELD_MASK = "'********'"
const RESIDENT_IDP_NAME = "LOCAL"

// Error codes
var ErrorCodes = map[int]string{

	400: "Bad request. Provided parameters are not in correct format.",
	401: "Unauthorized access.\nPlease check your server configurations.",
	403: "Forbidden request.",
	404: "Resource not found for the given ID.",
	409: "A resource with the same name already exists.",
	500: "Internal server error.",
}

// Identifiers of array types
var arrayIdentifiers = map[string]string{

	"inboundAuthenticationRequestConfigs": "inboundAuthKey",
	"spProperties":                        "name",
	"authenticationSteps":                 "stepOrder",
	"localAuthenticatorConfigs":           "name",
	"federatedIdentityProviders":          "identityProviderName",
	"federatedAuthenticatorConfigs":       "name",
	"properties":                          "name",
	"subProperties":                       "name",
	"idpProperties":                       "name",
	"provisioningConnectorConfigs":        "name",
	"provisioningIdentityProviders":       "identityProviderName",
	"requestPathAuthenticatorConfigs":     "name",
	"roleMappings":                        "localRole.localRoleName",
	"claimMappings":                       "localClaim.claimUri",
	"idpClaims":                           "claimId",
	"provisioningProperties":              "name",
	"claimAttributeMappings":              "claimURI",
}
