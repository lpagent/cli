package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	verbose    bool
}

func New(baseURL, apiKey string, verbose bool) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
	}
}

type APIError struct {
	StatusCode int
	Status     string `json:"status"`
	Message    string `json:"message"`
	RawBody    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.RawBody)
}

func (c *Client) Get(path string, params map[string]string) (json.RawMessage, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	return c.do("GET", u.String(), nil)
}

func (c *Client) Post(path string, body any) (json.RawMessage, error) {
	u := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	return c.do("POST", u, bodyReader)
}

func (c *Client) do(method, rawURL string, body io.Reader) (json.RawMessage, error) {
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "lpagent-cli/1.0")

	if c.verbose {
		fmt.Fprintf(os.Stderr, "→ %s %s\n", method, rawURL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "← %d (%d bytes)\n", resp.StatusCode, len(respBody))
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode, RawBody: string(respBody)}
		_ = json.Unmarshal(respBody, apiErr)

		switch resp.StatusCode {
		case 401:
			return nil, fmt.Errorf("unauthorized: invalid or expired API key. Run: lpagent auth set-key")
		case 429:
			return nil, fmt.Errorf("rate limited: too many requests. Please wait and try again")
		default:
			return nil, apiErr
		}
	}

	return json.RawMessage(respBody), nil
}
