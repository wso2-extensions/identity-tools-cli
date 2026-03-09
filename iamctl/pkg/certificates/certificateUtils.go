/**
* Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package certificates

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type certificate struct {
	Alias string `json:"alias"`
}

func getCertificateList() ([]certificate, error) {

	var list []certificate
	resp, err := utils.SendGetListRequest(utils.CERTIFICATES, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving certificate list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved certificate list. %w", err)
		}

		err = json.Unmarshal(body, &list)
		if err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved certificate list. %w", err)
		}

		return list, nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving certificate list. Status code: %d, Error: %s", statusCode, error)
	}
	return nil, fmt.Errorf("error while retrieving certificate list")
}

func getDeployedCertificateAliases() []string {

	certs, err := getCertificateList()
	if err != nil {
		return []string{}
	}

	var aliases []string
	for _, cert := range certs {
		aliases = append(aliases, cert.Alias)
	}
	return aliases
}

func getCertificateKeywordMapping(alias string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.CertificateConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(alias, utils.KEYWORD_CONFIGS.CertificateConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func isCertificateExists(alias string, existingCertList []certificate) bool {
	for _, cert := range existingCertList {
		if cert.Alias == alias {
			return true
		}
	}
	return false
}

func getEncodedCertificate(alias string) (map[string]interface{}, error) {

	body, err := utils.SendGetRequest(utils.CERTIFICATES, alias, utils.WithContentType(utils.MEDIA_TYPE_PKIX_CERT))
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"alias":       alias,
		"certificate": string(body),
	}, nil
}
