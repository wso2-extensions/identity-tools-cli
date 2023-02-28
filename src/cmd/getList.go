package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
)

func getAppList() (spIdList []Application) {

	var APPURL = SERVER_CONFIGS.ServerUrl + "/t/" + SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications"
	var list List

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, _ := http.NewRequest("GET", APPURL, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+SERVER_CONFIGS.Token)
	req.Header.Set("accept", "*/*")
	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer writer.Flush()

	err = json.Unmarshal(body, &list)
	if err != nil {
		log.Fatalln(err)
	}
	resp.Body.Close()

	return list.Applications
}

func getAppNames() []string {

	// Get the list of applications.
	apps := getAppList()

	// Extract application names from the list.
	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}

	return appNames
}
