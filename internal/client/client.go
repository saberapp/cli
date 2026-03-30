package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout         = 60 * time.Second
	maxRetries             = 3
	rateLimitWarnThreshold = 10
)

// Client wraps http.Client with auth, retry, and verbose logging.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	version    string
	verbose    bool
	stderr     io.Writer
}

// New creates a new API client.
func New(baseURL, apiKey, version string, verbose bool, stderr io.Writer) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		version:    version,
		verbose:    verbose,
		stderr:     stderr,
	}
}

// APIError represents an error response from the API.
type APIError struct {
	StatusCode int
	Message    string
	RetryAfter int // seconds, for 429
}

func (e *APIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("API error %d: %s (retry after %ds)", e.StatusCode, e.Message, e.RetryAfter)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// errorBody handles both response formats:
// - Go Platform: {"error":{"type":"...","code":"...","message":"...","requestId":"..."}}
// - Rate limiter: {"statusCode":429,"message":"...","retryAfter":<seconds>}
type errorBody struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retryAfter"`
	Error      *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (e *errorBody) message() string {
	if e.Error != nil && e.Error.Message != "" {
		return e.Error.Message
	}
	if e.Message != "" {
		return e.Message
	}
	return ""
}

// Do executes an HTTP request with retry/backoff logic.
// bodyBytes may be nil for requests without a body.
// If rawDst is non-nil, the response body is copied directly to it (for --json mode).
// Otherwise, the response body is JSON-decoded into dst.
func (c *Client) Do(ctx context.Context, method, path string, bodyBytes []byte, extraHeaders map[string]string, dst any, rawDst io.Writer) error {
	url := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter: 1s, 2s ± up to 500ms
			base := time.Duration(1<<(attempt-1)) * time.Second
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond
			select {
			case <-time.After(base + jitter):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("User-Agent", "saber-cli/"+c.version)
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}

		if c.verbose {
			maskedKey := MaskKey(c.apiKey)
			c.logf("> %s %s\n", method, url)
			c.logf("> Authorization: Bearer %s\n", maskedKey)
			for k, v := range extraHeaders {
				c.logf("> %s: %s\n", k, v)
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if c.verbose {
			c.logf("< HTTP %d\n", resp.StatusCode)
			if v := resp.Header.Get("X-Ratelimit-Limit"); v != "" {
				c.logf("< X-Ratelimit-Limit: %s\n", v)
				c.logf("< X-Ratelimit-Remaining: %s\n", resp.Header.Get("X-Ratelimit-Remaining"))
				c.logf("< X-Ratelimit-Reset: %s\n", resp.Header.Get("X-Ratelimit-Reset"))
			}
		}

		// Warn on low rate limit
		if remaining := resp.Header.Get("X-Ratelimit-Remaining"); remaining != "" {
			if r, err := strconv.Atoi(remaining); err == nil && r <= rateLimitWarnThreshold {
				c.logf("warning: only %d API requests remaining in this window\n", r)
			}
		}

		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			var eb errorBody
			_ = json.Unmarshal(respBody, &eb)
			waitSecs := eb.RetryAfter
			if waitSecs <= 0 {
				waitSecs = 60
			}
			c.logf("rate limited; waiting %ds...\n", waitSecs)
			select {
			case <-time.After(time.Duration(waitSecs) * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
			lastErr = &APIError{StatusCode: 429, Message: eb.message(), RetryAfter: waitSecs}
			continue
		}

		if resp.StatusCode >= 500 {
			var eb errorBody
			_ = json.Unmarshal(respBody, &eb)
			msg := eb.message()
			if msg == "" {
				msg = http.StatusText(resp.StatusCode)
			}
			lastErr = &APIError{StatusCode: resp.StatusCode, Message: msg}
			continue
		}

		if resp.StatusCode >= 400 {
			var eb errorBody
			_ = json.Unmarshal(respBody, &eb)
			msg := eb.message()
			if msg == "" {
				msg = http.StatusText(resp.StatusCode)
			}
			return &APIError{StatusCode: resp.StatusCode, Message: msg}
		}

		// Success — no content
		if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
			return nil
		}

		// Raw copy mode (--json)
		if rawDst != nil {
			_, err = io.Copy(rawDst, bytes.NewReader(respBody))
			return err
		}

		if dst != nil {
			if err := json.Unmarshal(respBody, dst); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}
		return nil
	}
	return lastErr
}

// Get is a convenience wrapper for GET requests.
func (c *Client) Get(ctx context.Context, path string, dst any, rawDst io.Writer) error {
	return c.Do(ctx, http.MethodGet, path, nil, nil, dst, rawDst)
}

// Post is a convenience wrapper for POST requests.
func (c *Client) Post(ctx context.Context, path string, body any, dst any, rawDst io.Writer) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}
	return c.Do(ctx, http.MethodPost, path, bodyBytes, nil, dst, rawDst)
}

// PostWithHeaders is like Post but allows passing extra request headers.
func (c *Client) PostWithHeaders(ctx context.Context, path string, headers map[string]string, body any, dst any, rawDst io.Writer) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}
	return c.Do(ctx, http.MethodPost, path, bodyBytes, headers, dst, rawDst)
}

// Put is a convenience wrapper for PUT requests.
func (c *Client) Put(ctx context.Context, path string, body any, dst any, rawDst io.Writer) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}
	return c.Do(ctx, http.MethodPut, path, bodyBytes, nil, dst, rawDst)
}

// Patch is a convenience wrapper for PATCH requests.
func (c *Client) Patch(ctx context.Context, path string, body any, dst any, rawDst io.Writer) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}
	return c.Do(ctx, http.MethodPatch, path, bodyBytes, nil, dst, rawDst)
}

// Delete is a convenience wrapper for DELETE requests.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil, nil)
}

// logf writes a formatted message to c.stderr, ignoring write errors.
// stderr writes in a CLI tool cannot meaningfully fail, so the error is intentionally discarded.
func (c *Client) logf(format string, args ...any) {
	fmt.Fprintf(c.stderr, format, args...)
}

// MaskKey returns a masked version of an API key for logging and display.
func MaskKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:8] + strings.Repeat("*", len(key)-12) + key[len(key)-4:]
}
