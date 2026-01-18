package alphavantage

import (
	"os"
	"testing"
)

func TestGetPrice_SPY_ReturnsPositivePrice(t *testing.T) {
	// Skip if no API key is set
	apiKey := os.Getenv("ALPHAVANTAGE_API_KEY")
	if apiKey == "" {
		t.Skip("ALPHAVANTAGE_API_KEY not set, skipping integration test")
	}

	client := NewClientWithKey(apiKey)

	price, err := client.GetPrice("SPY")
	if err != nil {
		t.Fatalf("GetPrice: %v", err)
	}

	if price.Price <= 0 {
		t.Errorf("expected positive price, got %f", price.Price)
	}

	if price.Symbol != "SPY" {
		t.Errorf("expected symbol SPY, got %s", price.Symbol)
	}

	if price.Source != "alphavantage" {
		t.Errorf("expected source alphavantage, got %s", price.Source)
	}

	t.Logf("SPY price: $%.2f", price.Price)
}

func TestNewClient_NoAPIKey_ReturnsError(t *testing.T) {
	// Save and clear the env var
	original := os.Getenv("ALPHAVANTAGE_API_KEY")
	os.Unsetenv("ALPHAVANTAGE_API_KEY")
	defer os.Setenv("ALPHAVANTAGE_API_KEY", original)

	_, err := NewClient()
	if err == nil {
		t.Error("expected error when API key is not set, got nil")
	}
}

func TestNewClientWithKey_CreatesClient(t *testing.T) {
	client := NewClientWithKey("test-key")
	if client == nil {
		t.Error("expected non-nil client")
	}
	if client.apiKey != "test-key" {
		t.Errorf("expected api key 'test-key', got '%s'", client.apiKey)
	}
}
