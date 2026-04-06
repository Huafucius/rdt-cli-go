package client

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL        = "https://www.reddit.com"
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	readDelay      = 1.0 // seconds between requests
)

// fingerprint headers mimic Chrome 133 on macOS for anti-detection.
var fingerprintHeaders = map[string]string{
	"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"sec-ch-ua":        `"Chromium";v="133", "Not(A:Brand";v="99", "Google Chrome";v="133"`,
	"sec-ch-ua-mobile": "?0",
	"sec-ch-ua-platform": `"macOS"`,
	"Sec-Fetch-Dest":   "empty",
	"Sec-Fetch-Mode":   "cors",
	"Sec-Fetch-Site":   "same-origin",
	"Accept":           "application/json, text/plain, */*",
	"Accept-Language":  "en-US,en;q=0.9",
}

// Client is a Reddit API HTTP client with rate limiting and retry logic.
type Client struct {
	http          *http.Client
	lastRequestAt time.Time
}

// New creates a new Client.
func New() *Client {
	return &Client{
		http: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Get performs a GET request to a Reddit JSON endpoint and decodes the response.
func (c *Client) Get(path string, params map[string]string) (any, error) {
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid path %q: %w", path, err)
	}

	q := u.Query()
	q.Set("raw_json", "1")
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return c.doWithRetry("GET", u.String(), nil)
}

func (c *Client) doWithRetry(method, rawURL string, body io.Reader) (any, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		c.rateLimit()

		req, err := http.NewRequest(method, rawURL, body)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		for k, v := range fingerprintHeaders {
			req.Header.Set(k, v)
		}

		resp, err := c.http.Do(req)
		c.lastRequestAt = time.Now()

		if err != nil {
			lastErr = err
			c.backoff(attempt)
			continue
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			retryAfter := 5.0
			if v := resp.Header.Get("Retry-After"); v != "" {
				fmt.Sscanf(v, "%f", &retryAfter)
			}
			if attempt+1 >= maxRetries {
				return nil, fmt.Errorf("rate limited (429)")
			}
			time.Sleep(time.Duration(retryAfter * float64(time.Second)))
			continue
		case http.StatusInternalServerError, http.StatusBadGateway,
			http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			lastErr = fmt.Errorf("server error %d", resp.StatusCode)
			c.backoff(attempt)
			continue
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("session expired (401)")
		case http.StatusForbidden:
			return nil, fmt.Errorf("forbidden (403)")
		case http.StatusNotFound:
			return nil, fmt.Errorf("not found (404)")
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}

		// Detect HTML redirect (auth wall)
		if len(data) > 0 && data[0] == '<' {
			return nil, fmt.Errorf("received HTML instead of JSON (possible auth redirect)")
		}
		if len(data) == 0 {
			return map[string]any{}, nil
		}

		var result any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("decode JSON: %w", err)
		}
		return result, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// rateLimit pauses to respect the per-request delay with Gaussian jitter.
func (c *Client) rateLimit() {
	if c.lastRequestAt.IsZero() {
		return
	}
	elapsed := time.Since(c.lastRequestAt).Seconds()
	target := readDelay

	if elapsed < target {
		jitter := math.Max(0, rand.NormFloat64()*0.15+0.3)
		// 5% chance of a longer pause to mimic reading
		if rand.Float64() < 0.05 {
			jitter += rand.Float64()*3 + 2
		}
		sleep := target - elapsed + jitter
		time.Sleep(time.Duration(sleep * float64(time.Second)))
	}
}

// backoff waits with exponential backoff + jitter.
func (c *Client) backoff(attempt int) {
	wait := math.Pow(2, float64(attempt)) + rand.Float64()
	time.Sleep(time.Duration(wait * float64(time.Second)))
}

// castMap safely casts any to map[string]any.
func CastMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// castSlice safely casts any to []any.
func CastSlice(v any) []any {
	s, _ := v.([]any)
	return s
}

// castString safely casts any to string.
func CastString(v any) string {
	s, _ := v.(string)
	return s
}

// castFloat safely casts any to float64.
func CastFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	}
	return 0
}

// castInt safely casts any to int.
func CastInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

// castBool safely casts any to bool.
func CastBool(v any) bool {
	b, _ := v.(bool)
	return b
}

// QueryString builds URL params from a map, skipping empty values.
func QueryString(pairs ...string) map[string]string {
	m := make(map[string]string, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			m[pairs[i]] = pairs[i+1]
		}
	}
	return m
}

// JoinPath formats a path with named parameters.
func JoinPath(tmpl string, replacements ...string) string {
	result := tmpl
	for i := 0; i+1 < len(replacements); i += 2 {
		result = strings.ReplaceAll(result, "{"+replacements[i]+"}", replacements[i+1])
	}
	return result
}
