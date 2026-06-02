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

package identityproviders

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/claims"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type identityProvider struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type idpConfig struct {
	Id                      string `json:"id" yaml:"id"`
	Name                    string `json:"name" yaml:"name"`
	FederatedAuthenticators *struct {
		DefaultAuthenticatorId string        `json:"defaultAuthenticatorId" yaml:"defaultAuthenticatorId"`
		Authenticators         []interface{} `json:"authenticators" yaml:"authenticators"`
	} `json:"federatedAuthenticators" yaml:"federatedAuthenticators"`
	Provisioning *struct {
		Jit                interface{} `json:"jit" yaml:"jit"`
		OutboundConnectors *struct {
			DefaultConnectorId string        `json:"defaultConnectorId" yaml:"defaultConnectorId"`
			Connectors         []interface{} `json:"connectors" yaml:"connectors"`
		} `json:"outboundConnectors" yaml:"outboundConnectors"`
	} `json:"provisioning" yaml:"provisioning"`
	Claims              interface{} `json:"claims" yaml:"claims"`
	Roles               interface{} `json:"roles" yaml:"roles"`
	Groups              interface{} `json:"groups" yaml:"groups"`
	ImplicitAssociation interface{} `json:"implicitAssociation" yaml:"implicitAssociation"`
	Certificate         *struct {
		Certificates []string `json:"certificates" yaml:"certificates"`
		JwksUri      string   `json:"jwksUri" yaml:"jwksUri"`
	} `json:"certificate" yaml:"certificate"`
}

type resourceMeta struct {
	Properties []struct {
		Key            string `json:"key"`
		IsConfidential bool   `json:"isConfidential"`
	} `json:"properties"`
}

var idpPatchSkipKeys = map[string]bool{
	"id":                      true,
	"name":                    true,
	"certificate":             true,
	"federatedAuthenticators": true,
	"provisioning":            true,
	"claims":                  true,
	"roles":                   true,
	"groups":                  true,
	"implicitAssociation":     true,
	"templateId":              true,
}

var customAuthSecretsWarningLogged bool

func getIdpList() ([]identityProvider, error) {

	data, err := utils.SendPaginatedGetListRequest(
		utils.IDENTITY_PROVIDERS,
		"totalResults",
		"count",
		"offset",
		"limit",
		"identityProviders",
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving IDP list. %w", err)
	}
	var idps []identityProvider
	if err := json.Unmarshal(data, &idps); err != nil {
		return nil, fmt.Errorf("error when unmarshalling IDP list. %w", err)
	}
	return idps, nil
}

func getDeployedIdpNames() []string {

	idps, err := getIdpList()
	if err != nil {
		return []string{}
	}

	var idpNames []string
	for _, idp := range idps {
		idpNames = append(idpNames, idp.Name)
	}
	return idpNames
}

func getIdpKeywordMapping(idpName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.IdpConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(idpName, utils.KEYWORD_CONFIGS.IdpConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func getIdpId(idpName string, existingIdpList []identityProvider) string {

	for _, idp := range existingIdpList {
		if idp.Name == idpName {
			return idp.Id
		}
	}
	return ""
}

func processFederatedAuthenticators(idpId string, idpStruct idpConfig, idpMap map[string]interface{}, excludeSecrets bool) error {

	fedAuths, ok := idpMap["federatedAuthenticators"].(map[string]interface{})
	if !ok || idpStruct.FederatedAuthenticators == nil {
		return fmt.Errorf("invalid format for federated authenticators")
	}

	auths := []interface{}{}
	for _, auth := range idpStruct.FederatedAuthenticators.Authenticators {
		authMap, ok := auth.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for federated authenticator")
		}
		authId, ok := authMap["authenticatorId"].(string)
		if !ok {
			return fmt.Errorf("id not found for federated authenticator")
		}
		isEnabled, ok := authMap["isEnabled"].(bool)
		if !ok {
			return fmt.Errorf("isEnabled flag not found for federated authenticator: %s", authId)
		}
		if !isEnabled {
			continue
		}

		fullAuth, err := utils.GetResourceData(utils.IDENTITY_PROVIDERS, idpId+"/federated-authenticators/"+authId)
		if err != nil {
			return fmt.Errorf("error while retrieving federated authenticator %s: %w", authId, err)
		}
		fullAuthMap, ok := fullAuth.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for retrieved federated authenticator: %s", authId)
		}

		definedByRaw, exists := fullAuthMap["definedBy"]
		definedBy, ok := definedByRaw.(string)
		if exists && !ok {
			return fmt.Errorf("unexpected format for definedBy field of federated authenticator: %s", authId)
		}
		if definedBy == "USER" {
			if !excludeSecrets && !customAuthSecretsWarningLogged {
				utils.PrintLog(utils.LogLevelWarn, utils.IDENTITY_PROVIDERS, "", "Secrets exclusion cannot be disabled for custom authenticators(service-based). All secrets will be masked.")
				customAuthSecretsWarningLogged = true
			}
			if err := processEndpointAuthProperties(fullAuthMap); err != nil {
				return fmt.Errorf("error processing endpoint auth properties for authenticator %s: %v", authId, err)
			}
		} else if excludeSecrets {
			if err := maskSecretProperties(fullAuthMap, "meta/federated-authenticators/"+authId); err != nil {
				return fmt.Errorf("error masking secrets for authenticator %s: %v", authId, err)
			}
		}

		auths = append(auths, fullAuthMap)
	}
	fedAuths["authenticators"] = auths

	if idpStruct.FederatedAuthenticators.DefaultAuthenticatorId == "" {
		fedAuths["defaultAuthenticatorId"] = ""
	}
	return nil
}

func processOutboundConnectors(idpId string, idpStruct idpConfig, idpMap map[string]interface{}, excludeSecrets bool) error {

	provisioning, ok := idpMap["provisioning"].(map[string]interface{})
	if !ok || idpStruct.Provisioning == nil || idpStruct.Provisioning.OutboundConnectors == nil {
		return fmt.Errorf("invalid format for provisioning")
	}
	outbound, ok := provisioning["outboundConnectors"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid format for outbound connectors")
	}

	connectors := []interface{}{}
	for _, conn := range idpStruct.Provisioning.OutboundConnectors.Connectors {
		connMap, ok := conn.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for outbound connector")
		}
		connId, ok := connMap["connectorId"].(string)
		if !ok {
			return fmt.Errorf("id not found for outbound connector")
		}
		isEnabled, ok := connMap["isEnabled"].(bool)
		if !ok {
			return fmt.Errorf("isEnabled flag not found for outbound connector: %s", connId)
		}
		if !isEnabled {
			continue
		}

		fullConn, err := utils.GetResourceData(utils.IDENTITY_PROVIDERS, idpId+"/provisioning/outbound-connectors/"+connId)
		if err != nil {
			return fmt.Errorf("error while retrieving outbound connector %s: %w", connId, err)
		}
		if excludeSecrets {
			fullConnMap, ok := fullConn.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected format for retrieved outbound connector: %s", connId)
			}
			if err := maskSecretProperties(fullConnMap, "meta/outbound-provisioning-connectors/"+connId); err != nil {
				return fmt.Errorf("error masking secrets for connector %s: %v", connId, err)
			}
		}

		connectors = append(connectors, fullConn)
	}
	outbound["connectors"] = connectors

	if idpStruct.Provisioning.OutboundConnectors.DefaultConnectorId == "" {
		outbound["defaultConnectorId"] = ""
	}
	return nil
}

func processClaims(idpMap map[string]interface{}) error {

	claimMap, ok := idpMap["claims"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid format for claims")
	}

	for _, claimKey := range []string{"userIdClaim", "roleClaim"} {
		claim, ok := claimMap[claimKey].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid format for claim: %s", claimKey)
		}
		if len(claim) == 0 {
			claimMap[claimKey] = map[string]interface{}{"uri": ""}
			continue
		}

		if claimKey == "roleClaim" {
			roleClaimUri, ok := claim["uri"].(string)
			if !ok {
				return fmt.Errorf("invalid format for roleClaim uri")
			}
			if claims.RoleClaimUnsupported() && roleClaimUri == "http://wso2.org/claims/role" {
				claim["uri"] = ""
			}
		}
	}
	return nil
}

func processGroups(idpMap map[string]interface{}) error {

	raw, exists := idpMap["groups"]
	if !exists {
		return nil
	}
	groups, ok := raw.([]interface{})
	if !ok {
		return fmt.Errorf("invalid format for groups")
	}
	for _, grp := range groups {
		group, ok := grp.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid format for group")
		}
		delete(group, "id")
	}
	return nil
}

func processImplicitAssociation(idpMap map[string]interface{}) error {

	raw, exists := idpMap["implicitAssociation"]
	if !exists {
		return nil
	}
	assoc, ok := raw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid format for implicit association")
	}

	attrs, ok := assoc["lookupAttribute"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid format for lookupAttribute in implicit association")
	}
	if len(attrs) == 0 {
		assoc["lookupAttribute"] = append(attrs, "")
	}
	return nil
}

func processEndpointAuthProperties(authenticatorMap map[string]interface{}) error {

	endpoint, ok := authenticatorMap["endpoint"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for endpoint")
	}
	auth, ok := endpoint["authentication"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for endpoint authentication")
	}
	authType, ok := auth["type"].(string)
	if !ok {
		return fmt.Errorf("unexpected format for endpoint authentication type")
	}

	var props map[string]interface{}
	if rawProps, exists := auth["properties"]; !exists {
		props = map[string]interface{}{}
	} else {
		props, ok = rawProps.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for authentication properties")
		}
	}

	switch authType {
	case "NONE":
	case "BASIC":
		if _, exists := props["username"]; !exists {
			props["username"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		}
		props["password"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "BEARER":
		props["accessToken"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "API_KEY":
		if _, exists := props["header"]; !exists {
			props["header"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		}
		props["value"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "CLIENT_CREDENTIAL":
		props["clientSecret"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	case "PASSWORD_CREDENTIAL":
		props["clientSecret"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		props["password"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
	default:
		return fmt.Errorf("unknown endpoint authentication type: %s", authType)
	}

	auth["properties"] = props

	if _, exists := endpoint["allowedHeaders"]; !exists {
		endpoint["allowedHeaders"] = []interface{}{}
	}
	if _, exists := endpoint["allowedParameters"]; !exists {
		endpoint["allowedParameters"] = []interface{}{}
	}
	return nil
}

func maskSecretProperties(resourceMap map[string]interface{}, metaPath string) error {

	body, err := utils.SendGetRequest(utils.IDENTITY_PROVIDERS, metaPath)
	if err != nil {
		return fmt.Errorf("error fetching metadata from %s: %w", metaPath, err)
	}
	var meta resourceMeta
	if err := json.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("error parsing metadata from %s: %w", metaPath, err)
	}

	confidentialKeys := map[string]bool{}
	for _, prop := range meta.Properties {
		if prop.IsConfidential {
			confidentialKeys[prop.Key] = true
		}
	}
	properties, ok := resourceMap["properties"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected format for properties")
	}

	for _, p := range properties {
		propMap, ok := p.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for property")
		}
		key, ok := propMap["key"].(string)
		if !ok {
			return fmt.Errorf("unexpected format for property key")
		}
		if confidentialKeys[key] {
			propMap["value"] = utils.SENSITIVE_FIELD_MASK_WITHOUT_QUOTES
		}
	}
	return nil
}

func preprocessIdpKeys(data interface{}) (interface{}, error) {

	data = utils.ConvertToStringKeyMap(data)
	d, ok := data.(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for IDP data")
	}

	if claims, ok := d["claims"].(map[string]interface{}); ok {
		if v, exists := claims["mappings"]; exists {
			claims["claimMappings"] = v
			delete(claims, "mappings")
		}
	}
	if roles, ok := d["roles"].(map[string]interface{}); ok {
		if v, exists := roles["mappings"]; exists {
			roles["roleMappings"] = v
			delete(roles, "mappings")
		}
	}

	return data, nil
}

func postprocessIdpKeys(data interface{}) (interface{}, error) {

	d, ok := data.(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("invalid format for IDP data")
	}

	if claims, ok := d["claims"].(map[string]interface{}); ok {
		if v, exists := claims["claimMappings"]; exists {
			claims["mappings"] = v
			delete(claims, "claimMappings")
		}
	}
	if roles, ok := d["roles"].(map[string]interface{}); ok {
		if v, exists := roles["roleMappings"]; exists {
			roles["mappings"] = v
			delete(roles, "roleMappings")
		}
	}

	return data, nil
}

func createPostRequestBody(idpMap map[string]interface{}) ([]byte, error) {

	delete(idpMap, "id")
	delete(idpMap, "isEnabled")
	body, err := json.Marshal(idpMap)
	if err != nil {
		return nil, fmt.Errorf("error marshalling identity provider: %w", err)
	}
	return body, nil
}

func buildIdpPatchOps(idpMap map[string]interface{}) []map[string]interface{} {

	patchOps := []map[string]interface{}{}
	for key, value := range idpMap {
		if idpPatchSkipKeys[key] {
			continue
		}
		patchOps = append(patchOps, map[string]interface{}{
			"operation": "REPLACE",
			"path":      "/" + key,
			"value":     value,
		})
	}
	return patchOps
}

func buildCertificatePatchOps(idpId string, localIdpStruct idpConfig) ([]map[string]interface{}, error) {

	body, err := utils.SendGetRequest(utils.IDENTITY_PROVIDERS, idpId)
	if err != nil {
		return nil, fmt.Errorf("error fetching deployed identity provider: %w", err)
	}
	var deployedIdp idpConfig
	if err := json.Unmarshal(body, &deployedIdp); err != nil {
		return nil, fmt.Errorf("error parsing deployed identity provider: %w", err)
	}

	deployedCert := deployedIdp.Certificate
	localCert := localIdpStruct.Certificate
	if deployedCert == nil && localCert == nil {
		return nil, nil
	}

	deployedHasJwks := deployedCert != nil && deployedCert.JwksUri != ""
	localHasJwks := localCert != nil && localCert.JwksUri != ""

	localCertSet := make(map[string]bool)
	if localCert != nil {
		for _, c := range localCert.Certificates {
			localCertSet[c] = true
		}
	}
	deployedCertSet := make(map[string]bool)
	if deployedCert != nil {
		for _, c := range deployedCert.Certificates {
			deployedCertSet[c] = true
		}
	}

	var patchOps []map[string]interface{}

	// JwksUri: removal (remove before add; mutually exclusive with certificates)
	if deployedHasJwks && !localHasJwks {
		patchOps = append(patchOps, map[string]interface{}{
			"operation": "REMOVE",
			"path":      "/certificate/jwksUri",
		})
	}

	// Certificates: removals (descending index to keep remaining indexes stable)
	if deployedCert != nil {
		for i := len(deployedCert.Certificates) - 1; i >= 0; i-- {
			if !localCertSet[deployedCert.Certificates[i]] {
				patchOps = append(patchOps, map[string]interface{}{
					"operation": "REMOVE",
					"path":      fmt.Sprintf("/certificate/certificates/%d", i),
					"value":     nil,
				})
			}
		}
	}

	// JwksUri: addition or replacement
	if localHasJwks {
		op := "ADD"
		if deployedHasJwks {
			op = "REPLACE"
		}
		patchOps = append(patchOps, map[string]interface{}{
			"operation": op,
			"path":      "/certificate/jwksUri",
			"value":     localCert.JwksUri,
		})
	}

	// Certificates: additions (appended after the surviving certs)
	if localCert != nil {
		keptCount := 0
		for c := range deployedCertSet {
			if localCertSet[c] {
				keptCount++
			}
		}
		addIndex := keptCount
		for _, c := range localCert.Certificates {
			if !deployedCertSet[c] {
				patchOps = append(patchOps, map[string]interface{}{
					"operation": "ADD",
					"path":      fmt.Sprintf("/certificate/certificates/%d", addIndex),
					"value":     c,
				})
				addIndex++
			}
		}
	}
	return patchOps, nil
}

func patchIdp(idpId string, patchOps []map[string]interface{}) error {

	body, err := json.Marshal(patchOps)
	if err != nil {
		return fmt.Errorf("error marshalling patch operations for identity provider: %w", err)
	}

	resp, err := utils.SendPatchRequest(utils.IDENTITY_PROVIDERS, idpId, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func removeOutboundProvisioningRoles(idpMap map[string]interface{}) error {

	if shouldRemoveOutboundProvisioningRoles() {
		roles, ok := idpMap["roles"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected format for roles")
		}
		delete(roles, "outboundProvisioningRoles")
	}
	return nil
}

func removeProvisioningRole(fileContent []byte) []byte {

	if !shouldRemoveOutboundProvisioningRoles() {
		return fileContent
	}
	yamlPattern := regexp.MustCompile(`(?m)(^\s*provisioningRole:)[^\n]*`)
	result := yamlPattern.ReplaceAllString(string(fileContent), `${1} ""`)

	jsonPattern := regexp.MustCompile(`("provisioningRole"\s*:\s*)"[^"]*"`)
	result = jsonPattern.ReplaceAllString(result, `${1}""`)

	return []byte(result)
}

func processIdpGroupFields(fileContent []byte) []byte {

	idpGroupId := regexp.MustCompile(`(?m)^\s+-?\s*idpGroupId:[^\n]*\n`)
	result := idpGroupId.ReplaceAllString(string(fileContent), "")

	idpId := regexp.MustCompile(`(?m)^\s+-?\s*idpId:[^\n]*\n`)
	result = idpId.ReplaceAllString(result, "")

	subField := regexp.MustCompile(`(?m)^( +)  (idpGroupName:)`)
	result = subField.ReplaceAllString(result, "${1}- ${2}")

	return []byte(result)
}

func init() {

	utils.DataPreprocessFuncs[utils.IDENTITY_PROVIDERS] = preprocessIdpKeys
}

func shouldRemoveOutboundProvisioningRoles() bool {

	if utils.SERVER_CONFIGS.ServerVersion == "" {
		return true
	}
	cmp, err := utils.CompareVersions(utils.SERVER_CONFIGS.ServerVersion, utils.MIN_VERSION_OUTBOUND_PROV_GROUPS)
	// Consider outbound provisioning groups exist when the server version is "" (Asgardeo)
	return err != nil || cmp >= 0
}
