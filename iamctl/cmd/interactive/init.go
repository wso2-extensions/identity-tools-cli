/*
 * Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package interactive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var sampleSPCmd = &cobra.Command{
	Use:   "init",
	Short: "you can set your sample SP",
	Long:  "You can set your sample service provider client secret, client id, tenant ",
	Run: func(cmd *cobra.Command, args []string) {
		setSampleSP()
	},
}

var serverInit = []*survey.Question{
	{
		Name:     "server",
		Prompt:   &survey.Input{Message: "Enter IAM URL [<schema>://<host>]:"},
		Validate: survey.Required,
	},
}
var sampleSP = []*survey.Question{
	{
		Name:     "clientID",
		Prompt:   &survey.Input{Message: "Enter Client key:"},
		Validate: survey.Required,
	},
	{
		Name:     "clientSecret",
		Prompt:   &survey.Password{Message: "Enter Client Secret:"},
		Validate: survey.Required,
	},
	{
		Name:     "tenantDomain",
		Prompt:   &survey.Input{Message: "Enter Tenant domain:"},
		Validate: survey.Required,
	},
}

func init() {

	cmd.RootCmd.AddCommand(sampleSPCmd)
}

func setSampleSP() {

	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render(utils.AppName)
	fmt.Print(renderStr)

	sampleServer := struct {
		Server string `survey:"server"`
	}{}
	sampleSPAnswer := struct {
		ClientID     string `survey:"clientID"`
		ClientSecret string `survey:"clientSecret"`
		Tenant       string `survey:"tenantDomain"`
	}{}

	err := survey.Ask(serverInit, &sampleServer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	_, err = url.ParseRequestURI(sampleServer.Server)
	if err != nil {
		log.Fatalln(err)
	} else {
		err1 := survey.Ask(sampleSP, &sampleSPAnswer)
		if err1 != nil {
			fmt.Println(err1.Error())
			return
		}
		writeSampleAPPFile(sampleServer.Server, sampleSPAnswer.ClientID, sampleSPAnswer.ClientSecret, sampleSPAnswer.Tenant)
	}
}

func writeSampleAPPFile(server string, clientID string, clientSecret string, tenant string) {

	var data utils.SampleSP
	file, _ := ioutil.ReadFile(utils.PathSampleSPDetails)

	err := json.Unmarshal(file, &data)
	if err != nil {
		log.Fatalln(err)
	}

	data.Server = server
	data.ClientID = clientID
	data.ClientSecret = clientSecret
	data.Tenant = tenant

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}

	_ = ioutil.WriteFile(utils.PathSampleSPDetails, jsonData, 0644)
	fmt.Println("successfully set service provider  Client_key: " + clientID + " Client_Secret: ****************************  Tenant Domain " + tenant + " in " + server)
}

func readSPConfig() (string, string, string, string) {

	var data utils.SampleSP

	file, _ := ioutil.ReadFile(utils.PathSampleSPDetails)
	err := json.Unmarshal(file, &data)
	if err != nil {
		log.Fatalln(err)
	}

	return data.Server, data.ClientID, data.ClientSecret, data.Tenant
}
