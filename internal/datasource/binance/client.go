package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"prediction-bot/pkg/types"
)

const (
	baseURL = "https://api.binance.com/api/v3"
)

// Client is a Binance API client.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Binance client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// tickerResponse represents the Binance ticker price response.
type tickerResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// GetPrice fetches the current price for a symbol from Binance.
func (c *Client) GetPrice(symbol string) (types.Price, error) {
	url := fmt.Sprintf("%s/ticker/price?symbol=%s", baseURL, symbol)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return types.Price{}, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.Price{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ticker tickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return types.Price{}, fmt.Errorf("decode response: %w", err)
	}

	price, err := strconv.ParseFloat(ticker.Price, 64)
	if err != nil {
		return types.Price{}, fmt.Errorf("parse price: %w", err)
	}

	return types.Price{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now(),
		Source:    "binance",
	}, nil
}
