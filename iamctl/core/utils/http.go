package utils

import (
	"crypto/tls"
	"net/http"
)

func CreateHttpClient(skipSSLVerification bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerification},
	}
	return &http.Client{Transport: tr}
}
