package tests

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

// newTestRateLimitConfig returns a rate limit config with short backoffs, to keep the tests fast.
func newTestRateLimitConfig(maxRetries int) utils.RateLimitConfig {

	config := utils.DefaultRateLimitConfig()
	config.MaxRetries = maxRetries
	config.InitialBackoffSecs = 0.01
	config.MaxBackoffSecs = 0.02
	return config
}

func withRateLimitConfig(config utils.RateLimitConfig, testFunc func()) {

	previous := utils.TOOL_CONFIGS.RateLimit
	utils.TOOL_CONFIGS.RateLimit = config
	defer func() { utils.TOOL_CONFIGS.RateLimit = previous }()
	testFunc()
}

func TestSendHTTPRequestRetriesOnRateLimit(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&requestCount, 1) < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	withRateLimitConfig(newTestRateLimitConfig(5), func() {
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatalf("error creating the request: %s", err)
		}
		resp, err := utils.SendHTTPRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code 200, got %d", resp.StatusCode)
		}
		if count := atomic.LoadInt32(&requestCount); count != 3 {
			t.Errorf("expected 3 requests, got %d", count)
		}
	})
}

func TestSendHTTPRequestStopsAfterMaxRetries(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	withRateLimitConfig(newTestRateLimitConfig(2), func() {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := utils.SendHTTPRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("expected status code 429, got %d", resp.StatusCode)
		}
		// One initial attempt and two retries.
		if count := atomic.LoadInt32(&requestCount); count != 3 {
			t.Errorf("expected 3 requests, got %d", count)
		}
	})
}

func TestSendHTTPRequestDoesNotRetryWhenDisabled(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	config := newTestRateLimitConfig(5)
	config.Enabled = false
	withRateLimitConfig(config, func() {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := utils.SendHTTPRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		defer resp.Body.Close()

		if count := atomic.LoadInt32(&requestCount); count != 1 {
			t.Errorf("expected 1 request when rate limit handling is disabled, got %d", count)
		}
	})
}

func TestSendHTTPRequestReplaysRequestBody(t *testing.T) {

	var requestCount int32
	receivedBodies := make(chan string, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		receivedBodies <- string(body)
		if atomic.AddInt32(&requestCount, 1) < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	withRateLimitConfig(newTestRateLimitConfig(3), func() {
		req, _ := http.NewRequest("POST", server.URL, strings.NewReader("test-payload"))
		resp, err := utils.SendHTTPRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		defer resp.Body.Close()
		close(receivedBodies)

		count := 0
		for body := range receivedBodies {
			count++
			if body != "test-payload" {
				t.Errorf("expected the request body to be replayed, got %q", body)
			}
		}
		if count != 2 {
			t.Errorf("expected 2 requests, got %d", count)
		}
	})
}

func TestSendHTTPRequestDoesNotRetryServerErrors(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	withRateLimitConfig(newTestRateLimitConfig(3), func() {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := utils.SendHTTPRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		defer resp.Body.Close()

		if count := atomic.LoadInt32(&requestCount); count != 1 {
			t.Errorf("expected only rate limited requests to be retried, got %d requests", count)
		}
	})
}

func TestSendHTTPRequestAbortsOnLongRetryAfter(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	config := newTestRateLimitConfig(3)
	config.MaxRetryAfterSeconds = 10
	withRateLimitConfig(config, func() {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := utils.SendHTTPRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		defer resp.Body.Close()

		if count := atomic.LoadInt32(&requestCount); count != 1 {
			t.Errorf("expected retrying to be aborted when Retry-After exceeds the configured maximum, got %d requests", count)
		}
	})
}
