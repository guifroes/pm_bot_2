package kalshi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// baseURL is the Kalshi Trade API base URL (host only)
	baseURL = "https://api.elections.kalshi.com"
	// apiPath is the API version prefix
	apiPath = "/trade-api/v2"
)

// Client is a Kalshi Trade API client.
type Client struct {
	httpClient *http.Client
	creds      Credentials
	baseURL    string
}

// Balance represents account balance information.
type Balance struct {
	Available        float64 `json:"balance"`
	Reserved         float64 `json:"payout"`
	BonusCashBalance float64 `json:"bonus_cash_balance"`
}

// NewClient creates a new Kalshi client from environment variables.
// Supports KALSHI_PRIVATE_KEY (PEM content) or KALSHI_PRIVATE_KEY_PATH (file path).
func NewClient() (*Client, error) {
	apiKey := os.Getenv("KALSHI_API_KEY")
	privateKey := os.Getenv("KALSHI_PRIVATE_KEY")
	privateKeyPath := os.Getenv("KALSHI_PRIVATE_KEY_PATH")

	// Load private key from file if path is provided
	if privateKey == "" && privateKeyPath != "" {
		data, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read private key file: %w", err)
		}
		privateKey = string(data)
	}

	if apiKey == "" || privateKey == "" {
		return nil, fmt.Errorf("missing Kalshi credentials: KALSHI_API_KEY and KALSHI_PRIVATE_KEY (or KALSHI_PRIVATE_KEY_PATH) required")
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		creds: Credentials{
			APIKey:     apiKey,
			PrivateKey: privateKey,
		},
		baseURL: baseURL,
	}, nil
}

// NewClientWithCreds creates a new Kalshi client with explicit credentials.
func NewClientWithCreds(creds Credentials) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		creds:   creds,
		baseURL: baseURL,
	}
}

// doRequest performs an authenticated request to the Kalshi API.
func (c *Client) doRequest(method, path string, body []byte) ([]byte, error) {
	timestamp := getTimestampMS()

	// Full path includes API version prefix
	fullPath := apiPath + path

	// Strip query parameters for signature
	signPath := fullPath
	if idx := strings.Index(signPath, "?"); idx != -1 {
		signPath = signPath[:idx]
	}

	signature, err := generateSignature(c.creds.PrivateKey, timestamp, method, signPath)
	if err != nil {
		return nil, fmt.Errorf("generate signature: %w", err)
	}

	fullURL := c.baseURL + fullPath
	var reqBody io.Reader
	if body != nil {
		reqBody = &byteReader{data: body}
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set authentication headers
	req.Header.Set("KALSHI-ACCESS-KEY", c.creds.APIKey)
	req.Header.Set("KALSHI-ACCESS-SIGNATURE", signature)
	req.Header.Set("KALSHI-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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

// doPublicRequest performs a non-authenticated request to the Kalshi API.
func (c *Client) doPublicRequest(method, path string) ([]byte, error) {
	fullURL := c.baseURL + apiPath + path

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

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

// Ping tests the connection to Kalshi API using a public endpoint.
func (c *Client) Ping() error {
	// Use the exchange status endpoint which is public
	resp, err := c.httpClient.Get(c.baseURL + apiPath + "/exchange/status")
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ping: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetBalance returns the account balance.
func (c *Client) GetBalance() (*Balance, error) {
	body, err := c.doRequest("GET", "/portfolio/balance", nil)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	var response struct {
		Balance          float64 `json:"balance"`
		Payout           float64 `json:"payout"`
		BonusCashBalance float64 `json:"bonus_cash_balance"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse balance response: %w", err)
	}

	// Convert cents to dollars
	return &Balance{
		Available:        float64(response.Balance) / 100.0,
		Reserved:         float64(response.Payout) / 100.0,
		BonusCashBalance: float64(response.BonusCashBalance) / 100.0,
	}, nil
}

// GetExchangeStatus returns the exchange status (public endpoint, no auth needed).
func (c *Client) GetExchangeStatus() (string, error) {
	body, err := c.doPublicRequest("GET", "/exchange/status")
	if err != nil {
		return "", err
	}

	var response struct {
		ExchangeActive bool   `json:"exchange_active"`
		TradingActive  bool   `json:"trading_active"`
		Status         string `json:"status"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("parse status response: %w", err)
	}

	return response.Status, nil
}

// BuildURL builds a URL with query parameters.
func BuildURL(basePath string, params map[string]string) string {
	if len(params) == 0 {
		return basePath
	}

	values := url.Values{}
	for k, v := range params {
		if v != "" {
			values.Add(k, v)
		}
	}

	if len(values) == 0 {
		return basePath
	}

	return basePath + "?" + values.Encode()
}
