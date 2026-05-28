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

package claims

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, "", "Importing claims...")
	importFilePath := filepath.Join(inputDirPath, utils.CLAIMS.String())

	if !utils.IsEntitySupportedInOrg(utils.CLAIMS) || utils.IsResourceTypeExcluded(utils.CLAIMS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, "", "No claim dialects to import.")
		return
	}

	existingClaimDialectList, err := getClaimDialectsList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CLAIMS, "", fmt.Sprintf("Error when retrieving the deployed claim dialect list: %s", err))
		return
	}

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CLAIMS, "", fmt.Sprintf("Error importing claim dialects: %s", err))
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedClaimdialect(files, existingClaimDialectList)
	}

	// Move the local claims file to the front of the array to import it first
	for i, file := range files {
		if strings.Contains(file.Name(), "http_wso2_org_claims.") {
			files[0], files[i] = files[i], files[0]
			break
		}
	}
	localClaimDialectSummary = LocalClaimDialectSummary{}

	for _, file := range files {
		claimFilePath := filepath.Join(importFilePath, file.Name())

		dialectUri, err := getDialectURIFromFile(claimFilePath)
		if err != nil {
			utils.PrintLog(utils.LogLevelError, utils.CLAIMS, "", fmt.Sprintf("Error reading dialect URI from file: %s", err))
			continue
		}
		dialectId := getClaimDialectId(dialectUri, existingClaimDialectList)

		if !utils.IsResourceExcluded(dialectUri, utils.TOOL_CONFIGS.ClaimConfigs) {
			err = importClaimDialect(dialectId, dialectUri, claimFilePath)
			if err != nil {
				utils.PrintLog(utils.LogLevelError, utils.CLAIMS, dialectUri, fmt.Sprintf("Error importing claim dialect: %s", err))
				utils.UpdateFailureSummary(utils.CLAIMS, dialectUri)
			}
		}
	}
	removeStaleClaimsFromLocalDialect()
}

func importClaimDialect(dialectId, dialectUri, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for claim dialect: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	claimKeywordMapping := getClaimKeywordMapping(dialectUri)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), claimKeywordMapping)

	// Min version requirement for claims export api is removed. CRUD apis used for all versions
	if utils.ExportAPIExists(utils.CLAIMS) {
		if dialectId == "" {
			return importDialect(dialectUri, importFilePath, modifiedFileData)
		}
		return updateDialect(dialectId, dialectUri, importFilePath, modifiedFileData)
	}

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for claim dialect: %w", err)
	}

	claims, err := parseClaims([]byte(modifiedFileData), format)
	if err != nil {
		return fmt.Errorf("error parsing claims from dialect file: %w", err)
	}

	if dialectId == "" {
		return importClaimDialectWithCRUD(dialectUri, claims)
	}
	return updateClaimDialectWithCRUD(dialectId, dialectUri, claims)
}

func importDialect(dialectUri, importFilePath, modifiedFileData string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectUri, "Creating new claim dialect")
	resp, err := utils.SendImportRequest(importFilePath, modifiedFileData, utils.CLAIMS)
	if err != nil {
		return fmt.Errorf("error when importing claim dialect: %s", err)
	}
	defer resp.Body.Close()
	utils.UpdateSuccessSummary(utils.CLAIMS, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectUri, "Imported successfully")
	return nil
}

func updateDialect(dialectId, dialectUri, importFilePath, modifiedFileData string) error {

	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectUri, "Updating claim dialect")
	err := utils.SendUpdateRequest(dialectId, importFilePath, modifiedFileData, utils.CLAIMS)
	if err != nil {
		return fmt.Errorf("error when updating claim dialect: %s", err)
	}
	utils.UpdateSuccessSummary(utils.CLAIMS, utils.UPDATE)
	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectUri, "Updated successfully")
	return nil
}

func importClaimDialectWithCRUD(dialectURI string, claims []map[string]interface{}) error {

	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectURI, "Creating new claim dialect")

	newDialectId, err := createDialect(dialectURI)
	if err != nil {
		return fmt.Errorf("error when creating claim dialect: %w", err)
	}
	for _, claim := range claims {
		if err := createClaim(newDialectId, claim); err != nil {
			return fmt.Errorf("error when creating claims of dialect %s: %w", dialectURI, err)
		}
	}

	utils.UpdateSuccessSummary(utils.CLAIMS, utils.IMPORT)
	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectURI, "Imported successfully")
	return nil
}

func updateClaimDialectWithCRUD(dialectId, dialectURI string, localClaims []map[string]interface{}) error {

	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectURI, "Updating claim dialect")

	deployedClaims, err := getClaimsList(dialectId)
	if err != nil {
		return fmt.Errorf("error retrieving deployed claims for dialect: %w", err)
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		if dialectId == utils.LOCAL_CLAIM_DIALECT {
			localClaimDialectSummary.DialectURI = dialectURI
			localClaimDialectSummary.LocalClaims = localClaims
			localClaimDialectSummary.DeployedClaims = deployedClaims
		} else {
			err := removeDeletedDeployedClaims(dialectId, deployedClaims, localClaims)
			if err != nil {
				return fmt.Errorf("error removing deleted claims of dialect: %w", err)
			}
		}
	}

	if err := updateChangedClaims(dialectId, localClaims, deployedClaims); err != nil {
		return fmt.Errorf("error updating changed claims of dialect: %w", err)
	}

	if dialectId == utils.LOCAL_CLAIM_DIALECT && utils.TOOL_CONFIGS.AllowDelete {
		localClaimDialectSummary.Success = true
	} else {
		utils.UpdateSuccessSummary(utils.CLAIMS, utils.UPDATE)
	}
	utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, dialectURI, "Updated successfully")
	return nil
}

func createDialect(dialectURI string) (dialectId string, err error) {

	dialectJSON, err := json.Marshal(map[string]string{"dialectURI": dialectURI})
	if err != nil {
		return "", fmt.Errorf("error serializing claim dialect: %w", err)
	}

	resp, err := utils.SendPostRequest(utils.CLAIMS, dialectJSON)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in create dialect response")
	}
	return path.Base(location), nil
}

func updateChangedClaims(dialectId string, localClaims, deployedClaims []map[string]interface{}) error {

	deployedByURI := make(map[string]map[string]interface{})
	for _, c := range deployedClaims {
		deployedByURI[getClaimURI(c)] = c
	}

	for _, claim := range localClaims {
		uri := getClaimURI(claim)
		if deployed, exists := deployedByURI[uri]; exists {
			if !claimChanged(dialectId, claim, deployed) {
				continue
			}
			if err := updateClaim(dialectId, getClaimID(deployed), claim); err != nil {
				return err
			}
		} else {
			if err := createClaim(dialectId, claim); err != nil {
				return err
			}
		}
	}
	return nil
}

func createClaim(dialectId string, claim map[string]interface{}) error {

	claimJSON, err := createClaimReqBody(dialectId, claim)
	if err != nil {
		return fmt.Errorf("error when marshalling claim %s: %w", getClaimURI(claim), err)
	}

	resp, err := utils.SendPostRequest(utils.CLAIMS, claimJSON, utils.WithPathSuffix(dialectId+"/claims"))
	if err != nil {
		return fmt.Errorf("error when creating claim %s: %w", getClaimURI(claim), err)
	}
	resp.Body.Close()
	return nil
}

func updateClaim(dialectId, claimId string, claim map[string]interface{}) error {

	claimJSON, err := createClaimReqBody(dialectId, claim)
	if err != nil {
		return fmt.Errorf("error when marshalling claim %s: %w", getClaimURI(claim), err)
	}

	resp, err := utils.SendPutRequest(utils.CLAIMS, dialectId+"/claims/"+claimId, claimJSON)
	if err != nil {
		return fmt.Errorf("error when updating claim %s: %w", getClaimURI(claim), err)
	}
	resp.Body.Close()
	return nil
}

func removeDeletedDeployedClaimdialect(localFiles []os.FileInfo, deployedClaimDialects []claimDialect) {

	if len(deployedClaimDialects) == 0 {
		return
	}

	localDialectNames := make(map[string]struct{})
	for _, file := range localFiles {
		localDialectNames[utils.GetFileInfo(file.Name()).ResourceName] = struct{}{}
	}

	for _, claimDialect := range deployedClaimDialects {
		if _, existsLocally := localDialectNames[formatFileName(claimDialect.DialectURI)]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(claimDialect.DialectURI, utils.TOOL_CONFIGS.ClaimConfigs) {
			utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, claimDialect.DialectURI, "Excluded from deletion.")
			continue
		}
		utils.PrintLog(utils.LogLevelInfo, utils.CLAIMS, claimDialect.DialectURI, "Not found locally. Deleting.")
		if err := utils.SendDeleteRequest(claimDialect.Id, utils.CLAIMS); err != nil {
			utils.UpdateFailureSummary(utils.CLAIMS, claimDialect.DialectURI)
			utils.PrintLog(utils.LogLevelError, utils.CLAIMS, claimDialect.DialectURI, fmt.Sprintf("Error deleting claim dialect: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.CLAIMS, utils.DELETE)
		}
	}
}

func removeDeletedDeployedClaims(dialectId string, deployedClaims, localClaims []map[string]interface{}) error {

	if len(deployedClaims) == 0 {
		return nil
	}

	localByURI := make(map[string]struct{})
	for _, claim := range localClaims {
		localByURI[getClaimURI(claim)] = struct{}{}
	}

	for _, deployed := range deployedClaims {
		if _, existsLocally := localByURI[getClaimURI(deployed)]; existsLocally {
			continue
		}
		if err := utils.SendDeleteRequest(dialectId+"/claims/"+getClaimID(deployed), utils.CLAIMS); err != nil {
			return fmt.Errorf("error deleting claim %s of dialect: %w", getClaimURI(deployed), err)
		}
	}
	return nil
}
