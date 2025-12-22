/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package interactive

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	questions "github.com/wso2-extensions/identity-tools-cli/iamctl/cmd/interactive/survey"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/api"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"

	"github.com/AlecAivazis/survey/v2"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   internal.LOGIN_COMMAND,
	Short: "Login and connect with the Server",
	Long: `Login and connect with the Server using Client ID, Client Secret and Organization Name (Asgardeo) or Tenant Domain (Identity Server)
		You will be asked to select the server type (Asgardeo/Identity Server) and provide the required details to login.
		You can provide the Client ID and Organization Name (Asgardeo) or Tenant Domain (Identity Server) as flags, or you will be prompted to enter them interactively.
		You can also provide the Client Secret as a flag. If not provided, you will be prompted to enter it securely. We recommend using flags for non-interactive usage (Automation) and secure prompts for interactive usage.`,
	Example: internal.ROOT_COMMAND + " " + internal.LOGIN_COMMAND + ` -c <client-id> -o <org-name>` + "\n" + internal.ROOT_COMMAND + " " + internal.LOGIN_COMMAND,
	Run: func(cmd *cobra.Command, args []string) {
		clientIDFlag, _ := cmd.Flags().GetString("client-id")
		orgNameFlag, _ := cmd.Flags().GetString("org")
		clientSecretFlag, _ := cmd.Flags().GetString("client-secret")

		var serverType string

		err := survey.Ask(questions.ServerPrompt, &serverType)
		if err != nil {
			fmt.Println("Error reading server type:", err)
			return
		}
		var identityServerURL string
		if serverType == internal.IS {
			err := survey.AskOne(questions.IdentityServerURLPrompt, &identityServerURL)
			if err != nil {
				fmt.Println("Error reading server url:", err)
				return
			}
		}

		var orgName string
		if orgNameFlag != "" {
			orgName = orgNameFlag
		} else {
			err := survey.AskOne(questions.OrgNamePrompt, &orgName)
			if err != nil {
				fmt.Println("Error reading Organization Name:", err)
				return
			}
		}

		var clientID string
		if clientIDFlag != "" {
			clientID = clientIDFlag
		} else {
			err := survey.AskOne(questions.ClientIDPrompt, &clientID)
			if err != nil {
				fmt.Println("Error reading Client ID:", err)
				return
			}
		}

		var clientSecret string
		if clientSecretFlag != "" {
			clientSecret = clientSecretFlag
		} else {
			err := survey.AskOne(questions.ClientSecretPrompt, &clientSecret)
			if err != nil {
				fmt.Println("Error reading Client Secret:", err)
				return
			}
		}
		err = api.Login(serverType, clientID, clientSecret, orgName, identityServerURL)
		if err != nil {
			fmt.Println("Error during login:", err)
			return
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	loginCmd.Flags().StringP("client-id", "c", "", "Client ID of the M2M application")
	loginCmd.Flags().StringP("org", "o", "", "Name of the Organization (Asgardeo) or Tenant Domain (Identity Server)")
	loginCmd.Flags().StringP("client-secret", "s", "", "Client Secret of the M2M application")
}
