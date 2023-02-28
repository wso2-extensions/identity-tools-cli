package cmd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var importAllCmd = &cobra.Command{
	Use:   "importAll",
	Short: "You can import all service providers",
	Long:  `You can import all service providers`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDirPath, _ := cmd.Flags().GetString("inputDir")
		configFile, _ := cmd.Flags().GetString("config")

		SERVER_CONFIGS = loadServerConfigsFromFile(configFile)
		importAllApps(inputDirPath)
	},
}

func init() {

	rootCmd.AddCommand(importAllCmd)
	importAllCmd.Flags().StringP("inputDir", "i", "", "Path to the input directory")
	importAllCmd.Flags().StringP("config", "c", "", "Path to the config file")
}

type AppConfig struct {
	ApplicationName string `yaml:"applicationName"`
}

func importAllApps(inputDirPath string) {

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
	appList := getAppNames()

	for _, file := range files {
		appFilePath = importFilePath + file.Name()
		fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		// Read the content of the file.
		fileContent, err := ioutil.ReadFile(appFilePath)
		if err != nil {
			log.Fatal(err)
		}

		// Parse the YAML content.
		var appConfig AppConfig
		err = yaml.Unmarshal(fileContent, &appConfig)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(appConfig.ApplicationName)

		// Check if app exists.
		var appExists bool
		for _, app := range appList {
			if app == appConfig.ApplicationName {
				appExists = true
				break
			}
		}

		if appConfig.ApplicationName != fileName {
			log.Println("Application name in the file " + appFilePath + " is not matching with the file name.")
			// log.Println("Renaming file name to " + appConfig.ApplicationName + ".yml")
			// err := os.Rename(appFilePath, importFilePath+appConfig.ApplicationName+".yml")
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// appFilePath = importFilePath + appConfig.ApplicationName + ".yml"
		}

		importApp(appFilePath, appExists)
	}
}

func importApp(importFilePath string, update bool) {

	var reqUrl = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications/import"
	var err error

	// fileBytes, err := ioutil.ReadFile(importFilePath)
	file, err := os.Open(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	filename := filepath.Base(importFilePath)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Get file extension
	fileExtension := filepath.Ext(filename)

	mime.AddExtensionType(".yml", "text/yaml")
	mime.AddExtensionType(".xml", "application/xml")

	mimeType := mime.TypeByExtension(fileExtension)

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal(err)
	}

	defer writer.Close()

	var requestMethod string
	if update {
		log.Println("Updating app: " + filename)
		requestMethod = "PUT"
	} else {
		log.Println("Importing app: " + filename)
		requestMethod = "POST"
	}

	request, err := http.NewRequest(requestMethod, reqUrl, body)
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
		log.Println("An application with the same name already exists. Please rename the file accordingly.")
		importApp(importFilePath, true)
	case 500:
		log.Println("Internal server error.")
	case 201:
		log.Println("Application imported successfully.")
	}
}
