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
const OIDC_SCOPES_CONFIG = "OIDC_SCOPES"
const ROLES_CONFIG = "ROLES"
const CHALLENGE_QUESTIONS_CONFIG = "CHALLENGE_QUESTIONS"
const EMAIL_TEMPLATES_CONFIG = "EMAIL_TEMPLATES"
const SCRIPT_LIBRARIES_CONFIG = "SCRIPT_LIBRARIES"
const GOVERNANCE_CONNECTORS_CONFIG = "GOVERNANCE_CONNECTORS"

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
const TOOL_CONFIG_PATH = "TOOL_CONFIG_PATH"
const KEYWORD_CONFIG_PATH = "KEYWORD_CONFIG_PATH"
const TOKEN_CONFIG = "TOKEN"

// Resource types
type ResourceType string

const (
	APPLICATIONS          ResourceType = "Applications"
	IDENTITY_PROVIDERS    ResourceType = "IdentityProviders"
	CLAIMS                ResourceType = "Claims"
	USERSTORES            ResourceType = "UserStores"
	OIDC_SCOPES           ResourceType = "OidcScopes"
	ROLES                 ResourceType = "Roles"
	CHALLENGE_QUESTIONS   ResourceType = "ChallengeQuestions"
	EMAIL_TEMPLATES       ResourceType = "EmailTemplates"
	SCRIPT_LIBRARIES      ResourceType = "ScriptLibraries"
	GOVERNANCE_CONNECTORS ResourceType = "GovernanceConnectors"
)

// Config file names
const SERVER_CONFIG_FILE = "serverConfig.json"
const TOOL_CONFIG_FILE = "toolConfig.json"
const KEYWORD_CONFIG_FILE = "keywordConfig.json"

type Format string

const (
	FormatYAML Format = "yaml"
	FormatJSON Format = "json"
	FormatXML  Format = "xml"
)

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
const ADMIN = "admin"
const OAUTH2 = "oauth2"
const ALL_ITEMS = "all_items" // Wildcard to match all elements in an array

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

var challengeQuestionsArrayIdentifiers = map[string]string{

	"questions": "questionId",
}

type ResourceIdentifierMeta struct {
	IdentifierPath  string // Path to the ID field in the resource object
	UniqueValuePath string // Path to the unique identifier field
}

type ResourceReferenceMeta struct {
	ReferencedResourceType ResourceType // The resource type being referenced
	ReferencePaths         []string     // Paths to where the referenced resource's ID appears
}

// Maps resource types to their identifier metadata.
var RESOURCE_IDENTIFIER_METADATA = map[ResourceType]ResourceIdentifierMeta{}

// Maps resource types to the resources they reference.
var RESOURCE_REFERENCE_METADATA = map[ResourceType][]ResourceReferenceMeta{}

// Array field paths for each resource type
var oidcScopeArrayFields = []string{
	"claims",
}

var rolesArrayFields = []string{
	"permissions",
	"schemas",
}

var challengeQuestionsArrayFields = []string{
	"questions",
}

var governanceConnectorArrayFields = []string{
	"properties",
}

// XML root element tags for each resource type
const (
	XML_ROOT_OIDC_SCOPE           = "Scope"
	XML_ROOT_ROLE                 = "Role"
	XML_ROOT_CHALLENGE_QUESTION   = "ChallengeSet"
	XML_ROOT_EMAIL_TEMPLATE       = "EmailTemplate"
	XML_ROOT_SCRIPT_LIBRARY       = "ScriptLibrary"
	XML_ROOT_GOVERNANCE_CONNECTOR = "GovernanceConnector"
)
