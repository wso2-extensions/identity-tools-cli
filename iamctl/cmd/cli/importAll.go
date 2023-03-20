/**
* Copyright (c) 2022, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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
	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/applications"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var importAllCmd = &cobra.Command{
	Use:   "importAll",
	Short: "You can import all service providers",
	Long:  `You can import all service providers`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDirPath, _ := cmd.Flags().GetString("inputDir")
		configFile, _ := cmd.Flags().GetString("config")

		utils.TOOL_CONFIGS = utils.LoadToolConfigsFromFile(configFile)
		utils.SERVER_CONFIGS = utils.LoadServerConfigsFromFile(utils.TOOL_CONFIGS.ServerConfigFileLocation)
		applications.ImportAll(inputDirPath)
	},
}

func init() {

	cmd.RootCmd.AddCommand(importAllCmd)
	importAllCmd.Flags().StringP("inputDir", "i", "", "Path to the input directory")
	importAllCmd.Flags().StringP("config", "c", "", "Path to the config file")
}