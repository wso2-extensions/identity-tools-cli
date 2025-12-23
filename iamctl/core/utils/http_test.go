package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateHttpClient(t *testing.T) {
	t.Run("Should return successful client with correct TLS settings", func(t *testing.T) {
		client := CreateHttpClient(false)
		transport, ok := client.Transport.(*http.Transport)
		assert.True(t, ok, "Transport should be of type *http.Transport")
		assert.False(t, transport.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify should be false")

		client = CreateHttpClient(true)
		transport, ok = client.Transport.(*http.Transport)
		assert.True(t, ok, "Transport should be of type *http.Transport")
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify should be true")
	})

}
