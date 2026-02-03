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
	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/configs"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/applications"
	claims "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/claims"
	identityproviders "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/identityProviders"
	userstores "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/userStores"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var importAllCmd = &cobra.Command{
	Use:   "importAll",
	Short: "Import all applications",
	Long:  `You can import all applications to the target environment`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDirPath, _ := cmd.Flags().GetString("inputDir")
		configFile, _ := cmd.Flags().GetString("config")

		baseDir := utils.LoadConfigs(configFile)
		if inputDirPath == "" {
			inputDirPath = baseDir
		}

		importFunctions := map[string]func(string){
			configs.CLAIMS:             claims.ImportAll,
			configs.IDENTITY_PROVIDERS: identityproviders.ImportAll,
			configs.APPLICATIONS:       applications.ImportAll,
			configs.USERSTORES:         userstores.ImportAll,
		}

		for _, resourceType := range configs.ResourceOrder {
			if importFunc, exists := importFunctions[resourceType]; exists {
				importFunc(inputDirPath)
			}
		}

		utils.PrintSummary(utils.IMPORT)
	},
}

func init() {

	cmd.RootCmd.AddCommand(importAllCmd)
	importAllCmd.Flags().StringP("inputDir", "i", "", "Path to the input directory")
	importAllCmd.Flags().StringP("config", "c", "", "Path to the environment specific config folder")
	importAllCmd.MarkFlagRequired("config")
}
