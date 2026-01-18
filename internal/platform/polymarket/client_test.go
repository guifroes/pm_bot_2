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
