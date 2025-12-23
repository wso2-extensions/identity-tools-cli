package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/utils"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

func calculteTimeDifference(loginTime string) (string, error) {
	current := time.Now()
	layout := time.RFC3339
	parsedLoginTime, err := time.Parse(layout, loginTime)
	if err != nil {
		return "", err
	}
	diff := current.Sub(parsedLoginTime)
	return diff.String(), nil
}
func CheckToken() error {
	token, err := utils.GetfromKeyring(internal.ACCESS_TOKEN_KEY)
	if err != nil || token == "" {
		return errors.New("Token not found, Please login again")
	}
	lastLoginTime, err := utils.GetConfigValue(internal.LAST_LOGIN_KEY)
	if err != nil || lastLoginTime == "" {
		return errors.New("Token data tampered, Please login again")
	}
	timeDiff, err := calculteTimeDifference(lastLoginTime)
	if err != nil {
		return err
	}

	expiryDurationStr, err := utils.GetConfigValue(internal.TIME_REMAINING_KEY)
	if err != nil || expiryDurationStr == "" {
		return errors.New("Token expiry information not found, Please login again")
	}

	expiryDuration, err := time.ParseDuration(expiryDurationStr + "s")
	if err != nil {
		return err
	}
	timeDiffDuration, err := time.ParseDuration(timeDiff)
	if err != nil {
		return err
	}
	if timeDiffDuration >= expiryDuration {
		fmt.Println("Access token expired, refreshing token...")
		return refreshToken()
	}
	return nil
}

func refreshToken() error {
	clientID, err := utils.GetConfigValue(internal.CLIENT_ID_KEY)
	if err != nil || clientID == "" {
		return errors.New("Client ID not found, Please login again")
	}
	clientSecret, err := utils.GetfromKeyring(internal.CLIENT_SECRET_KEY)
	if err != nil || clientSecret == "" {
		return errors.New("Client Secret not found, Please login again")
	}
	orgName, err := utils.GetConfigValue(internal.ORG_NAME_KEY)
	if err != nil || orgName == "" {
		return errors.New("Organization Name/Tenant Domain not found, Please login again")
	}
	authURL, err := utils.GetConfigValue(internal.AUTH_URL_KEY)
	if err != nil || authURL == "" {
		return errors.New("Auth URL not found, Please login again")
	}
	prefixUrl, err := utils.GetConfigValue(internal.PREFIX_URL_KEY)
	if err != nil || prefixUrl == "" {
		return errors.New("Prefix URL not found, Please login again")
	}
	return loginAndGetToken(authURL, clientID, clientSecret, orgName, prefixUrl)
}
