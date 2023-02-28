package cmd

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var SERVER_CONFIGS ServerConfigs

var exportAllCmd = &cobra.Command{
	Use:   "exportAll",
	Short: "You can export all service providers",
	Long:  `You can export all service providers`,
	Run: func(cmd *cobra.Command, args []string) {
		outputDirPath, _ := cmd.Flags().GetString("outputDir")
		format, _ := cmd.Flags().GetString("format")
		configFile, _ := cmd.Flags().GetString("config")

		SERVER_CONFIGS = loadServerConfigsFromFile(configFile)
		exportAllApps(outputDirPath, format)
	},
}

func init() {

	rootCmd.AddCommand(exportAllCmd)
	exportAllCmd.Flags().StringP("outputDir", "o", "", "Path to the output directory")
	exportAllCmd.Flags().StringP("format", "f", "yaml", "Format of the exported files")
	exportAllCmd.Flags().StringP("config", "c", "", "Path to the config file")
}

func exportAllApps(outputDirPath string, format string) {

	var exportFilePath = "."
	if outputDirPath != "" {
		exportFilePath = outputDirPath
	}
	exportFilePath = exportFilePath + "/Applications/"
	os.MkdirAll(exportFilePath, 0700)

	apps := getAppList()
	for _, app := range apps {
		log.Println("Exporting app: " + app.Name)
		exportApp(app.Id, exportFilePath, format)
	}
}

func exportApp(appId string, outputDirPath string, format string) {

	var fileType = "application/yaml"
	if format == "json" {
		fileType = "application/json"
	} else if format == "xml" {
		fileType = "application/xml"
	}

	var APPURL = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications"
	var err error
	var reqUrl = APPURL + "/" + appId + "/exportFile"

	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
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

	statusCode := resp.StatusCode
	switch statusCode {
	case 401:
		log.Println("Unauthorized access.\nPlease check your Username and password.")
	case 400:
		log.Println("Provided parameters are not in correct format.")
	case 403:
		log.Println("Forbidden request.")
	case 404:
		log.Println("Service Provider not found for the given ID.")
	case 500:
		log.Println("Internal server error.")
	case 200:
		var attachmentDetail = resp.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(attachmentDetail)
		if err != nil {
			log.Println("Error while parsing the content disposition header")
			panic(err)
		}

		var fileName = params["filename"]

		body1, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		exportedFile := outputDirPath + fileName
		ioutil.WriteFile(exportedFile, body1, 0644)
		log.Println("Successfully created the export file : " + exportedFile)
	}
}
