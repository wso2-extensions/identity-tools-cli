/**
* Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
*
* WSO2 LLC. licenses this file to you under the Apache License,
* Version 2.0 (the "License"); you may not use this file except
* in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Default values used when the RATE_LIMIT block is not fully specified in the tool configs.
const (
	DEFAULT_RATE_LIMIT_ENABLED      = true
	DEFAULT_MAX_RETRIES             = 5
	DEFAULT_INITIAL_BACKOFF_SECONDS = 1
	DEFAULT_MAX_BACKOFF_SECONDS     = 60
	DEFAULT_BACKOFF_MULTIPLIER      = 2.0
	DEFAULT_MAX_RETRY_AFTER_SECONDS = 300
)

// RateLimitConfig holds the rate limit handling applied to all requests sent to the Identity Server.
type RateLimitConfig struct {
	Enabled              bool    `json:"ENABLED"`
	MaxRetries           int     `json:"MAX_RETRIES"`
	InitialBackoffSecs   float64 `json:"INITIAL_BACKOFF_SECONDS"`
	MaxBackoffSecs       float64 `json:"MAX_BACKOFF_SECONDS"`
	BackoffMultiplier    float64 `json:"BACKOFF_MULTIPLIER"`
	MaxRetryAfterSeconds float64 `json:"MAX_RETRY_AFTER_SECONDS"`
}

// DefaultRateLimitConfig returns the rate limit configs used when they are not overridden in the tool configs.
func DefaultRateLimitConfig() RateLimitConfig {

	return RateLimitConfig{
		Enabled:              DEFAULT_RATE_LIMIT_ENABLED,
		MaxRetries:           DEFAULT_MAX_RETRIES,
		InitialBackoffSecs:   DEFAULT_INITIAL_BACKOFF_SECONDS,
		MaxBackoffSecs:       DEFAULT_MAX_BACKOFF_SECONDS,
		BackoffMultiplier:    DEFAULT_BACKOFF_MULTIPLIER,
		MaxRetryAfterSeconds: DEFAULT_MAX_RETRY_AFTER_SECONDS,
	}
}

// normalize replaces invalid values with the defaults so that a partially filled config block is still usable.
func (c RateLimitConfig) normalize() RateLimitConfig {

	defaults := DefaultRateLimitConfig()
	if c.MaxRetries < 0 {
		c.MaxRetries = defaults.MaxRetries
	}
	if c.InitialBackoffSecs <= 0 {
		c.InitialBackoffSecs = defaults.InitialBackoffSecs
	}
	if c.MaxBackoffSecs <= 0 {
		c.MaxBackoffSecs = defaults.MaxBackoffSecs
	}
	if c.MaxBackoffSecs < c.InitialBackoffSecs {
		c.MaxBackoffSecs = c.InitialBackoffSecs
	}
	if c.BackoffMultiplier < 1 {
		c.BackoffMultiplier = defaults.BackoffMultiplier
	}
	if c.MaxRetryAfterSeconds <= 0 {
		c.MaxRetryAfterSeconds = defaults.MaxRetryAfterSeconds
	}
	return c
}

var (
	sharedHTTPClient *http.Client
	httpClientOnce   sync.Once

	jitterMutex  sync.Mutex
	jitterSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// GetHTTPClient returns the shared HTTP client used for all requests sent to the Identity Server.
func GetHTTPClient() *http.Client {

	httpClientOnce.Do(func() {
		sharedHTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})
	return sharedHTTPClient
}

// SendHTTPRequest sends the given request and retries it when the server responds with a rate limit
// (429), honouring the Retry-After header when present. All other responses are returned to the caller
// as they are. Retrying can be turned off with the RATE_LIMIT.ENABLED tool config.
func SendHTTPRequest(request *http.Request) (*http.Response, error) {

	client := GetHTTPClient()
	config := TOOL_CONFIGS.RateLimit.normalize()

	if !config.Enabled {
		return client.Do(request)
	}

	// A request body that cannot be rewound can only be sent once.
	if request.Body != nil && request.GetBody == nil {
		PrintLog(LogLevelDebug, UtilsResourceWrapper, "", fmt.Sprintf(
			"Request body of %s %s cannot be replayed. Retrying is skipped for this request.",
			request.Method, request.URL.String()))
		return client.Do(request)
	}

	var resp *http.Response
	var err error

	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			if err := rewindRequestBody(request); err != nil {
				return resp, err
			}
		}

		resp, err = client.Do(request)

		if attempt >= config.MaxRetries {
			return resp, err
		}

		waitTime, retry := getRetryWaitTime(resp, err, attempt, config)
		if !retry {
			return resp, err
		}

		if resp != nil {
			drainAndCloseBody(resp)
		}
		PrintLog(LogLevelWarn, UtilsResourceWrapper, "", fmt.Sprintf(
			"Rate limit exceeded in the server. Retrying %s %s in %s. (Attempt %d of %d)",
			request.Method, request.URL.String(),
			waitTime.Round(time.Millisecond), attempt+1, config.MaxRetries))
		time.Sleep(waitTime)
	}
}

// getRetryWaitTime decides whether the request was rate limited and how long to wait before retrying it.
func getRetryWaitTime(resp *http.Response, reqErr error, attempt int,
	config RateLimitConfig) (waitTime time.Duration, retry bool) {

	if reqErr != nil || resp == nil {
		return 0, false
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}

	if retryAfter, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
		if retryAfter.Seconds() > config.MaxRetryAfterSeconds {
			PrintLog(LogLevelWarn, UtilsResourceWrapper, "", fmt.Sprintf(
				"The server requested a retry after %s which exceeds the configured MAX_RETRY_AFTER_SECONDS of %g seconds. Retrying is aborted.",
				retryAfter, config.MaxRetryAfterSeconds))
			return 0, false
		}
		if retryAfter <= 0 {
			retryAfter = time.Duration(config.InitialBackoffSecs * float64(time.Second))
		}
		return retryAfter, true
	}
	return backoffDuration(attempt, config), true
}

// backoffDuration returns the exponentially increasing wait time for the given attempt, with jitter applied.
func backoffDuration(attempt int, config RateLimitConfig) time.Duration {

	backoff := config.InitialBackoffSecs
	for i := 0; i < attempt; i++ {
		backoff *= config.BackoffMultiplier
		if backoff >= config.MaxBackoffSecs {
			backoff = config.MaxBackoffSecs
			break
		}
	}

	// Apply jitter within the range [backoff/2, backoff) to avoid retrying all requests at the same time.
	jitterMutex.Lock()
	jitter := jitterSource.Float64()
	jitterMutex.Unlock()

	return time.Duration((backoff/2 + backoff/2*jitter) * float64(time.Second))
}

// parseRetryAfter parses the Retry-After header which can either hold a number of seconds or an HTTP date.
func parseRetryAfter(header string) (time.Duration, bool) {

	header = strings.TrimSpace(header)
	if header == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseFloat(header, 64); err == nil {
		return time.Duration(seconds * float64(time.Second)), true
	}
	if retryTime, err := http.ParseTime(header); err == nil {
		return time.Until(retryTime), true
	}
	return 0, false
}

func rewindRequestBody(request *http.Request) error {

	if request.GetBody == nil {
		return nil
	}
	body, err := request.GetBody()
	if err != nil {
		return fmt.Errorf("error rewinding the request body for a retry: %w", err)
	}
	request.Body = body
	return nil
}

// drainAndCloseBody discards the body of a response that is being retried, so that the connection can be reused.
func drainAndCloseBody(resp *http.Response) {

	if resp.Body == nil {
		return
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}
