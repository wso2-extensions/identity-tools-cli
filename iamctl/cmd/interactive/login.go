/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package interactive

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	question "github.com/wso2-extensions/identity-tools-cli/iamctl/cmd/interactive/survey"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/api"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/styles"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   internal.LOGIN_COMMAND,
	Short: "Login and connect with the Server",
	Long: `Login and connect with the Server using Client ID, Client Secret and Organization Name (Asgardeo) or Tenant Domain (Identity Server)
		You will be asked to select the server type (Asgardeo/Identity Server) and provide the required details to login.
		You can provide the Client ID and Organization Name (Asgardeo) or Tenant Domain (Identity Server) as flags, or you will be prompted to enter them interactively.
		You can also provide the Client Secret as a flag. If not provided, you will be prompted to enter it securely. We recommend using flags for non-interactive usage (Automation) and secure prompts for interactive usage.`,
	Example: internal.ROOT_COMMAND + " " + internal.LOGIN_COMMAND + ` --client-id <client-id> --org <org-name> --client-secret <client-secret> --server-type "Asgardeo"` + "\n" +
		internal.ROOT_COMMAND + " " + internal.LOGIN_COMMAND + ` --client-id <client-id> --org <tenant-domain> --client-secret <client-secret> --server-type "Identity Server" --identity-server-url https://localhost:9443` + "\n" + internal.ROOT_COMMAND + " " + internal.LOGIN_COMMAND,
	Run: func(cmd *cobra.Command, args []string) {
		var loginTheme = styles.GetLoginTheme()

		clientIDFlag, _ := cmd.Flags().GetString("client-id")
		orgNameFlag, _ := cmd.Flags().GetString("org")
		clientSecretFlag, _ := cmd.Flags().GetString("client-secret")
		serverTypeFlag, _ := cmd.Flags().GetString("server-type")
		identityServerURLFlag, _ := cmd.Flags().GetString("identity-server-url")

		var serverType string
		if serverTypeFlag != "" {
			serverType = serverTypeFlag
		} else {
			if err := question.SelectServerPrompt.Value(&serverType).WithTheme(loginTheme).Run(); err != nil {
				fmt.Println("Error while selecting server type:", err)
				return
			}
		}

		var identityServerURL string
		if serverType == internal.IS {
			if identityServerURLFlag != "" {
				identityServerURL = identityServerURLFlag
			} else {
				if err := question.IdentityServerURLPrompt.Value(&identityServerURL).WithTheme(loginTheme).Run(); err != nil {
					fmt.Println("Error while entering Identity Server URL:", err)
					return
				}
			}
		}

		var orgName string
		if orgNameFlag != "" {
			orgName = orgNameFlag
		} else {
			if err := question.OrgNamePrompt.Value(&orgName).WithTheme(loginTheme).Run(); err != nil {
				fmt.Println("Error while entering Organization Name/Tenant Domain:", err)
				return
			}
		}

		var clientID string
		if clientIDFlag != "" {
			clientID = clientIDFlag
		} else {
			if err := question.ClientIDPrompt.Value(&clientID).WithTheme(loginTheme).Run(); err != nil {
				fmt.Println("Error while entering Client ID:", err)
				return
			}
		}

		var clientSecret string
		if clientSecretFlag != "" {
			clientSecret = clientSecretFlag
		} else {
			if err := question.ClientSecretPrompt.Value(&clientSecret).WithTheme(loginTheme).Run(); err != nil {
				fmt.Println("Error while entering Client Secret:", err)
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

	loginCmd.Flags().StringP("client-id", "c", "", "Client ID of the M2M application")
	loginCmd.Flags().StringP("org", "o", "", "Name of the Organization (Asgardeo) or Tenant Domain (Identity Server)")
	loginCmd.Flags().StringP("client-secret", "s", "", "Client Secret of the M2M application")
	loginCmd.Flags().StringP("server-type", "t", "", "Type of the server to connect to (Asgardeo/Identity Server)")
	loginCmd.Flags().StringP("identity-server-url", "i", "", "URL of the Identity Server in the format [<protocol>://<host>] (Example https://localhost:9443)")
}
