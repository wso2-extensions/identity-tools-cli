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

func ExportAll(exportFilePath string, format string) {

	log.Println("Exporting challenge question sets...")
	exportFilePath = filepath.Join(exportFilePath, utils.CHALLENGE_QUESTIONS.String())

	if utils.IsResourceTypeExcluded(utils.CHALLENGE_QUESTIONS) {
		return
	}
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedSetIds := getDeployedChallengeSetIds()
			utils.RemoveDeletedLocalResources(exportFilePath, deployedSetIds)
		}
	}

	sets, err := getChallengeSetList()
	if err != nil {
		log.Println("Error: when exporting challenge question sets.", err)
		return
	}

	for _, set := range sets {
		if !utils.IsResourceExcluded(set.QuestionSetId, utils.TOOL_CONFIGS.ChallengeQuestionConfigs) {
			log.Println("Exporting challenge question set:", set.QuestionSetId)
			err := exportChallengeSet(set.QuestionSetId, exportFilePath, format)
			if err != nil {
				utils.UpdateFailureSummary(utils.CHALLENGE_QUESTIONS, set.QuestionSetId)
				log.Printf("Error while exporting challenge question set: %s. %s", set.QuestionSetId, err)
			} else {
				utils.UpdateSuccessSummary(utils.CHALLENGE_QUESTIONS, utils.EXPORT)
				log.Println("Challenge question set exported successfully:", set.QuestionSetId)
			}
		}
	}
}

func exportChallengeSet(setId string, outputDirPath string, formatString string) error {

	set, err := utils.GetResourceData(utils.CHALLENGE_QUESTIONS, setId)
	if err != nil {
		return fmt.Errorf("error while getting challenge question set: %w", err)
	}

	format := utils.FormatFromString(formatString)
	exportedFileName := utils.GetExportedFilePath(outputDirPath, setId, format)

	keywordMapping := getChallengeQuestionKeywordMapping(setId)
	modifiedSet, err := utils.ProcessExportedData(set, exportedFileName, format, keywordMapping, utils.CHALLENGE_QUESTIONS)
	if err != nil {
		return fmt.Errorf("error while processing exported content: %w", err)
	}

	modifiedFile, err := utils.Serialize(modifiedSet, format, utils.CHALLENGE_QUESTIONS)
	if err != nil {
		return fmt.Errorf("error while serializing challenge question set: %w", err)
	}

	if err = os.WriteFile(exportedFileName, modifiedFile, 0644); err != nil {
		return fmt.Errorf("error when writing exported content to file: %w", err)
	}

	return nil
}
