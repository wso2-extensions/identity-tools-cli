package internal

const APP_NAME string = "IAMCTL"

// server names
const (
	ASGARDEO string = "Asgardeo"
	IS       string = "Identity Server"
)

// commands
const (
	ROOT_COMMAND  string = "iamctl"
	LOGIN_COMMAND string = "login"
)

// URL Components
const ASGARDEO_URL_PREFIX string = "https://api.asgardeo.io/t/"

const (
	AUTH_TOKEN_ENDPOINT string = "/oauth2/token"
)

// Auth grant type
const AUTH_GRANT_TYPE string = "client_credentials"

// scopes needed for the functionality
const REQUIRED_SCOPES string = "internal_userstore_create internal_userstore_view internal_userstore_update internal_userstore_delete internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update internal_application_mgt_create internal_application_mgt_delete internal_application_mgt_update internal_application_mgt_view internal_group_mgt_delete internal_group_mgt_create internal_group_mgt_update internal_group_mgt_view internal_idp_update internal_idp_view internal_idp_delete internal_idp_create internal_application_script_update internal_role_mgt_users_update internal_role_mgt_meta_update internal_role_mgt_view internal_role_mgt_groups_update internal_role_mgt_meta_create internal_role_mgt_delete"

// Keyring keys
const (
	CLIENT_ID_KEY     string = "client_id"
	CLIENT_SECRET_KEY string = "client_secret"
	ORG_NAME_KEY      string = "org_name"
	SERVER_URL_KEY    string = "server_url"

	ACCESS_TOKEN_KEY string = "access_token"
)

// Constants from old iamctl tool
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
const TOOL_CONFIG_PATH = "TOOL_CONFIG_PATH"
const KEYWORD_CONFIG_PATH = "KEYWORD_CONFIG_PATH"
const TOKEN_CONFIG = "TOKEN"

// Resource types
const APPLICATIONS = "Applications"
const IDENTITY_PROVIDERS = "IdentityProviders"
const CLAIMS = "Claims"
const USERSTORES = "UserStores"

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
