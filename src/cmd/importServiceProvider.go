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

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "You can import a service provider",
	Long:  `You can import a service provider`,
	Run: func(cmd *cobra.Command, args []string) {
		importFilePath, errEXL := cmd.Flags().GetString("importFilePath")

		if errEXL != nil {
			log.Fatalln(errEXL)
		}
		importApplication(importFilePath)
	},
}

func init() {

	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringP("importFilePath", "i", "", "set the export file name")
}

func importApplication(importFilePath string) bool {
	importedSp := false

	SERVER, CLIENTID, CLIENTSECRET, TENANTDOMAIN = readSPConfig()

	start(SERVER, "admin", "admin")

	var ADDAPPURL = SERVER + "/t/" + TENANTDOMAIN + "/api/server/v1/applications/import"
	var err error

	token := readFile()

	// fileBytes, err := ioutil.ReadFile(importFilePath)
	file, err := os.Open(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	filename := filepath.Base(importFilePath)
	fmt.Println("File name:", filename)

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

	request, err := http.NewRequest("POST", ADDAPPURL, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
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
	} else {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(bodyBytes))

		importedSp = true
	}

	return importedSp
}
