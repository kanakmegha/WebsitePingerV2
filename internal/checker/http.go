package checker

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"
)

type HTTPResult struct {
	StatusCode        int           `json:"status_code"`
	ResponseTimeMS    int64         `json:"response_time_ms"`
	ResponseSizeBytes int64         `json:"response_size_bytes"`
	FinalURL          string        `json:"final_url"`
	Error             string        `json:"error,omitempty"`
	Success           bool          `json:"success"`
}

func executeSingleHTTPCheck(ctx context.Context, targetURL string, timeoutSec int) *HTTPResult {
	if timeoutSec <= 0 {
		timeoutSec = 10
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return &HTTPResult{Success: false, Error: err.Error()}
	}

	req.Header.Set("User-Agent", "Pinger-Monitor-Bot/1.0 (+https://pinger.io)")

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &HTTPResult{
			Success:        false,
			ResponseTimeMS: latency,
			Error:          err.Error(),
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	bodySize := int64(len(bodyBytes))
	if err != nil && err != io.EOF {
		bodySize = 0
	}

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 400

	return &HTTPResult{
		StatusCode:        resp.StatusCode,
		ResponseTimeMS:    latency,
		ResponseSizeBytes: bodySize,
		FinalURL:          resp.Request.URL.String(),
		Success:           isSuccess,
	}
}

// CheckHTTP performs HTTP ping check with up to 3 retry attempts using exponential backoff (1s, 2s, 4s).
func CheckHTTP(ctx context.Context, targetURL string, timeoutSec int) (*HTTPResult, error) {
	maxAttempts := 3
	backoffDelays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	var lastResult *HTTPResult

	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastResult = executeSingleHTTPCheck(ctx, targetURL, timeoutSec)

		// If check returned successful status (2xx-3xx), stop immediately & return success
		if lastResult.Success {
			return lastResult, nil
		}

		// If failed and attempts remain, wait for backoff delay (1s, 2s, 4s)
		if attempt < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return lastResult, ctx.Err()
			case <-time.After(backoffDelays[attempt]):
			}
		}
	}

	// Returns final official down status after 3 failed attempts
	return lastResult, nil
}
