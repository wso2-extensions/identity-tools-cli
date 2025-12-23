package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/components"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/core/utils"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

type AuthResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

var UrlPrefix string = ""

func buildTokenRequest(serverUrl, clientID, clientSecret string, body url.Values) (*http.Request, error) {
	req, err := http.NewRequest("POST", serverUrl, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)
	return req, nil
}

func parseAuthResponse(resp *http.Response) (*AuthResponse, error) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Login failed with status: " + resp.Status + " Response: " + string(data))
	}

	var authResponse AuthResponse
	if err := json.Unmarshal(data, &authResponse); err != nil {
		return nil, err
	}
	return &authResponse, nil
}

func loginAndGetToken(serverUrl string, clientID string, clientSecret string, orgName string) error {

	body := url.Values{}
	body.Set("grant_type", internal.AUTH_GRANT_TYPE)
	body.Set("scope", internal.REQUIRED_SCOPES)

	client := utils.CreateHttpClient(true)
	req, err := buildTokenRequest(serverUrl, clientID, clientSecret, body)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	authResponse, err := parseAuthResponse(resp)
	if err != nil {
		return err
	}

	log.Println(components.StylizeSuccessMessage("Logged in Successfully!"))
	utils.StoretoKeyring(internal.ACCESS_TOKEN_KEY, authResponse.AccessToken)
	utils.StoretoKeyring(internal.CLIENT_SECRET_KEY, clientSecret)

	utils.SetConfigValue(internal.ORG_NAME_KEY, orgName)
	utils.SetConfigValue(internal.CLIENT_ID_KEY, clientID)
	utils.SetConfigValue(internal.SERVER_URL_KEY, UrlPrefix)

	return nil

}
func loginToAsgardeo(clientID string, clientSecret string, orgName string) error {
	serverUrl := internal.ASGARDEO_URL_PREFIX + orgName + internal.AUTH_TOKEN_ENDPOINT
	UrlPrefix = internal.ASGARDEO_URL_PREFIX + orgName + "/"
	return loginAndGetToken(serverUrl, clientID, clientSecret, orgName)
}

func loginToIS(clientID string, clientSecret string, orgName string, serverUrl string) error {
	fullURL := serverUrl + internal.AUTH_TOKEN_ENDPOINT
	UrlPrefix = serverUrl + "/" + "t/" + orgName + "/"
	return loginAndGetToken(fullURL, clientID, clientSecret, orgName)
}

func Login(server string, clientID string, clientSecret string, orgName string, serverUrl string) error {
	switch server {
	case internal.ASGARDEO:
		return loginToAsgardeo(clientID, clientSecret, orgName)
	case internal.IS:
		return loginToIS(clientID, clientSecret, orgName, serverUrl)
	default:
		return errors.New("Invalid Server Type (" + server + ") provided. Supported Server Types are Asgardeo and Identity Server")
	}
}
