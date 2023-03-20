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

// var SERVER_CONFIGS utils.EnvConfigs

var exportAllCmd = &cobra.Command{
	Use:   "exportAll",
	Short: "You can export all service providers",
	Long:  `You can export all service providers`,
	Run: func(cmd *cobra.Command, args []string) {
		outputDirPath, _ := cmd.Flags().GetString("outputDir")
		format, _ := cmd.Flags().GetString("format")
		configFile, _ := cmd.Flags().GetString("config")

		utils.TOOL_CONFIGS = utils.LoadToolConfigsFromFile(configFile)
		utils.SERVER_CONFIGS = utils.LoadServerConfigsFromFile(utils.TOOL_CONFIGS.ServerConfigFileLocation)

		applications.ExportAll(outputDirPath, format)
	},
}

func init() {

	cmd.RootCmd.AddCommand(exportAllCmd)
	exportAllCmd.Flags().StringP("outputDir", "o", "", "Path to the output directory")
	exportAllCmd.Flags().StringP("format", "f", "yaml", "Format of the exported files")
	exportAllCmd.Flags().StringP("config", "c", "", "Path to the config file")
}
