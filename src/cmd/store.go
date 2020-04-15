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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var dir, _ = os.Getwd()
var path = dir + "/iamctl.json"

type ServerDetails struct {
	Server       string `json:"server"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type myJSON struct {
	Array []ServerDetails
}

func createFile() {

	// detect if file exists
	var _, err = os.Stat(path)
	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		checkError(err)
		defer file.Close()

		jsonData := &myJSON{Array: []ServerDetails{}}
		encodeJson, _ := json.Marshal(jsonData)

		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile(path, encodeJson, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func writeFiles(server string, token string, refreshToken string) {

	var err error
	var data myJSON
	var msg = new(ServerDetails)

	file, err := ioutil.ReadFile(path)
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
	err = ioutil.WriteFile(path, jsonData, 0644)
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println("Authorization is done for : " + server)
	}
	checkError(err)
}

func readFile() string {

	var a ServerDetails
	var data myJSON

	file, err := ioutil.ReadFile(path)
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
