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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type identityProvider struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type idpList struct {
	IdentityProviders []identityProvider `json:"identityProviders"`
}

type idpConfig struct {
	IdentityProviderName string `yaml:"identityProviderName"`
	IdentityProviderId   string
}

func getIdpList() ([]identityProvider, error) {

	var reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/identity-providers"
	var list idpList

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, _ := http.NewRequest("GET", reqUrl, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	req.Header.Set("accept", "*/*")
	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve available IDP list. %w", err)
	}

	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrived IDP list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrived IDP list. %w", err)
		}
		resp.Body.Close()

		return list.IdentityProviders, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving IDP list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("error while retrieving identity provider list")
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

	if utils.TOOL_CONFIGS.IdpConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(idpName, utils.TOOL_CONFIGS.IdpConfigs)
	}
	return utils.TOOL_CONFIGS.KeywordMappings
}
