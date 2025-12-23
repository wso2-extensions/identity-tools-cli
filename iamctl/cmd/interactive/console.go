/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package interactive

import (
	"log"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/components"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/utils"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		serverUrl, err := utils.GetConfigValue(internal.PREFIX_URL_KEY)
		if err != nil {
			log.Printf(components.StylizeErrorMessage("Error while getting server URL: %s"), err.Error())
		}
		if serverUrl == "" {
			log.Println(components.StylizeErrorMessage("Server URL is not set. Please login first."))
			return
		}
		consoleUrl := serverUrl + internal.CONSOLE_URL_SUFFIX

		err = browser.OpenURL(consoleUrl)
		if err != nil {
			log.Printf(components.StylizeErrorMessage("Error while opening console URL: %s"), err.Error())
			return
		}
		log.Println(components.StylizeSuccessMessage("Console opened successfully in the default browser."))

	},
}

func init() {
	cmd.RootCmd.AddCommand(consoleCmd)
}
