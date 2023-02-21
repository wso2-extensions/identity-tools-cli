package cmd

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/spf13/cobra"
)

var importAllCmd = &cobra.Command{
	Use:   "importAll",
	Short: "You can import all service providers",
	Long:  `You can import all service providers`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDirPath, _ := cmd.Flags().GetString("inputDir")
		format, _ := cmd.Flags().GetString("format")
		configFile, _ := cmd.Flags().GetString("config")

		SERVER_CONFIGS = loadServerConfigsFromFile(configFile)
		importAllApps(inputDirPath, format)
	},
}

func init() {

	rootCmd.AddCommand(importAllCmd)
	importAllCmd.Flags().StringP("inputDir", "i", "", "Path to the input directory")
	importAllCmd.Flags().StringP("format", "f", "yaml", "Format of the imported files")
	importAllCmd.Flags().StringP("config", "c", "", "Path to the config file")
}

func importAllApps(inputDirPath string, format string) {

	var importFilePath = "."
	if inputDirPath != "" {
		importFilePath = inputDirPath
	}
	importFilePath = importFilePath + "/Applications/"

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var appFilePath string
	for _, file := range files {
		appFilePath = importFilePath + file.Name()
		log.Println("Importing app: " + file.Name())
		importApp(appFilePath, format)
	}
}

func importApp(importFilePath string, format string) {

	var reqUrl = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications/import/" + format
	var err error

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	extraParams := map[string]string{
		"file": string(fileBytes),
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range extraParams {
		err := writer.WriteField(key, val)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer writer.Close()

	request, err := http.NewRequest("POST", reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	statusCode := resp.StatusCode
	switch statusCode {
	case 401:
		log.Println("Unauthorized access.\nPlease check your Username and password.")
	case 400:
		log.Println("Provided parameters are not in correct format.")
	case 403:
		log.Println("Forbidden request.")
	case 409:
		log.Println("An application with the same name already exists.")
	case 500:
		log.Println("Internal server error.")
	case 201:
		log.Println("Application imported successfully.")
	}
}
