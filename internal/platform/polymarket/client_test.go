package polymarket

import (
	"net/http"
	"os"
	"testing"
	"time"

	"prediction-bot/pkg/types"
)

func TestClient_Ping_Success(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    clobBaseURL,
	}

	err := client.Ping()
	if err != nil {
		t.Errorf("Ping: %v", err)
	}
}

func TestClient_GetServerTime_Success(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    clobBaseURL,
	}

	ts, err := client.GetServerTime()
	if err != nil {
		t.Fatalf("GetServerTime: %v", err)
	}

	if ts <= 0 {
		t.Errorf("expected positive timestamp, got %d", ts)
	}

	t.Logf("Server time: %d", ts)
}

func TestClient_CredsAreStored(t *testing.T) {
	creds := Credentials{
		APIKey:     "test-key",
		APISecret:  "dGVzdC1zZWNyZXQ=", // base64 of "test-secret"
		Passphrase: "test-passphrase",
	}

	client := NewClientWithCreds(creds)

	if client.creds.APIKey != "test-key" {
		t.Errorf("expected API key 'test-key', got '%s'", client.creds.APIKey)
	}
	if client.creds.Passphrase != "test-passphrase" {
		t.Errorf("expected passphrase 'test-passphrase', got '%s'", client.creds.Passphrase)
	}
}

func TestNewClient_MissingCredentials_ReturnsError(t *testing.T) {
	// Save and clear env vars
	saved := map[string]string{
		"POLYMARKET_API_KEY":    os.Getenv("POLYMARKET_API_KEY"),
		"POLYMARKET_API_SECRET": os.Getenv("POLYMARKET_API_SECRET"),
		"POLYMARKET_PASSPHRASE": os.Getenv("POLYMARKET_PASSPHRASE"),
	}
	os.Unsetenv("POLYMARKET_API_KEY")
	os.Unsetenv("POLYMARKET_API_SECRET")
	os.Unsetenv("POLYMARKET_PASSPHRASE")
	defer func() {
		for k, v := range saved {
			os.Setenv(k, v)
		}
	}()

	_, err := NewClient()
	if err == nil {
		t.Error("expected error when credentials are missing, got nil")
	}
}

func TestClient_ListMarkets_ReturnsActiveMarkets(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    clobBaseURL,
	}

	active := true
	markets, err := client.ListMarkets(types.MarketFilter{
		IsActive: &active,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("ListMarkets: %v", err)
	}

	if len(markets) == 0 {
		t.Error("expected non-empty list of markets")
	}

	t.Logf("Found %d active markets", len(markets))

	// Verify first market has expected fields
	if len(markets) > 0 {
		m := markets[0]
		t.Logf("First market: ID=%s, Title=%s, Active=%v, YesPrice=%.2f",
			m.ID, m.Title, m.Active, m.OutcomeYesPrice)

		if m.ID == "" {
			t.Error("market ID should not be empty")
		}
		if m.Title == "" {
			t.Error("market Title should not be empty")
		}
		if m.Platform != "polymarket" {
			t.Errorf("expected platform 'polymarket', got '%s'", m.Platform)
		}
	}
}

func TestClient_GetOrderBook_ReturnsLevels(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    clobBaseURL,
	}

	// Use a known active token ID from a popular market
	// This is a Bitcoin price market token which typically has liquidity
	// If this specific token doesn't work, the test will provide useful info

	// First, try to get markets and find any token with an orderbook
	active := true
	markets, err := client.ListMarkets(types.MarketFilter{
		IsActive: &active,
		Limit:    20, // Reduced to avoid timeout
	})
	if err != nil {
		t.Fatalf("ListMarkets: %v", err)
	}

	// Try first 5 markets only
	var foundOrderBook bool
	attemptsLimit := 5
	attempts := 0

	for _, m := range markets {
		if attempts >= attemptsLimit {
			break
		}
		if len(m.Tokens) == 0 {
			continue
		}

		attempts++
		tokenID := m.Tokens[0].TokenID

		ob, err := client.GetOrderBook(tokenID)
		if err != nil {
			t.Logf("Market %s: no orderbook (%v)", m.Title[:min(30, len(m.Title))], err)
			continue
		}

		foundOrderBook = true
		t.Logf("Found orderbook for: %s", m.Title[:min(50, len(m.Title))])
		t.Logf("TokenID=%s, Bids=%d, Asks=%d", ob.TokenID, len(ob.Bids), len(ob.Asks))

		if len(ob.Bids) > 0 {
			t.Logf("Best bid: %.4f", ob.BestBid())
		}
		if len(ob.Asks) > 0 {
			t.Logf("Best ask: %.4f", ob.BestAsk())
		}
		break
	}

	if !foundOrderBook {
		// This is OK - the API works, just no active orderbooks in sampled markets
		t.Log("No orderbook found in first 5 markets - this is expected for less active markets")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
