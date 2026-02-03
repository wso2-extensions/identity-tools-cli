/**
* Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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
const CLAIM_CONFIG = "CLAIMS"
const USERSTORES_CONFIG = "USERSTORES"

// Tool configs
const EXCLUDE_CONFIG = "EXCLUDE"
const INCLUDE_ONLY_CONFIG = "INCLUDE_ONLY"
const EXCLUDE_SECRETS_CONFIG = "EXCLUDE_SECRETS"
const ALLOW_DELETE_CONFIG = "ALLOW_DELETE"

// Keyword configs
const KEYWORD_MAPPINGS_CONFIG = "KEYWORD_MAPPINGS"

// Server configs
const SERVER_URL_CONFIG = "SERVER_URL"
const CLIENT_ID_CONFIG = "CLIENT_ID"
const CLIENT_SECRET_CONFIG = "CLIENT_SECRET"
const TENANT_DOMAIN_CONFIG = "TENANT_DOMAIN"
const ORGANIZATION_CONFIG = "ORGANIZATION"
const SERVER_VERSION_CONFIG = "SERVER_VERSION"
const TOOL_CONFIG_PATH = "TOOL_CONFIG_PATH"
const KEYWORD_CONFIG_PATH = "KEYWORD_CONFIG_PATH"
const TOKEN_CONFIG = "TOKEN"

// Config file names
const SERVER_CONFIG_FILE = "serverConfig.json"
const TOOL_CONFIG_FILE = "toolConfig.json"
const KEYWORD_CONFIG_FILE = "keywordConfig.json"

// Media types
const MEDIA_TYPE_JSON = "application/json"
const MEDIA_TYPE_XML = "application/xml"
const MEDIA_TYPE_YAML = "application/yaml"
const MEDIA_TYPE_FORM = "application/x-www-form-urlencoded"

const DEFAULT_TENANT_DOMAIN = "carbon.super"
const SENSITIVE_FIELD_MASK = "'********'"
const RESIDENT_IDP_NAME = "LOCAL"
const CONSOLE = "Console"
const MY_ACCOUNT = "My Account"
const OAUTH2 = "oauth2"

// Error codes
var ErrorCodes = map[int]string{

	400: "Bad request. Provided parameters are not in correct format.",
	401: "Unauthorized access.\nPlease check your server configurations.",
	403: "Forbidden request.",
	404: "Resource not found for the given ID.",
	409: "A resource with the same name already exists.",
	500: "Internal server error.",
}

// Identifiers of array types for each resource type
var applicationArrayIdentifiers = map[string]string{

	"inboundAuthenticationRequestConfigs": "inboundAuthKey",
	"spProperties":                        "name",
	"authenticationSteps":                 "stepOrder",
	"localAuthenticatorConfigs":           "name",
	"federatedIdentityProviders":          "identityProviderName",
	"federatedAuthenticatorConfigs":       "name",
	"properties":                          "name",
	"provisioningConnectorConfigs":        "name",
	"provisioningIdentityProviders":       "identityProviderName",
	"requestPathAuthenticatorConfigs":     "name",
	"roleMappings":                        "localRole.localRoleName",
	"claimMappings":                       "localClaim.claimUri",
	"idpClaims":                           "claimId",
	"provisioningProperties":              "name",
	"applicationRoleMappingConfig":        "idPName",
}

var idpArrayIdentifiers = map[string]string{

	"claimMappings":                 "localClaim.claimUri",
	"idpClaims":                     "claimId",
	"properties":                    "name",
	"idpProperties":                 "name",
	"permissions":                   "value",
	"roleMappings":                  "localRole.localRoleName",
	"provisioningConnectorConfigs":  "name",
	"federatedAuthenticatorConfigs": "name",
	"provisioningProperties":        "name",
}

var userStoreArrayIdentifiers = map[string]string{

	"claimAttributeMappings": "claimURI",
	"properties":             "name",
}

var claimArrayIdentifiers = map[string]string{

	"properties":       "key",
	"attributeMapping": "mappedAttribute",
	"claims":           "id",
}
