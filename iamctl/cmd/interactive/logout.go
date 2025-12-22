/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package interactive

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/components"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/api"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   internal.LOGOUT_COMMAND,
	Short: "Logout and disconnect from the Server",
	Long:  "Logout and disconnect from the Server by removing the stored credentials.",
	Run: func(cmd *cobra.Command, args []string) {
		err := api.Logout()
		if err != nil {
			log.Println(fmt.Sprintf(components.StylizeErrorMessage("Error Logging out: %s"), err.Error()))
			return
		}
		log.Println(components.StylizeSuccessMessage("Successfully logged out!"))
	},
}

func init() {
	cmd.RootCmd.AddCommand(logoutCmd)
}
