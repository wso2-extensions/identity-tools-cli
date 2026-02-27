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
	"encoding/json"
	"fmt"
	"io"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type challengeSet struct {
	QuestionSetId string `json:"questionSetId"`
}

func getChallengeSetList() ([]challengeSet, error) {

	var list []challengeSet
	resp, err := utils.SendGetListRequest(utils.CHALLENGE_QUESTIONS, -1)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving challenge question set list. %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error when reading the retrieved challenge question set list. %w", err)
		}
		if err = json.Unmarshal(body, &list); err != nil {
			return nil, fmt.Errorf("error when unmarshalling the retrieved challenge question set list. %w", err)
		}
		return list, nil
	} else if errMsg, ok := utils.ErrorCodes[statusCode]; ok {
		return nil, fmt.Errorf("error while retrieving challenge question set list. Status code: %d, Error: %s", statusCode, errMsg)
	}
	return nil, fmt.Errorf("error while retrieving challenge question set list")
}

func getDeployedChallengeSetIds() []string {

	sets, err := getChallengeSetList()
	if err != nil {
		return []string{}
	}

	var ids []string
	for _, set := range sets {
		ids = append(ids, set.QuestionSetId)
	}
	return ids
}

func getChallengeQuestionKeywordMapping(setId string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ChallengeQuestionConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(setId, utils.KEYWORD_CONFIGS.ChallengeQuestionConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func isChallengeSetExists(setId string, existingSets []challengeSet) bool {

	for _, set := range existingSets {
		if set.QuestionSetId == setId {
			return true
		}
	}
	return false
}

func buildUpdateRequestBody(requestBody []byte, format utils.Format) ([]byte, error) {

	parsed, err := utils.Deserialize(requestBody, format, utils.CHALLENGE_QUESTIONS)
	if err != nil {
		return nil, fmt.Errorf("error deserializing data: %w", err)
	}

	if interfaceMap, ok := parsed.(map[interface{}]interface{}); ok {
		parsed = utils.ConvertToStringKeyMap(interfaceMap)
	}
	dataMap, ok := parsed.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format for challenge question set")
	}

	questions, ok := dataMap["questions"]
	if !ok {
		return nil, fmt.Errorf("questions field not found in the question set data")
	}

	questionsBytes, err := utils.Serialize(questions, utils.FormatJSON, utils.CHALLENGE_QUESTIONS)
	if err != nil {
		return nil, fmt.Errorf("error serializing to JSON: %w", err)
	}

	return questionsBytes, nil
}
