package alphavantage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"prediction-bot/pkg/types"
)

const (
	baseURL = "https://www.alphavantage.co/query"
)

// Client is an Alpha Vantage API client.
type Client struct {
	httpClient *http.Client
	apiKey     string
}

// NewClient creates a new Alpha Vantage client.
// It reads the API key from ALPHAVANTAGE_API_KEY environment variable.
func NewClient() (*Client, error) {
	apiKey := os.Getenv("ALPHAVANTAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ALPHAVANTAGE_API_KEY environment variable not set")
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiKey: apiKey,
	}, nil
}

// NewClientWithKey creates a new Alpha Vantage client with an explicit API key.
func NewClientWithKey(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiKey: apiKey,
	}
}

// globalQuoteResponse represents the Alpha Vantage GLOBAL_QUOTE response.
type globalQuoteResponse struct {
	GlobalQuote struct {
		Symbol string `json:"01. symbol"`
		Price  string `json:"05. price"`
	} `json:"Global Quote"`
}

// GetPrice fetches the current price for a symbol from Alpha Vantage.
func (c *Client) GetPrice(symbol string) (types.Price, error) {
	url := fmt.Sprintf("%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		baseURL, symbol, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return types.Price{}, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.Price{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var quote globalQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return types.Price{}, fmt.Errorf("decode response: %w", err)
	}

	if quote.GlobalQuote.Symbol == "" {
		return types.Price{}, fmt.Errorf("empty response (rate limit or invalid symbol)")
	}

	price, err := strconv.ParseFloat(quote.GlobalQuote.Price, 64)
	if err != nil {
		return types.Price{}, fmt.Errorf("parse price: %w", err)
	}

	return types.Price{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now(),
		Source:    "alphavantage",
	}, nil
}
