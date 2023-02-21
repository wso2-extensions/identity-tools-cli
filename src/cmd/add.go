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

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/url"
)

var createUsingCommand = &cobra.Command{
	Use:   "add",
	Short: "create SP using commands ",
	Long:  `A brief description about use flags to service provider `,
	Run: func(cmd *cobra.Command, args []string) {
		typeOfAPP, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		domain, _ := cmd.Flags().GetString("serverDomain")
		_, err = url.ParseRequestURI(domain)
		if err != nil && err.Error() != "parse : empty url" {
			fmt.Println(err)
		} else if err == nil {
			userName, _ := cmd.Flags().GetString("userName")
			password, _ := cmd.Flags().GetString("password")

			if userName == "" && password == "" {
				token := readFile()
				if token == "" {
					fmt.Println("required flag(s) \"password\",\"userName\" not set \nFlags:\n-u, --userName string       Username for Identity Server\n-p, --password string       Password for Identity Server")
					return
				}
			} else {
				if password == "" {
					fmt.Println("required flag(s) \"password\" not set \nFlag:\n-p, --password string       Password for Identity Server ")
					return
				} else if userName == "" {
					fmt.Println("required flag(s) \"userName\" not set \nFlag:\n-u, --userName string       Username for Identity Server ")
					return
				} else {
					SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()
					if CLIENTID == "" {
						setSampleSP()
						start(domain, userName, password)
						if readFile() == "" {
							return
						}
					} else {
						start(domain, userName, password)
						if readFile() == "" {
							return
						}
					}
				}
			}
		} else {
			SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()
			if CLIENTID == "" {
				setSampleSP()
				SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()
				setServerWithInit(SERVER)
				if readFile() == "" {
					return
				}
			} else {
				token := readFile()
				if token == "" {
					setServer()
					if readFile() == "" {
						return
					}
				}
			}
		}

		if typeOfAPP == "basic" {
			if description == "" {
				createSPBasicApplication(name, name)
			} else {
				createSPBasicApplication(name, description)
			}
		} else {
			callbackURl, _ := cmd.Flags().GetString("callbackURl")

			if callbackURl == "" {
				grantTypes := []string{"password", "client_credentials", "refresh_token", "urn:ietf:params:oauth:grant-type:device_code", "iwa:ntlm", "urn:ietf:params:oauth:grant-type:jwt-bearer", "account_switch", "urn:ietf:params:oauth:grant-type:saml2-bearer"}
				if description != "" {
					createSPOauthApplication(name, description, callbackURl, grantTypes)
				} else {
					createSPOauthApplication(name, name, callbackURl, grantTypes)
				}
			} else {
				grantTypes := []string{"authorization_code", "implicit", "password", "client_credentials", "refresh_token", "urn:ietf:params:oauth:grant-type:device_code", "iwa:ntlm", "urn:ietf:params:oauth:grant-type:jwt-bearer", "account_switch", "urn:ietf:params:oauth:grant-type:saml2-bearer"}
				_, err := url.ParseRequestURI(callbackURl)
				if err != nil {
					log.Fatalln(err)
				} else {
					if description != "" {
						createSPOauthApplication(name, description, callbackURl, grantTypes)
					} else {
						createSPOauthApplication(name, name, callbackURl, grantTypes)
					}
				}
			}
		}
	},
}

var serverDomain string
var applicationName string

func init() {

	createSPCmd.AddCommand(createUsingCommand)

	createUsingCommand.Flags().StringP("type", "t", "oauth", "Enter application type")
	name := createUsingCommand.Flags()
	name.StringVarP(&applicationName, "name", "n", "", "name of service provider - **compulsory")
	err := cobra.MarkFlagRequired(name, "name")
	if err != nil {
		log.Fatalln(err)
	}
	createUsingCommand.Flags().StringP("description", "d", "", "description of SP - **for basic application")
	createUsingCommand.Flags().StringP("callbackURl", "c", "", "callbackURL  of SP - **for oauth application")
	createUsingCommand.Flags().StringVarP(&serverDomain, "serverDomain", "s", "", "server Domain")
	createUsingCommand.Flags().StringP("userName", "u", "", "Username for Identity Server")
	createUsingCommand.Flags().StringP("password", "p", "", "Password for Identity Server")
}
