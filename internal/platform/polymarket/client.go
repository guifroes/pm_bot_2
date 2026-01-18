package polymarket

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// clobBaseURL is the Polymarket CLOB API base URL
	clobBaseURL = "https://clob.polymarket.com"
)

// Client is a Polymarket CLOB API client.
type Client struct {
	httpClient *http.Client
	creds      Credentials
	baseURL    string
}

// NewClient creates a new Polymarket client from environment variables.
func NewClient() (*Client, error) {
	apiKey := os.Getenv("POLYMARKET_API_KEY")
	apiSecret := os.Getenv("POLYMARKET_API_SECRET")
	passphrase := os.Getenv("POLYMARKET_PASSPHRASE")

	if apiKey == "" || apiSecret == "" || passphrase == "" {
		return nil, fmt.Errorf("missing Polymarket credentials in environment")
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		creds: Credentials{
			APIKey:     apiKey,
			APISecret:  apiSecret,
			Passphrase: passphrase,
		},
		baseURL: clobBaseURL,
	}, nil
}

// NewClientWithCreds creates a new Polymarket client with explicit credentials.
func NewClientWithCreds(creds Credentials) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		creds:   creds,
		baseURL: clobBaseURL,
	}
}

// doRequest performs an authenticated request to the Polymarket API.
func (c *Client) doRequest(method, path string, body []byte) ([]byte, error) {
	timestamp := getTimestamp()

	signature, err := generateL2Signature(c.creds, timestamp, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("generate signature: %w", err)
	}

	url := c.baseURL + path
	var reqBody io.Reader
	if body != nil {
		reqBody = &byteReader{data: body}
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set authentication headers
	req.Header.Set("POLY_API_KEY", c.creds.APIKey)
	req.Header.Set("POLY_SIGNATURE", signature)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_PASSPHRASE", c.creds.Passphrase)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// byteReader wraps a byte slice to implement io.Reader
type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// Ping tests the connection to Polymarket API.
func (c *Client) Ping() error {
	// Use the /time endpoint which doesn't require auth but tests connectivity
	resp, err := c.httpClient.Get(c.baseURL + "/time")
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// doPublicRequest performs a non-authenticated request to the Polymarket API.
func (c *Client) doPublicRequest(method, path string) ([]byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetServerTime fetches the server time (public endpoint, no auth needed).
func (c *Client) GetServerTime() (int64, error) {
	body, err := c.doPublicRequest("GET", "/time")
	if err != nil {
		return 0, err
	}

	// Response is just a timestamp string
	var timestamp int64
	if _, err := fmt.Sscanf(string(body), "%d", &timestamp); err != nil {
		return 0, fmt.Errorf("parse timestamp: %w", err)
	}

	return timestamp, nil
}
