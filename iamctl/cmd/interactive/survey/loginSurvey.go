package survey

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

var serverQuestion = &survey.Question{
	Name: "serverType",
	Prompt: &survey.Select{
		Message: "Select the server type:",
		Options: []string{internal.ASGARDEO, internal.IS},
	},
	Validate: survey.Required,
}
var ServerPrompt = []*survey.Question{serverQuestion}

// var ClientIDPrompt = []*survey.Question{clientIDQuestion}
var ClientIDPrompt = &survey.Input{
	Message: "Enter Client ID:",
	Help:    "Enter the Client ID for the application. You can find it from the 'protocol' tab of the application",
}

var OrgNamePrompt = &survey.Input{
	Message: "Enter Organization Name:",
	Help:    "Enter the Organization Name (Asgardeo) or Tenant Domain (Identity Server). You can find it in the console url of the server",
}

var ClientSecretPrompt = &survey.Password{
	Message: "Enter Client Secret:",
	Help:    "Enter the Client Secret for the application. You can find it from the 'protocol' tab of the application",
}

var IdentityServerURLPrompt = &survey.Input{
	Message: "Enter Identity Server URL: ",
	Help:    "Enter the URL of the Identity Server in the format [<protocol>://<host>] (Example https://localhost:9443)",
}
