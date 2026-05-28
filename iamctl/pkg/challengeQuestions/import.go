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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string) {

	utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, "", "Importing challenge question sets...")
	importFilePath := filepath.Join(inputDirPath, utils.CHALLENGE_QUESTIONS.String())

	if !utils.IsEntitySupportedInVersion(utils.CHALLENGE_QUESTIONS) || !utils.IsEntitySupportedInOrg(utils.CHALLENGE_QUESTIONS) || utils.IsResourceTypeExcluded(utils.CHALLENGE_QUESTIONS) {
		return
	}
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, "", "No challenge question sets to import.")
		return
	}

	existingSets, err := getChallengeSetList()
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CHALLENGE_QUESTIONS, "", fmt.Sprintf("Error retrieving the deployed challenge question set list: %s", err))
		return
	}

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		utils.PrintLog(utils.LogLevelError, utils.CHALLENGE_QUESTIONS, "", fmt.Sprintf("Error importing challenge question sets: %s", err))
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
				utils.PrintLog(utils.LogLevelError, utils.CHALLENGE_QUESTIONS, setId, fmt.Sprintf("Error importing challenge question set: %s", err))
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

	fileBytes, err := ioutil.ReadFile(importFilePath)
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

	utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, setId, "Creating new challenge question set")

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
	utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, setId, "Created successfully")
	return nil
}

func updateChallengeSet(setId string, requestBody []byte, format utils.Format) error {

	utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, setId, "Updating challenge question set")

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
	utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, setId, "Updated successfully")
	return nil
}

func removeDeletedDeployedChallengeSets(localFiles []os.FileInfo, deployedSets []challengeSet) {

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
			utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, set.QuestionSetId, "Excluded from deletion")
			continue
		}

		utils.PrintLog(utils.LogLevelInfo, utils.CHALLENGE_QUESTIONS, set.QuestionSetId, "Not found locally. Deleting.")
		if err := utils.SendDeleteRequest(set.QuestionSetId, utils.CHALLENGE_QUESTIONS); err != nil {
			utils.UpdateFailureSummary(utils.CHALLENGE_QUESTIONS, set.QuestionSetId)
			utils.PrintLog(utils.LogLevelError, utils.CHALLENGE_QUESTIONS, set.QuestionSetId, fmt.Sprintf("Error deleting challenge question set: %s", err))
		} else {
			utils.UpdateSuccessSummary(utils.CHALLENGE_QUESTIONS, utils.DELETE)
		}
	}
}
