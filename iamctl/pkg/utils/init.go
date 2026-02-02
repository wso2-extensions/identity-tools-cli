/**
* Copyright (c) 2020-2025, WSO2 LLC. (https://www.wso2.com).
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

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var dir, _ = os.Getwd()
var Path = dir + "/iamctl.json"
var PathSampleSPDetails = dir + "/init.json"

const SCOPE string = "internal_application_mgt_update internal_application_mgt_create internal_application_mgt_view internal_application_mgt_delete internal_idp_update internal_idp_create internal_idp_view internal_idp_delete internal_userstore_view internal_userstore_create internal_userstore_update internal_userstore_delete internal_claim_meta_create internal_claim_meta_view internal_claim_meta_update internal_claim_meta_delete internal_oidc_scope_mgt_view internal_oidc_scope_mgt_create internal_oidc_scope_mgt_update internal_oidc_scope_mgt_delete internal_org_idp_delete internal_org_idp_view internal_org_idp_update internal_org_idp_create internal_org_claim_meta_update internal_org_claim_meta_view internal_org_application_mgt_update internal_org_application_mgt_create internal_org_application_mgt_view internal_org_application_mgt_delete internal_org_userstore_create internal_org_userstore_view internal_org_userstore_delete internal_org_userstore_update " +
	"internal_application_mgt_client_secret_create  internal_org_application_mgt_client_secret_create internal_application_mgt_client_secret_view internal_org_application_mgt_client_secret_view " +
	"internal_application_script_update internal_org_application_script_update " +
	"internal_application_business_api_update internal_application_internal_api_update internal_org_application_business_api_update internal_org_application_internal_api_update"

const (
	AppName       = "IAM-CTL"
	ShortAppDesc  = "Service Provider configuration"
	LongAPPConfig = "Service Provider configuration"
)

type SampleSP struct {
	Server       string `json:"server"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	Tenant       string `json:"tenant"`
}

type ServerDetails struct {
	Server       string `json:"server"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type MyJSON struct {
	Array []ServerDetails
}

func CreateFile() {

	// detect if file exists
	var _, err = os.Stat(Path)
	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(Path)
		checkError(err)
		defer file.Close()

		jsonData := &MyJSON{Array: []ServerDetails{}}
		encodeJson, _ := json.Marshal(jsonData)

		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile(Path, encodeJson, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func WriteFiles(server string, token string, refreshToken string) {

	var err error
	var data MyJSON
	var msg = new(ServerDetails)

	file, err := ioutil.ReadFile(Path)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Fatalln(err)
	}

	msg.AccessToken = token
	msg.Server = server
	msg.RefreshToken = refreshToken

	if len(data.Array) == 0 {
		data.Array = append(data.Array, *msg)
	} else {
		for i := 0; i < len(data.Array); i++ {
			if data.Array[i].Server == server {
				data.Array[i].AccessToken = token
				data.Array[i].RefreshToken = refreshToken
			} else {
				data.Array = append(data.Array, *msg)
			}
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile(Path, jsonData, 0644)
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println("Authorization is done for : " + server)
	}
	checkError(err)
}

func ReadFile() string {

	var a ServerDetails
	var data MyJSON

	file, err := ioutil.ReadFile(Path)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Fatalln(err)
	}
	//as the single host this worked. For multiple host need to read relevant accessToken according to given server
	for i := 0; i < len(data.Array); i++ {
		a = data.Array[i]
	}
	return a.AccessToken
}

func checkError(err error) {

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}

func CreateSampleSPFile() {

	// detect if file exists
	var _, err = os.Stat(PathSampleSPDetails)
	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(PathSampleSPDetails)
		checkError(err)
		defer file.Close()
		jsonData := &SampleSP{}
		encodeJson, _ := json.Marshal(jsonData)
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile(PathSampleSPDetails, encodeJson, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func ReadSPConfig() (string, string, string, string) {

	var data SampleSP

	file, _ := ioutil.ReadFile(PathSampleSPDetails)
	err := json.Unmarshal(file, &data)
	if err != nil {
		log.Fatalln(err)
	}

	return data.Server, data.ClientID, data.ClientSecret, data.Tenant
}
