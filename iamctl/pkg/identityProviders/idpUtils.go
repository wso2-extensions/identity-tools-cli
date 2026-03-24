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
	"io/ioutil"
	"log"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type identityProvider struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type idpList struct {
	IdpCount          int                `json:"totalResults"`
	IdentityProviders []identityProvider `json:"identityProviders"`
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
	Claims      interface{} `json:"claims" yaml:"claims"`
	Roles       interface{} `json:"roles" yaml:"roles"`
	Certificate *struct {
		Certificates []string `json:"certificates" yaml:"certificates"`
		JwksUri      string   `json:"jwksUri" yaml:"jwksUri"`
	} `json:"certificate" yaml:"certificate"`
}

func getIdpList() ([]identityProvider, error) {

	idpCount, err := getTotalIdpCount()
	if err != nil {
		log.Println("Error: when retrieving IDP count. Retrieving only the default count.", err)
	}
	var list idpList
	resp, err := utils.SendGetListRequest(utils.IDENTITY_PROVIDERS, idpCount)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve available IDP list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved IDP list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved IDP list. %w", err)
		}
		resp.Body.Close()

		return list.IdentityProviders, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving IDP list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("error while retrieving identity provider list")
}

func getTotalIdpCount() (count int, err error) {

	var list idpList
	resp, err := utils.SendGetListRequest(utils.IDENTITY_PROVIDERS, -1)
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve available IDP list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return -1, fmt.Errorf("error when reading the retrieved IDP list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return -1, fmt.Errorf("error when unmarshalling the retrieved IDP list. %w", err)
		}
		resp.Body.Close()

		return list.IdpCount, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return -1, fmt.Errorf("error while retrieving IDP count. Status code: %d, Error: %s", statusCode, error)
	}
	return -1, fmt.Errorf("error while retrieving identity provider count")
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

func processFederatedAuthenticators(idpId string, idpStruct idpConfig, idpMap map[string]interface{}) error {

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
		fullAuth, err := utils.GetResourceData(utils.IDENTITY_PROVIDERS, idpId+"/federated-authenticators/"+authId)
		if err != nil {
			return fmt.Errorf("error while retrieving federated authenticator %s: %w", authId, err)
		}
		auths = append(auths, fullAuth)
	}
	fedAuths["authenticators"] = auths

	if idpStruct.FederatedAuthenticators.DefaultAuthenticatorId == "" {
		fedAuths["defaultAuthenticatorId"] = ""
	}
	return nil
}

func processOutboundConnectors(idpId string, idpStruct idpConfig, idpMap map[string]interface{}) error {

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
		fullConn, err := utils.GetResourceData(utils.IDENTITY_PROVIDERS, idpId+"/provisioning/outbound-connectors/"+connId)
		if err != nil {
			return fmt.Errorf("error while retrieving outbound connector %s: %w", connId, err)
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

	claims, ok := idpMap["claims"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid format for claims")
	}

	for _, claimKey := range []string{"userIdClaim", "roleClaim"} {
		claim, ok := claims[claimKey].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid format for claim: %s", claimKey)
		}
		if len(claim) == 0 {
			claims[claimKey] = map[string]interface{}{"uri": ""}
		}
	}
	return nil
}

func preprocessIdpKeys(data interface{}) (interface{}, error) {

	if utils.ExportAPIExists(utils.IDENTITY_PROVIDERS) {
		return data, nil
	}

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

func patchIdp(idpId string, patchOps []map[string]interface{}) error {

	body, err := json.Marshal(patchOps)
	if err != nil {
		return fmt.Errorf("error marshalling patch operations for identity provider: %w", err)
	}

	resp, err := utils.SendPatchRequest(utils.IDENTITY_PROVIDERS, idpId, body)
	if err != nil {
		return fmt.Errorf("error patching identity provider: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func init() {

	utils.DataPreprocessFuncs[utils.IDENTITY_PROVIDERS] = preprocessIdpKeys
}
