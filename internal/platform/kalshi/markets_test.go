package kalshi

import (
	"net/http"
	"testing"
	"time"

	"prediction-bot/pkg/types"
)

func TestClient_ListMarkets_ReturnsActiveMarkets(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}

	// Filter for active markets only
	isActive := true
	filter := types.MarketFilter{
		IsActive: &isActive,
		Limit:    10,
	}

	markets, err := client.ListMarkets(filter)
	if err != nil {
		t.Fatalf("ListMarkets failed: %v", err)
	}

	// Should return at least one market
	if len(markets) == 0 {
		t.Error("expected at least one active market, got 0")
	}

	// Validate returned markets
	for _, m := range markets {
		if m.ID == "" {
			t.Error("market ID should not be empty")
		}
		if m.Platform != "kalshi" {
			t.Errorf("expected platform 'kalshi', got '%s'", m.Platform)
		}
		if m.Title == "" {
			t.Error("market title should not be empty")
		}
		// Active markets should have Active=true
		if !m.Active {
			t.Errorf("expected active market, got inactive for market %s", m.ID)
		}
	}

	t.Logf("Found %d active markets", len(markets))
	if len(markets) > 0 {
		t.Logf("First market: ID=%s, Title=%s", markets[0].ID, markets[0].Title)
	}
}

func TestClient_ListMarkets_RespectsLimit(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}

	filter := types.MarketFilter{
		Limit: 5,
	}

	markets, err := client.ListMarkets(filter)
	if err != nil {
		t.Fatalf("ListMarkets failed: %v", err)
	}

	if len(markets) > 5 {
		t.Errorf("expected at most 5 markets, got %d", len(markets))
	}

	t.Logf("Got %d markets with limit=5", len(markets))
}

func TestClient_ListMarkets_MapsMktFieldsCorrectly(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}

	filter := types.MarketFilter{
		Limit: 1,
	}

	markets, err := client.ListMarkets(filter)
	if err != nil {
		t.Fatalf("ListMarkets failed: %v", err)
	}

	if len(markets) == 0 {
		t.Skip("no markets available to test field mapping")
	}

	m := markets[0]

	// Log all fields for debugging
	t.Logf("Market: ID=%s", m.ID)
	t.Logf("Market: Platform=%s", m.Platform)
	t.Logf("Market: Title=%s", m.Title)
	t.Logf("Market: Description=%s", m.Description)
	t.Logf("Market: EndDate=%v", m.EndDate)
	t.Logf("Market: Volume=%f", m.Volume)
	t.Logf("Market: Active=%v", m.Active)
	t.Logf("Market: Closed=%v", m.Closed)
	t.Logf("Market: YesPrice=%f", m.OutcomeYesPrice)
	t.Logf("Market: NoPrice=%f", m.OutcomeNoPrice)

	// EndDate should be in the future for active markets
	if m.Active && !m.EndDate.IsZero() && m.EndDate.Before(time.Now()) {
		t.Logf("Warning: active market with EndDate in the past: %v", m.EndDate)
	}
}
