// Package http provides HTTP client utilities and retry logic
package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient interface for making HTTP requests (allows for mocking in tests)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client wraps the standard HTTP client with retry logic and timeout handling
type Client struct {
	httpClient HTTPClient
	timeout    time.Duration
	maxRetries int
}

// NewClient creates a new HTTP client with the specified configuration
func NewClient(timeout time.Duration, maxRetries int) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

// NewClientWithHTTPClient creates a new HTTP client with a custom HTTP client
func NewClientWithHTTPClient(httpClient HTTPClient, timeout time.Duration, maxRetries int) *Client {
	return &Client{
		httpClient: httpClient,
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

// Post makes a POST request with retry logic
func (c *Client) Post(ctx context.Context, url string, headers map[string]string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default content type if not provided
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doWithRetry(req)
}

// Get makes a GET request with retry logic
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doWithRetry(req)
}

// doWithRetry executes the request with retry logic
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Clone the request for retry attempts
		reqClone := req.Clone(req.Context())

		// If there's a body, we need to reset it for retries
		if req.Body != nil {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %w", err)
			}
			req.Body.Close()
			reqClone.Body = io.NopCloser(bytes.NewReader(body))
			// Reset original request body for potential future retries
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		resp, err := c.httpClient.Do(reqClone)
		if err != nil {
			lastErr = err
			if attempt < c.maxRetries && c.shouldRetryError(err) {
				c.waitBeforeRetry(attempt)
				continue
			}
			return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", attempt+1, err)
		}

		// Check if we should retry based on status code
		if c.shouldRetryStatus(resp.StatusCode) && attempt < c.maxRetries {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			c.waitBeforeRetry(attempt)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// shouldRetryError determines if an error should trigger a retry
func (c *Client) shouldRetryError(err error) bool {
	// Retry on network errors, timeouts, etc.
	// This is a simplified implementation - in production you might want more sophisticated logic
	return true
}

// shouldRetryStatus determines if an HTTP status code should trigger a retry
func (c *Client) shouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case 429: // Rate limited
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	default:
		return false
	}
}

// waitBeforeRetry implements exponential backoff
func (c *Client) waitBeforeRetry(attempt int) {
	// Exponential backoff: 1s, 2s, 4s, 8s, etc.
	backoff := time.Duration(1<<uint(attempt)) * time.Second

	// Cap the backoff at 30 seconds
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}

	time.Sleep(backoff)
}
