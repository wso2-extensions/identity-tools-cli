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
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/applications"
	challengeQuestions "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/challengeQuestions"
	claims "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/claims"
	emailTemplates "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/emailTemplates"
	governanceConnectors "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/governanceConnectors"
	identityproviders "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/identityProviders"
	oidcScopes "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/oidcScopes"
	roles "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/roles"
	scriptLibraries "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/scriptLibraries"
	userstores "github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/userStores"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var exportAllCmd = &cobra.Command{
	Use:   "exportAll",
	Short: "Export all resources",
	Long:  `You can export all resources available in the target environment`,
	Run: func(cmd *cobra.Command, args []string) {
		outputDirPath, _ := cmd.Flags().GetString("outputDir")
		format, _ := cmd.Flags().GetString("format")
		configFile, _ := cmd.Flags().GetString("config")

		baseDir := utils.LoadConfigs(configFile)
		if outputDirPath == "" {
			outputDirPath = baseDir
		}

		exportFunctions := map[utils.ResourceType]func(string, string){
			utils.CLAIMS:                claims.ExportAll,
			utils.IDENTITY_PROVIDERS:    identityproviders.ExportAll,
			utils.APPLICATIONS:          applications.ExportAll,
			utils.USERSTORES:            userstores.ExportAll,
			utils.OIDC_SCOPES:           oidcScopes.ExportAll,
			utils.ROLES:                 roles.ExportAll,
			utils.CHALLENGE_QUESTIONS:   challengeQuestions.ExportAll,
			utils.EMAIL_TEMPLATES:       emailTemplates.ExportAll,
			utils.SCRIPT_LIBRARIES:      scriptLibraries.ExportAll,
			utils.GOVERNANCE_CONNECTORS: governanceConnectors.ExportAll,
		}

		for _, resourceType := range utils.ResourceOrder {
			if exportFunc, exists := exportFunctions[resourceType]; exists {
				exportFunc(outputDirPath, format)
			}
		}

		utils.PrintSummary(utils.EXPORT)
	},
}

func init() {

	cmd.RootCmd.AddCommand(exportAllCmd)
	exportAllCmd.Flags().StringP("outputDir", "o", "", "Path to the output directory")
	exportAllCmd.Flags().StringP("format", "f", "yaml", "Format of the exported files")
	exportAllCmd.Flags().StringP("config", "c", "", "Path to the environment specific config folder")
}
