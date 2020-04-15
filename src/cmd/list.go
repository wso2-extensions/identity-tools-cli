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
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"
)

var getListCmd = &cobra.Command{
	Use:   "list",
	Short: "get service providers List",
	Long:  `You can get service provider List`,
	Run: func(cmd *cobra.Command, args []string) {
		server, err := cmd.Flags().GetString("server")
		if server == "" {
			ascii := figlet4go.NewAsciiRender()
			renderStr, _ := ascii.Render(appName)
			fmt.Print(renderStr)

			getList()
		} else {
			_, err = url.ParseRequestURI(server)
			if err != nil && err.Error() != "parse : empty url" {
				log.Fatalln(err)
			} else if err == nil {
				userName, _ := cmd.Flags().GetString("userName")
				password, _ := cmd.Flags().GetString("password")

				if userName == "" && password == "" {
					token := readFile()
					if token == "" {
						fmt.Println("required flag(s) \"password\",\"userName\" not set \nFlags:\n-u, --userName string       Username for Identity Server\n-p, --password string       Password for Identity Server")
						return
					} else {
						getList()
					}
				} else {
					if password == "" {
						token := readFile()
						if token == "" {
							fmt.Println("required flag(s) \"password\" not set \nFlag:\n-p, --password string       Password for Identity Server ")
							return
						} else {
							getList()
						}
					} else if userName == "" {
						token := readFile()
						if token == "" {
							fmt.Println("required flag(s) \"userName\" not set \nFlag:\n-u, --userName string       Username for Identity Server ")
							return
						} else {
							getList()
						}

					} else {
						SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()
						if CLIENTID == "" {
							setSampleSP()
							start(server, userName, password)
							if readFile() == "" {
								return
							} else {
								getList()
							}
						} else {
							start(server, userName, password)
							if readFile() == "" {
								return
							} else {
								getList()
							}
						}
					}
				}
			}
		}
	},
}

type List struct {
	TotalResults int           `json:"totalResults"`
	StartIndex   int           `json:"startIndex"`
	Count        int           `json:"count"`
	Applications []Application `json:"applications"`
	Links        []string      `json:"links"`
}
type Application struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Self        string `json:"self"`
}

func init() {

	createSPCmd.AddCommand(getListCmd)

	getListCmd.Flags().StringP("server", "s", "", "server")
	getListCmd.Flags().StringP("userName", "u", "", "User name for Identity Server")
	getListCmd.Flags().StringP("password", "p", "", "Password for Identity Server")
}
func getList() {

	SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()

	var GETLISTURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications"
	var status int
	var list List
	var app Application

	token := readFile()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", GETLISTURL, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("accept", "*/*")
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
		if readFile() == "" {
			return
		} else {
			getList()
		}
	} else if status == 400 {
		fmt.Println("Bad Request")
	} else if status == 200 {
		fmt.Println("Successfully Got the service provider List at " + resp.Header.Get("Date"))
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		writer := new(tabwriter.Writer)
		writer.Init(os.Stdout, 8, 8, 0, '\t', 0)
		defer writer.Flush()

		err = json.Unmarshal(body, &list)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintf(writer, "\n %s\t%s\t%s\t", "Application Id ", "Name", "Description")
		fmt.Fprintf(writer, "\n %s\t%s\t%s\t", " ----", "----", "----", )

		for i := 0; i < len(list.Applications); i++ {
			app.Id = list.Applications[i].Id
			app.Name = list.Applications[i].Name
			app.Description = list.Applications[i].Description
			fmt.Fprintf(writer, "\n %s\t%s\t%s\t", app.Id, app.Name, app.Description)
		}

		resp.Body.Close()

	} else if status == 403 {
		fmt.Println("Forbidden")
	} else if status == 404 {
		fmt.Println("Not Found")
	} else if status == 500 {
		fmt.Println("Server Error")
	} else if status == 501 {
		fmt.Println("Not Implemented")
	}
}
