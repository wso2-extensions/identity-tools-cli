package organizations

import (
	"fmt"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

type organization struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func GetSuperOrganizationId() (id string, err error) {

	org, err := utils.SendGetRequest(utils.ORGANIZATIONS, "self")
	if err != nil {
		return "", fmt.Errorf("error while getting organization: %w", err)
	}

	var superOrg organization
	if _, err := utils.Deserialize(org, utils.FormatJSON, utils.ORGANIZATIONS, &superOrg); err != nil {
		return "", fmt.Errorf("error while JSON response: %w", err)
	}
	return superOrg.Id, nil
}
