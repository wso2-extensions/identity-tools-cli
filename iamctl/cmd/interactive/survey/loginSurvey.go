package survey

import (
	"errors"
	"regexp"

	"github.com/charmbracelet/huh"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

func ValidateIdentityServerURL(s string) error {
	regex := `^https?://[a-zA-Z0-9.-]+(:[0-9]+)?$`
	matched, err := regexp.MatchString(regex, s)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("invalid URL format, expected format: [<protocol>://<host>] (Example https://localhost:9443)")
	}
	return nil
}

var SelectServerPrompt = huh.NewSelect[string]().Title("Select the server type:").Options(
	huh.NewOption("Asgardeo", internal.ASGARDEO),
	huh.NewOption("Identity Server", internal.IS),
).Description("Choose the server you want to connect to.")

var ClientIDPrompt = huh.NewInput().Title("Enter Client ID:").Description("Enter the Client ID for the application. You can find it from the 'protocol' tab of the application.")

var OrgNamePrompt = huh.NewInput().Title("Enter Organization Name:").Description("Enter the Organization Name (Asgardeo) or Tenant Domain (Identity Server). You can find it in the console url of the server.")

var ClientSecretPrompt = huh.NewInput().EchoMode(huh.EchoModePassword).Title("Enter Client Secret:").Description("Enter the Client Secret for the application. You can find it from the 'protocol' tab of the application.")

var IdentityServerURLPrompt = huh.NewInput().Title("Enter Identity Server URL:").Description("Enter the URL of the Identity Server in the format [<protocol>://<host>] (Example https://localhost:9443)").Validate(ValidateIdentityServerURL)
