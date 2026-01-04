/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package interactive

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/components"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/utils"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   internal.STATUS_COMMAND,
	Short: "View the current login status of the tool",
	Long: `View the current login status of the tool.
	
	This command displays whether you are currently logged in or logged out of the IAMCTL tool.
	It provides information about the active session, including the organization name and server details if logged in.
	`,
	Example: internal.ROOT_COMMAND + " " + internal.STATUS_COMMAND,
	Run: func(cmd *cobra.Command, args []string) {
		status, orgName, serverName, err := utils.GetLoginDetails()
		fmt.Println(components.StylizeLoginStatus(status, orgName, serverName, err))
	},
}

func init() {
	cmd.RootCmd.AddCommand(statusCmd)
}
