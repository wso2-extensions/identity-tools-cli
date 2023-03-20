/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
)

var serverConfigTemplate = map[string]string{

	"SERVER_URL":    "",
	"CLIENT_ID":     "",
	"CLIENT_SECRET": "",
	"TENANT_DOMAIN": "",
	"USERNAME":      "",
	"PASSWORD":      "",
}

var setupCmd = &cobra.Command{
	Use:   "setupCLI",
	Short: "Setup the CLI tool",
	Long:  `You can setup the config folder structure for the CLI tool`,
	Run: func(cmd *cobra.Command, args []string) {
		baseDirPath, _ := cmd.Flags().GetString("baseDir")

		createConfigFolders(baseDirPath)
	},
}

func init() {

	cmd.RootCmd.AddCommand(setupCmd)
	setupCmd.Flags().StringP("baseDir", "d", "", "Path to the base directory")
}

func createConfigFolders(baseDirPath string) {

	configFileName := "config.json"

	// If baseDirPath is not provided, create the config folder in the current working directory
	var err error
	if baseDirPath == "" {
		baseDirPath, err = os.Getwd()
		if err != nil {
			baseDirPath = "."
		}
	}

	// Create server config directory
	serverConfigDir := baseDirPath + "/configs/ServerConfigs/"
	os.MkdirAll(serverConfigDir, 0700)

	serverConfigs, err := json.Marshal(serverConfigTemplate)
	if err != nil {
		fmt.Println("Error in creating the server config template", err)
	}
	os.WriteFile(serverConfigDir+configFileName, serverConfigs, 0644)

	// Create tool config directory
	toolConfigDir := baseDirPath + "/configs/ToolConfigs/"
	os.MkdirAll(toolConfigDir, 0700)

	file, err := os.OpenFile(toolConfigDir+configFileName, os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error in creating the tool config file", err)
	}
	defer file.Close()
}
