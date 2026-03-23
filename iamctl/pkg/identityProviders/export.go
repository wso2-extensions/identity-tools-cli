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
	"mime"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all identity providers to the IdentityProviders folder.
	log.Println("Exporting identity providers...")
	exportFilePath = filepath.Join(exportFilePath, utils.IDENTITY_PROVIDERS.String())

	if utils.IsResourceTypeExcluded(utils.IDENTITY_PROVIDERS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedIdpNames := append(getDeployedIdpNames(), utils.RESIDENT_IDP_NAME)
			utils.RemoveDeletedLocalResources(exportFilePath, deployedIdpNames)
		}
	}

	excludeSecerts := utils.AreSecretsExcluded(utils.TOOL_CONFIGS.IdpConfigs)
	exportAPIExists := exportAPIExists()
	idps, err := getIdpList()
	if err != nil {
		log.Println("Error: when exporting identity providers.", err)
	} else {
		for _, idp := range idps {
			if !utils.IsResourceExcluded(idp.Name, utils.TOOL_CONFIGS.IdpConfigs) {
				log.Println("Exporting identity provider: ", idp.Name)

				var err error
				if exportAPIExists {
					err = exportIdp(idp.Id, exportFilePath, format, excludeSecerts)
				} else {
					err = exportIdpWithCRUD(idp.Id, idp.Name, exportFilePath, format)
				}
				if err != nil {
					utils.UpdateFailureSummary(utils.IDENTITY_PROVIDERS, idp.Name)
					log.Printf("Error while exporting identity providers: %s. %s", idp.Name, err)
				} else {
					utils.UpdateSuccessSummary(utils.IDENTITY_PROVIDERS, utils.EXPORT)
					log.Println("Identity provider exported successfully: ", idp.Name)
				}
			}
		}
	}
	if !utils.IsResourceExcluded(utils.RESIDENT_IDP_NAME, utils.TOOL_CONFIGS.IdpConfigs) && exportAPIExists {
		log.Println("Exporting Resident identity provider")
		err := exportIdp(utils.RESIDENT_IDP_NAME, exportFilePath, format, excludeSecerts)
		if err != nil {
			log.Printf("Error while exporting resident identity provider: %s", err)
		} else {
			log.Println("Resident identity provider exported successfully")
		}
	}
}

func exportIdp(idpId string, outputDirPath string, format string, excludeSecrets bool) error {

	var fileType string
	// TODO: Extend support for json and xml formats.
	switch format {
	case "json":
		fileType = utils.MEDIA_TYPE_JSON
	case "xml":
		fileType = utils.MEDIA_TYPE_XML
	default:
		fileType = utils.MEDIA_TYPE_YAML
	}

	resp, err := utils.SendExportRequest(idpId, fileType, utils.IDENTITY_PROVIDERS, excludeSecrets)
	defer resp.Body.Close()

	if err != nil {
		return fmt.Errorf("error while exporting the identity provider: %s", err)
	}
	var attachmentDetail = resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(attachmentDetail)
	if err != nil {
		return fmt.Errorf("error while parsing the content disposition header: %s", err)
	}

	fileName := params["filename"]
	exportedFileName := filepath.Join(outputDirPath, fileName)
	fileInfo := utils.GetFileInfo(exportedFileName)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response body when exporting IDP: %s. %s", fileName, err)
	}

	idpKeywordMapping := getIdpKeywordMapping(fileInfo.ResourceName)
	modifiedFile, err := utils.ProcessExportedContent(exportedFileName, body, idpKeywordMapping, utils.IDENTITY_PROVIDERS)
	if err != nil {
		return fmt.Errorf("error while processing the exported content: %s", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing the exported content to file: %w", err)
	}
	return nil
}

func exportIdpWithCRUD(idpId, idpName, outputDirPath, formatString string) error {

	idpMap, err := getIdp(idpId)
	if err != nil {
		return fmt.Errorf("error while getting IDP: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, idpName, format)

	idpKeywordMapping := getIdpKeywordMapping(idpName)
	modifiedIdp, err := utils.ProcessExportedData(idpMap, exportedFileName, format, idpKeywordMapping, utils.IDENTITY_PROVIDERS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedIdp, format, utils.IDENTITY_PROVIDERS)
	if err != nil {
		return fmt.Errorf("error while serializing IDP: %w", err)
	}

	err = os.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}

func getIdp(idpId string) (map[string]interface{}, error) {

	body, err := utils.SendGetRequest(utils.IDENTITY_PROVIDERS, idpId)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving IDP: %w", err)
	}

	var idpStruct idpConfig
	if err := json.Unmarshal(body, &idpStruct); err != nil {
		return nil, fmt.Errorf("error unmarshalling IDP response: %w", err)
	}
	var idpMap map[string]interface{}
	if err := json.Unmarshal(body, &idpMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling IDP response to map: %w", err)
	}

	auths, err := getFederatedAuthenticators(idpId, idpStruct)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving federated authenticators for IDP:. %w", err)
	}
	if fedAuths, ok := idpMap["federatedAuthenticators"].(map[string]interface{}); ok {
		fedAuths["authenticators"] = auths
	}

	connectors, err := getOutboundConnectors(idpId, idpStruct)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving outbound connectors for IDP: %w", err)
	}
	if provisioning, ok := idpMap["provisioning"].(map[string]interface{}); ok {
		if outbound, ok := provisioning["outboundConnectors"].(map[string]interface{}); ok {
			outbound["connectors"] = connectors
		}
	}

	return idpMap, nil
}

func getFederatedAuthenticators(idpId string, idpStruct idpConfig) ([]interface{}, error) {

	auths := []interface{}{}
	if idpStruct.FederatedAuthenticators == nil {
		return auths, nil
	}

	for _, auth := range idpStruct.FederatedAuthenticators.Authenticators {
		authMap, ok := auth.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for federated authenticator of IDP")
		}
		authId, ok := authMap["authenticatorId"].(string)
		if !ok {
			return nil, fmt.Errorf("id not found for federated authenticator of IDP")
		}

		auth, err := utils.GetResourceData(utils.IDENTITY_PROVIDERS, idpId+"/federated-authenticators/"+authId)
		if err != nil {
			return nil, fmt.Errorf("error while retrieving federated authenticator %s: %w", authId, err)
		}
		auths = append(auths, auth)
	}

	return auths, nil
}

func getOutboundConnectors(idpId string, idpStruct idpConfig) ([]interface{}, error) {

	connectors := []interface{}{}
	if idpStruct.Provisioning == nil || idpStruct.Provisioning.OutboundConnectors == nil {
		return connectors, nil
	}

	for _, conn := range idpStruct.Provisioning.OutboundConnectors.Connectors {
		connMap, ok := conn.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected format for outbound connector of IDP")
		}
		connId, ok := connMap["connectorId"].(string)
		if !ok {
			return nil, fmt.Errorf("id not found for outbound connector of IDP")
		}

		connector, err := utils.GetResourceData(utils.IDENTITY_PROVIDERS, idpId+"/provisioning/outbound-connectors/"+connId)
		if err != nil {
			return nil, fmt.Errorf("error while retrieving outbound connector %s: %w", connId, err)
		}
		connectors = append(connectors, connector)
	}

	return connectors, nil
}
