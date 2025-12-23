package api

import (
	"errors"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/utils"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

func Logout() error {
	if err := utils.DeletefromKeyring(internal.ACCESS_TOKEN_KEY); err != nil {
		if err.Error() == internal.KEYRING_ITEM_NOT_FOUND_ERROR {
			return errors.New(internal.REPEATED_LOGOUT_ERROR)
		}
		return err
	}
	if err := utils.DeletefromKeyring(internal.CLIENT_ID_KEY); err != nil {
		return err
	}
	if err := utils.DeletefromKeyring(internal.CLIENT_SECRET_KEY); err != nil {
		return err
	}
	if err := utils.DeletefromKeyring(internal.ORG_NAME_KEY); err != nil {
		return err
	}
	if err := utils.DeletefromKeyring(internal.SERVER_URL_KEY); err != nil {
		return err
	}
	return nil

}
