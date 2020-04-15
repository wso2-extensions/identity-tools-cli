/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ServiceProvider struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func createSPBasicApplication(spName string, spDescription string) {

	SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()

	var ADDAPPURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications"
	var err error
	var status int

	token := readFile()

	toJson := ServiceProvider{spName, spDescription}
	jsonData, err := json.Marshal(toJson)
	if err != nil {
		log.Fatalln(err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("POST", ADDAPPURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	status = resp.StatusCode
	defer resp.Body.Close()

	if status == 401 {
		fmt.Println("Unauthorized access.\nPlease enter your Username and password for server.")
		setServerWithInit(SERVER)
		createSPBasicApplication(spName, spDescription)
	} else if status == 400 {
		fmt.Println("Provided parameters are not in correct format.")
	} else if status == 403 {
		fmt.Println("Forbidden")
	} else if status == 201 {
		fmt.Println("Successfully created the service provider named '" + spName + "' at " + resp.Header.Get("Date"))
	} else if status == 409 {
		fmt.Println("Already exists an application with same name:" + spName)
	}
}
