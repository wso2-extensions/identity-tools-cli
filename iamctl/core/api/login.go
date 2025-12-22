package api

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/internal"
)

type AuthResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

func loginAndGetToken(serverUrl string, clientID string, clientSecret string, orgName string) error {

	body := url.Values{}
	body.Set("grant_type", internal.AUTH_GRANT_TYPE)
	body.Set("scope", internal.REQUIRED_SCOPES)

	client := &http.Client{}
	req, err := http.NewRequest("POST", serverUrl, strings.NewReader(body.Encode()))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	// Skip SSL verification for self-signed certificates (for development purposes only)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.Transport = tr

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Login failed with status: " + resp.Status + " Response: " + string(data))
	}

	var authResponse AuthResponse
	if err := json.Unmarshal(data, &authResponse); err != nil {
		return err
	}
	fmt.Println("Login successful!")
	fmt.Println("Access Token:", authResponse.AccessToken)
	return nil

}
func loginToAsgardeo(clientID string, clientSecret string, orgName string) error {
	serverUrl := internal.ASGARDEO_URL_PREFIX + orgName + internal.AUTH_TOKEN_ENDPOINT
	return loginAndGetToken(serverUrl, clientID, clientSecret, orgName)
}

func loginToIS(clientID string, clientSecret string, orgName string, serverUrl string) error {
	fullURL := serverUrl + internal.AUTH_TOKEN_ENDPOINT
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
