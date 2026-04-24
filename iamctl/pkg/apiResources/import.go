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

package apiResources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing API resources...")
	importFilePath := filepath.Join(inputDirPath, utils.API_RESOURCES.String())

	if !utils.IsEntitySupportedInVersion(utils.API_RESOURCES) || utils.IsResourceTypeExcluded(utils.API_RESOURCES) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No API resources to import.")
		return
	}

	deployedResources, err := GetApiResourceList(true)
	if err != nil {
		log.Println("Error retrieving the deployed API resource list:", err)
		return
	}
	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing API resources:", err)
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		deployedResources = removeDeletedDeployedApiResources(files, deployedResources)
	}

	localScopeMap, err := readLocalScopesMap(importFilePath)
	if err != nil {
		log.Println("Error reading local scope name map:", err)
		utils.UpdateFailureSummary(utils.API_RESOURCES, utils.API_RESOURCE_SCOPES.String())
		return
	}
	failedResources := removeDeletedDeployedScopes(localScopeMap, deployedResources)

	for _, file := range files {
		apiResFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(apiResFilePath)
		resourceName := fileInfo.ResourceName

		if resourceName == utils.API_RESOURCE_SCOPES.String() {
			continue
		}
		if _, failed := failedResources[resourceName]; failed {
			log.Printf("Skipping API resource %s: deleting stale scopes failed", resourceName)
			utils.UpdateFailureSummary(utils.API_RESOURCES, resourceName)
			continue
		}
		if !utils.IsResourceExcluded(resourceName, utils.TOOL_CONFIGS.ApiResourceConfigs) {
			resourceId := getApiResourceId(resourceName, deployedResources)
			if err := importApiResource(resourceId, resourceName, apiResFilePath); err != nil {
				log.Println("Error importing API resource:", err)
				utils.UpdateFailureSummary(utils.API_RESOURCES, resourceName)
			}
		}
	}
}

func importApiResource(resourceId, resourceIdentifier, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for API resource: %w", err)
	}

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for API resource: %w", err)
	}

	keywordMapping := getApiResourceKeywordMapping(resourceIdentifier)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	if resourceId == "" {
		return createApiResource(resourceIdentifier, []byte(modifiedFileData), format)
	}
	return updateApiResource(resourceId, resourceIdentifier, []byte(modifiedFileData), format)
}

func createApiResource(resourceIdentifier string, requestBody []byte, format utils.Format) error {

	log.Println("Creating new API resource:", resourceIdentifier)

	jsonBody, err := utils.PrepareJSONRequestBody(requestBody, format, utils.API_RESOURCES,
		"id", "type", "properties", "authorizationDetailsTypes")
	if err != nil {
		return err
	}

	resp, err := utils.SendPostRequest(utils.API_RESOURCES, jsonBody)
	if err != nil {
		return fmt.Errorf("error when creating API resource: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading create API response: %w", err)
	}
	var created ApiResource
	if err := json.Unmarshal(body, &created); err != nil {
		return fmt.Errorf("error parsing create API response: %w", err)
	}
	utils.AddToIdentifierMap(utils.API_RESOURCES, created.ID, resourceIdentifier, utils.IMPORT)

	utils.UpdateSuccessSummary(utils.API_RESOURCES, utils.IMPORT)
	log.Println("API resource created successfully.")
	return nil
}

func updateApiResource(resourceId, resourceIdentifier string, requestBody []byte, format utils.Format) error {

	log.Println("Updating API resource:", resourceIdentifier)

	dataMap, err := utils.DeserializeToMap(requestBody, format, utils.API_RESOURCES)
	if err != nil {
		return fmt.Errorf("error deserializing file: %w", err)
	}
	scopes, ok := dataMap["scopes"]
	if !ok {
		return fmt.Errorf("scopes array not found in API resource data")
	}

	jsonBody, err := json.Marshal(scopes)
	if err != nil {
		return fmt.Errorf("error serializing update scopes request body: %w", err)
	}

	resp, err := utils.SendPutRequest(utils.API_RESOURCES, resourceId+"/scopes", jsonBody)
	if err != nil {
		return fmt.Errorf("error when updating API resource scopes: %w", err)
	}
	defer resp.Body.Close()

	utils.AddToIdentifierMap(utils.API_RESOURCES, resourceId, resourceIdentifier, utils.IMPORT)
	utils.UpdateSuccessSummary(utils.API_RESOURCES, utils.UPDATE)
	log.Println("API resource updated successfully.")
	return nil
}

func removeDeletedDeployedApiResources(localFiles []os.FileInfo, deployedResources []ApiResource) (remainingResources []ApiResource) {

	if len(deployedResources) == 0 {
		return deployedResources
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, resource := range deployedResources {
		if _, existsLocally := localResourceNames[resource.Identifier]; existsLocally {
			remainingResources = append(remainingResources, resource)
			continue
		}
		if utils.IsResourceExcluded(resource.Identifier, utils.TOOL_CONFIGS.ApiResourceConfigs) {
			log.Println("API resource is excluded from deletion:", resource.Identifier)
			remainingResources = append(remainingResources, resource)
			continue
		}

		log.Printf("API resource: %s not found locally. Deleting.\n", resource.Identifier)
		if err := utils.SendDeleteRequest(resource.ID, utils.API_RESOURCES); err != nil {
			utils.UpdateFailureSummary(utils.API_RESOURCES, resource.Identifier)
			log.Println("Error deleting API resource:", resource.Identifier, err)
			remainingResources = append(remainingResources, resource)
		} else {
			utils.UpdateSuccessSummary(utils.API_RESOURCES, utils.DELETE)
		}
	}
	return remainingResources
}

func removeDeletedDeployedScopes(localScopeMap map[string]string, deployedResources []ApiResource) (failedResources map[string]struct{}) {

	failedResources = make(map[string]struct{})

	for _, resource := range deployedResources {
		scopes, err := getApiResourceScopes(resource.ID)
		if err != nil {
			log.Printf("Error retrieving scopes for API resource %s: %v", resource.Identifier, err)
			failedResources[resource.Identifier] = struct{}{}
			continue
		}

		for _, scope := range scopes {
			localApiResName, scopeInLocalMap := localScopeMap[scope.Name]

			shouldDelete := false
			if !scopeInLocalMap && utils.TOOL_CONFIGS.AllowDelete {
				shouldDelete = true
			} else if scopeInLocalMap && localApiResName != resource.Identifier {
				shouldDelete = true
			}
			if !shouldDelete {
				continue
			}

			if err := utils.SendDeleteRequest(resource.ID+"/scopes/id/"+scope.ID, utils.API_RESOURCES); err != nil {
				log.Printf("Error deleting scope %s from API resource %s: %v", scope.Name, resource.Identifier, err)
				failedResources[resource.Identifier] = struct{}{}
			}
		}
	}

	return failedResources
}
