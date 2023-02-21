package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "You can export a service provider",
	Long:  `You can export a service provider`,
	Run: func(cmd *cobra.Command, args []string) {
		setExportInfo()
	},
}

func init() {

	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("serviceProviderID", "s", "", "set the service provide ID")
	exportCmd.Flags().StringP("exportlocation", "p", "", "set the export location")
	exportCmd.Flags().StringP("fileType", "t", "", "set the file type")
}

type applicationsStruct struct {
	TotalResults int `json:"totalResults"`
	StartIndex   int `json:"startIndex"`
	Count        int `json:"count"`
	Applications []struct {
		appSummary
	} `json:"applications"`
	Links []interface{} `json:"links"`
}

type appSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
	AccessURL   string `json:"accessUrl"`
	Access      string `json:"access"`
	Self        string `json:"self"`
}

var exportQuestions = []*survey.Question{
	{
		Name:     "exportlocation",
		Prompt:   &survey.Input{Message: "Enter export location : "},
		Validate: survey.Required,
	},
	{
		Name:     "serviceProviderID",
		Prompt:   &survey.Input{Message: "Enter service provider id to be exported :"},
		Validate: survey.Required,
	},
	{
		Name:     "fileType",
		Prompt:   &survey.Input{Message: "Enter file type i.e application/json or application/yaml :"},
		Validate: survey.Required,
	},
}

func setExportInfo() {

	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render(appName)
	fmt.Print(renderStr)

	exportAnswers := struct {
		Exportlocation    string `survey:"exportlocation"`
		ServiceProviderID string `survey:"serviceProviderID"`
		FileType          string `survey:"fileType"`
	}{}

	err := survey.Ask(exportQuestions, &exportAnswers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	exportApplication(exportAnswers.Exportlocation, exportAnswers.ServiceProviderID, exportAnswers.FileType)
}

func retreiveApplications(reqUrl string) bool {

	var applications applicationsStruct

	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Client ZGFzaGJvYXJkOmRhc2hib2FyZA==")
	defer req.Body.Close()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode == 401 {
		type clientError struct {
			Description string `json:"error_description"`
			Error       string `json:"error"`
		}
		var err = new(clientError)

		err2 := json.Unmarshal(body1, &err)
		if err2 != nil {
			log.Fatalln(err2)
		}
		fmt.Println(err.Error + "\n" + err.Description)
		return false
	}

	err2 := json.Unmarshal(body1, &applications)
	if err2 != nil {
		log.Fatalln(err2)
	}

	if body1 != nil {

		appSummaryList := applications.Applications
		fmt.Println(string(appSummaryList[0].Name))
	}

	return true
}

func exportApplication(exportlocation string, serviceProviderID string, fileType string) bool {

	exported := false

	SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()

	start(SERVER, "admin", "admin")

	var ADDAPPURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications"
	var err error

	token := readFile()

	var reqUrl = ADDAPPURL + "/" + serviceProviderID + "/exportFile"
	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+token)
	defer req.Body.Close()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var attachmentDetail = resp.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(attachmentDetail)

	if err != nil {
		panic(err)
	}
	log.Println("Disposition" + disposition)

	var fileName = params["filename"]

	body1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if body1 != nil {
		exported = true
	}

	var exportedFilePath = exportlocation + "/" + fileName
	ioutil.WriteFile(exportedFilePath, body1, 0644)
	print("Successfully created the export file : " + exportedFilePath)

	return exported
}

func writeToYMLFile() {

}
