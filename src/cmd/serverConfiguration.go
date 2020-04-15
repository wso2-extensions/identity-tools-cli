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
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
	"log"
	"net/url"
)

var configCmd = &cobra.Command{
	Use:   "serverConfiguration",
	Short: "you can set your server domain",
	Long:  `You can set your server domain`,
	Run: func(cmd *cobra.Command, args []string) {
		server, err := cmd.Flags().GetString("server")
		if err != nil {
			log.Fatalln(err)
		}

		if server == "" {
			SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()
			if CLIENTID == "" {
				setSampleSP()
				SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()
				setServerWithInit(SERVER)
			} else {
				setServer()
			}

		} else {
			_, err := url.ParseRequestURI(server)
			if err != nil {
				log.Fatalln(err)
			}
			userName, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			start(server, userName, password)
		}
	},
}

var server = []*survey.Question{
	{
		Name:     "server",
		Prompt:   &survey.Input{Message: "Enter IAM URL [<schema>://<host>]:"},
		Validate: survey.Required,
	},
}
var userNamePassword = []*survey.Question{
	{
		Name:     "username",
		Prompt:   &survey.Input{Message: "Enter Username:"},
		Validate: survey.Required,
	},
	{
		Name:     "password",
		Prompt:   &survey.Password{Message: "Enter Password:"},
		Validate: survey.Required,
	},
}

func init() {

	rootCmd.AddCommand(configCmd)
	configCmd.Flags().StringP("server", "s", "", "set server domain")
	configCmd.Flags().StringP("username", "u", "", "enter your username")
	configCmd.Flags().StringP("password", "p", "", "enter your password")
}

func setServer() {

	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render(appName)
	fmt.Print(renderStr)

	serverAnswer := struct {
		Server string `survey:"server"`
	}{}
	userNamePasswordAnswer := struct {
		UserName string `survey:"username"`
		Password string `survey:"password"`
	}{}

	err1 := survey.Ask(server, &serverAnswer)
	if err1 != nil {
		log.Fatal(err1)
		return
	}
	_, err = url.ParseRequestURI(serverAnswer.Server)
	if err != nil {
		log.Fatalln(err)
	}

	err1 = survey.Ask(userNamePassword, &userNamePasswordAnswer)
	if err1 != nil {
		log.Fatal(err1)
		return
	}

	start(serverAnswer.Server, userNamePasswordAnswer.UserName, userNamePasswordAnswer.Password)
}
func setServerWithInit(server string) {

	userNamePasswordAnswer := struct {
		UserName string `survey:"username"`
		Password string `survey:"password"`
	}{}
	err1 := survey.Ask(userNamePassword, &userNamePasswordAnswer)
	if err1 != nil {
		log.Fatal(err1)
		return
	}

	start(server, userNamePasswordAnswer.UserName, userNamePasswordAnswer.Password)
}
