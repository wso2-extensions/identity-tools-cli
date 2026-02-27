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

package challengeQuestions

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing challenge question sets...")
	importFilePath := filepath.Join(inputDirPath, utils.CHALLENGE_QUESTIONS.String())

	if utils.IsResourceTypeExcluded(utils.CHALLENGE_QUESTIONS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No challenge question sets to import.")
		return
	}

	existingSets, err := getChallengeSetList()
	if err != nil {
		log.Println("Error retrieving the deployed challenge question set list:", err)
		return
	}

	files, err := os.ReadDir(importFilePath)
	if err != nil {
		log.Println("Error importing challenge question sets:", err)
		return
	}
	if utils.TOOL_CONFIGS.AllowDelete {
		removeDeletedDeployedChallengeSets(files, existingSets)
	}

	for _, file := range files {
		setFilePath := filepath.Join(importFilePath, file.Name())
		fileInfo := utils.GetFileInfo(setFilePath)
		setId := fileInfo.ResourceName

		if !utils.IsResourceExcluded(setId, utils.TOOL_CONFIGS.ChallengeQuestionConfigs) {
			setExists := isChallengeSetExists(setId, existingSets)
			err := importChallengeSet(setId, setExists, setFilePath)
			if err != nil {
				log.Println("Error importing challenge question set:", err)
				utils.UpdateFailureSummary(utils.CHALLENGE_QUESTIONS, setId)
			}
		}
	}
}

func importChallengeSet(setId string, setExists bool, importFilePath string) error {

	format, err := utils.FormatFromExtension(filepath.Ext(importFilePath))
	if err != nil {
		return fmt.Errorf("unsupported file format for challenge question set: %w", err)
	}

	fileBytes, err := os.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for challenge question set: %w", err)
	}

	keywordMapping := getChallengeQuestionKeywordMapping(setId)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), keywordMapping)

	if !setExists {
		return createChallengeSet([]byte(modifiedFileData), format, setId)
	}
	return updateChallengeSet(setId, []byte(modifiedFileData), format)
}

func createChallengeSet(requestBody []byte, format utils.Format, setId string) error {

	log.Println("Creating new challenge question set:", setId)

	parsed, err := utils.Deserialize(requestBody, format, utils.CHALLENGE_QUESTIONS)
	if err != nil {
		return fmt.Errorf("error deserializing data: %w", err)
	}
	wrappedBody, err := utils.Serialize([]interface{}{parsed}, utils.FormatJSON, utils.CHALLENGE_QUESTIONS)
	if err != nil {
		return fmt.Errorf("error serializing to JSON: %w", err)
	}

	resp, err := utils.SendPostRequest(utils.CHALLENGE_QUESTIONS, wrappedBody)
	if err != nil {
		return fmt.Errorf("error when creating challenge question set: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.CHALLENGE_QUESTIONS, utils.IMPORT)
	log.Println("Challenge question set created successfully.")
	return nil
}

func updateChallengeSet(setId string, requestBody []byte, format utils.Format) error {

	log.Println("Updating challenge question set:", setId)

	questionsBytes, err := buildUpdateRequestBody(requestBody, format)
	if err != nil {
		return err
	}

	resp, err := utils.SendPutRequest(utils.CHALLENGE_QUESTIONS, setId, questionsBytes)
	if err != nil {
		return fmt.Errorf("error when updating challenge question set: %w", err)
	}
	defer resp.Body.Close()

	utils.UpdateSuccessSummary(utils.CHALLENGE_QUESTIONS, utils.UPDATE)
	log.Println("Challenge question set updated successfully.")
	return nil
}

func removeDeletedDeployedChallengeSets(localFiles []os.DirEntry, deployedSets []challengeSet) {

	if len(deployedSets) == 0 {
		return
	}

	localResourceNames := make(map[string]struct{})
	for _, file := range localFiles {
		resourceName := utils.GetFileInfo(file.Name()).ResourceName
		localResourceNames[resourceName] = struct{}{}
	}

	for _, set := range deployedSets {
		if _, existsLocally := localResourceNames[set.QuestionSetId]; existsLocally {
			continue
		}
		if utils.IsResourceExcluded(set.QuestionSetId, utils.TOOL_CONFIGS.ChallengeQuestionConfigs) {
			log.Println("Challenge question set is excluded from deletion:", set.QuestionSetId)
			continue
		}

		log.Printf("Challenge question set: %s not found locally. Deleting.\n", set.QuestionSetId)
		if err := utils.SendDeleteRequest(set.QuestionSetId, utils.CHALLENGE_QUESTIONS); err != nil {
			utils.UpdateFailureSummary(utils.CHALLENGE_QUESTIONS, set.QuestionSetId)
			log.Println("Error deleting challenge question set:", set.QuestionSetId, err)
		} else {
			utils.UpdateSuccessSummary(utils.CHALLENGE_QUESTIONS, utils.DELETE)
		}
	}
}
