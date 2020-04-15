/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Parts struct {
	GrantTypes   []string `json:"grantTypes"`
	CallbackURLs []string `json:"callbackURLs"`
	PublicClient bool     `json:"publicClient"`
}
type Parts1 struct {
	Oidc Parts `json:"oidc"`
}
type ServiceProviderOAuth struct {
	Name                         string `json:"name"`
	Description                  string `json:"description"`
	InboundProtocolConfiguration Parts1 `json:"inboundProtocolConfiguration"`
}
type ServiceProviderXml struct {
	XMLName                     xml.Name `xml:"ServiceProvider"`
	Text                        string   `xml:",chardata"`
	ApplicationName             string   `xml:"ApplicationName"`
	Description                 string   `xml:"Description"`
	JwksUri                     string   `xml:"JwksUri"`
	InboundAuthenticationConfig struct {
		Text                                string `xml:",chardata"`
		InboundAuthenticationRequestConfigs struct {
			Text                               string `xml:",chardata"`
			InboundAuthenticationRequestConfig []struct {
				Text                 string `xml:",chardata"`
				InboundAuthKey       string `xml:"InboundAuthKey"`
				InboundAuthType      string `xml:"InboundAuthType"`
				InboundConfigType    string `xml:"InboundConfigType"`
				Properties           string `xml:"Properties"`
				InboundConfiguration string `xml:"inboundConfiguration"`
			} `xml:"InboundAuthenticationRequestConfig"`
		} `xml:"InboundAuthenticationRequestConfigs"`
	} `xml:"InboundAuthenticationConfig"`
	LocalAndOutBoundAuthenticationConfig struct {
		Text                                  string `xml:",chardata"`
		AuthenticationSteps                   string `xml:"AuthenticationSteps"`
		AuthenticationType                    string `xml:"AuthenticationType"`
		AlwaysSendBackAuthenticatedListOfIdPs string `xml:"alwaysSendBackAuthenticatedListOfIdPs"`
		UseTenantDomainInUsername             string `xml:"UseTenantDomainInUsername"`
		UseUserstoreDomainInRoles             string `xml:"UseUserstoreDomainInRoles"`
		UseUserstoreDomainInUsername          string `xml:"UseUserstoreDomainInUsername"`
		SkipConsent                           string `xml:"SkipConsent"`
		SkipLogoutConsent                     string `xml:"skipLogoutConsent"`
		EnableAuthorization                   string `xml:"EnableAuthorization"`
	} `xml:"LocalAndOutBoundAuthenticationConfig"`
	RequestPathAuthenticatorConfigs string `xml:"RequestPathAuthenticatorConfigs"`
	InboundProvisioningConfig       struct {
		Text                  string `xml:",chardata"`
		ProvisioningUserStore string `xml:"ProvisioningUserStore"`
		IsProvisioningEnabled string `xml:"IsProvisioningEnabled"`
		IsDumbModeEnabled     string `xml:"IsDumbModeEnabled"`
	} `xml:"InboundProvisioningConfig"`
	OutboundProvisioningConfig struct {
		Text                          string `xml:",chardata"`
		ProvisioningIdentityProviders string `xml:"ProvisioningIdentityProviders"`
	} `xml:"OutboundProvisioningConfig"`
	ClaimConfig struct {
		Text                           string `xml:",chardata"`
		RoleClaimURI                   string `xml:"RoleClaimURI"`
		LocalClaimDialect              string `xml:"LocalClaimDialect"`
		IdpClaim                       string `xml:"IdpClaim"`
		ClaimMappings                  string `xml:"ClaimMappings"`
		AlwaysSendMappedLocalSubjectId string `xml:"AlwaysSendMappedLocalSubjectId"`
		SPClaimDialects                string `xml:"SPClaimDialects"`
	} `xml:"ClaimConfig"`
	PermissionAndRoleConfig struct {
		Text         string `xml:",chardata"`
		Permissions  string `xml:"Permissions"`
		RoleMappings string `xml:"RoleMappings"`
		IdpRoles     string `xml:"IdpRoles"`
	} `xml:"PermissionAndRoleConfig"`
	IsSaaSApp      string `xml:"IsSaaSApp"`
	ImageUrl       string `xml:"ImageUrl"`
	AccessUrl      string `xml:"AccessUrl"`
	IsDiscoverable string `xml:"IsDiscoverable"`
}
type Export struct {
	ApplicationID string `json:"applicationId"`
}

func createSPOauthApplication(oauthAppName string, description string, callbackURLs string, grantTypes []string) {

	SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()

	var ADDAPPURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications"
	var err error
	var status int
	var xmlData ServiceProviderXml

	token := readFile()

	toJson := ServiceProviderOAuth{
		Name:        oauthAppName,
		Description: description,
		InboundProtocolConfiguration: Parts1{
			Parts{
				grantTypes,
				[]string{callbackURLs},
				false,
			},
		},
	}
	jsonData, err := json.Marshal(toJson)
	if err != nil {
		log.Fatalln(err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("POST", ADDAPPURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	status = resp.StatusCode
	defer resp.Body.Close()

	if status == 401 {
		fmt.Println("Unauthorized access.\nPlease enter your UserName and password for server.")
		setServerWithInit(SERVER)
		createSPOauthApplication(oauthAppName, description, callbackURLs, grantTypes)
	} else if status == 400 {
		fmt.Println("Provided parameters are not in correct format.")
	} else if status == 403 {
		fmt.Println("Forbidden")
	} else if status == 201 {
		fmt.Println("Successfully created the service provider named '" + oauthAppName + "' at " + resp.Header.Get("Date"))
		location := resp.Header.Get("Location")

		splits := strings.SplitAfter(location, "applications/")
		serviceProviderID := splits[1]

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		req, err := http.NewRequest("GET", ADDAPPURL+"/"+serviceProviderID+"/export", bytes.NewBuffer(nil))
		query := req.URL.Query()
		query.Add("exportSecrets", "true")
		req.URL.RawQuery = query.Encode()
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("accept", "*/*")

		defer req.Body.Close()

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		err = xml.Unmarshal(body, &xmlData)
		if err != nil {
			log.Fatalln(err)
		}

		configuration := xmlData.InboundAuthenticationConfig.InboundAuthenticationRequestConfigs.InboundAuthenticationRequestConfig[0].InboundConfiguration
		fmt.Println("oauthConsumerKey: " + between(configuration, "<oauthConsumerKey>", "</oauthConsumerKey>"))
		fmt.Println("oauthConsumerSecret: " + between(configuration, "<oauthConsumerSecret>", "</oauthConsumerSecret>"))
	} else if status == 409 {
		fmt.Println("Already exists an application with same name:" + oauthAppName)
	}
}
func between(fullString string, start string, end string) string {

	// Get substring between two strings.
	posFirst := strings.Index(fullString, start)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(fullString, end)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(start)
	if posFirstAdjusted >= posLast {
		return ""
	}

	return fullString[posFirstAdjusted:posLast]
}
