package scanner

import (
	"testing"
	"time"

	"prediction-bot/internal/config"
	"prediction-bot/internal/platform"
	"prediction-bot/pkg/types"
)

// MockPlatform implements platform.Platform for testing
type MockPlatform struct {
	name    string
	markets []types.Market
	err     error
}

func (m *MockPlatform) Name() string {
	return m.name
}

func (m *MockPlatform) ListMarkets(filter types.MarketFilter) ([]types.Market, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.markets, nil
}

func (m *MockPlatform) GetOrderBook(tokenID string) (*types.OrderBook, error) {
	return nil, nil
}

func (m *MockPlatform) GetBalance() (float64, error) {
	return 0, nil
}

func (m *MockPlatform) GetPositions() ([]types.Position, error) {
	return nil, nil
}

// Ensure MockPlatform implements Platform
var _ platform.Platform = (*MockPlatform)(nil)

func TestScanner_Scan_MockPlatform(t *testing.T) {
	// Setup mock platform with mixed markets
	now := time.Now()
	mockPlatform := &MockPlatform{
		name: "mock",
		markets: []types.Market{
			{
				ID:              "eligible-btc",
				Platform:        "mock",
				Title:           "Will Bitcoin be above $100,000 on Jan 20?",
				EndDate:         now.Add(24 * time.Hour), // 24h from now
				Active:          true,
				Closed:          false,
				OutcomeYesPrice: 0.92, // 92% probability
				OutcomeNoPrice:  0.08,
				Liquidity:       500.0,
			},
			{
				ID:              "ineligible-low-prob",
				Platform:        "mock",
				Title:           "Will Ethereum be above $5,000 on Jan 20?",
				EndDate:         now.Add(24 * time.Hour),
				Active:          true,
				Closed:          false,
				OutcomeYesPrice: 0.50, // 50% - too low
				OutcomeNoPrice:  0.50,
				Liquidity:       500.0,
			},
			{
				ID:              "ineligible-too-far",
				Platform:        "mock",
				Title:           "Will Solana be above $300 on Feb 15?",
				EndDate:         now.Add(72 * time.Hour), // 72h - too far
				Active:          true,
				Closed:          false,
				OutcomeYesPrice: 0.85,
				OutcomeNoPrice:  0.15,
				Liquidity:       500.0,
			},
			{
				ID:              "ineligible-closed",
				Platform:        "mock",
				Title:           "Will Bitcoin be below $90,000?",
				EndDate:         now.Add(12 * time.Hour),
				Active:          true,
				Closed:          true, // Already closed
				OutcomeYesPrice: 0.90,
				OutcomeNoPrice:  0.10,
				Liquidity:       500.0,
			},
			{
				ID:              "eligible-no-side",
				Platform:        "mock",
				Title:           "Will Bitcoin be below $95,000 on Jan 21?",
				EndDate:         now.Add(36 * time.Hour),
				Active:          true,
				Closed:          false,
				OutcomeYesPrice: 0.15,
				OutcomeNoPrice:  0.85, // NO side is 85%
				Liquidity:       200.0,
			},
			{
				ID:              "unparseable",
				Platform:        "mock",
				Title:           "Who will win the election?", // Cannot parse asset/strike
				EndDate:         now.Add(12 * time.Hour),
				Active:          true,
				Closed:          false,
				OutcomeYesPrice: 0.90,
				OutcomeNoPrice:  0.10,
				Liquidity:       1000.0,
			},
		},
	}

	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}

	scanner := NewScanner(params)
	eligible, err := scanner.Scan(mockPlatform)

	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	// Should have 2 eligible markets: eligible-btc and eligible-no-side
	if len(eligible) != 2 {
		t.Errorf("Expected 2 eligible markets, got %d", len(eligible))
		for _, e := range eligible {
			t.Logf("  - %s: %s (prob=%.2f, side=%s)", e.Market.ID, e.Market.Title, e.Probability, e.BetSide)
		}
	}

	// Verify first eligible market
	found := false
	for _, e := range eligible {
		if e.Market.ID == "eligible-btc" {
			found = true
			if e.Parsed.Asset != "BTC" {
				t.Errorf("Expected asset BTC, got %s", e.Parsed.Asset)
			}
			if e.Parsed.Strike != 100000 {
				t.Errorf("Expected strike 100000, got %.0f", e.Parsed.Strike)
			}
			if e.Parsed.Direction != "above" {
				t.Errorf("Expected direction above, got %s", e.Parsed.Direction)
			}
			if e.BetSide != "YES" {
				t.Errorf("Expected bet side YES, got %s", e.BetSide)
			}
			if e.Probability != 0.92 {
				t.Errorf("Expected probability 0.92, got %.2f", e.Probability)
			}
		}
	}
	if !found {
		t.Error("eligible-btc market not found in results")
	}

	// Verify NO side market
	found = false
	for _, e := range eligible {
		if e.Market.ID == "eligible-no-side" {
			found = true
			if e.BetSide != "NO" {
				t.Errorf("Expected bet side NO, got %s", e.BetSide)
			}
			if e.Probability != 0.85 {
				t.Errorf("Expected probability 0.85, got %.2f", e.Probability)
			}
		}
	}
	if !found {
		t.Error("eligible-no-side market not found in results")
	}
}

// TestScanner_Scan_EmptyPlatform tests scanning a platform with no markets
func TestScanner_Scan_EmptyPlatform(t *testing.T) {
	mockPlatform := &MockPlatform{
		name:    "empty",
		markets: []types.Market{},
	}

	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}

	scanner := NewScanner(params)
	eligible, err := scanner.Scan(mockPlatform)

	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if len(eligible) != 0 {
		t.Errorf("Expected 0 eligible markets for empty platform, got %d", len(eligible))
	}
}

// TestScanner_Scan_AllUnparseable tests scanning when all markets are unparseable
func TestScanner_Scan_AllUnparseable(t *testing.T) {
	now := time.Now()
	mockPlatform := &MockPlatform{
		name: "unparseable",
		markets: []types.Market{
			{
				ID:              "political",
				Platform:        "mock",
				Title:           "Who will win the 2024 election?",
				EndDate:         now.Add(24 * time.Hour),
				Active:          true,
				Closed:          false,
				OutcomeYesPrice: 0.85,
				OutcomeNoPrice:  0.15,
				Liquidity:       1000.0,
			},
			{
				ID:              "sports",
				Platform:        "mock",
				Title:           "Will the Lakers win the championship?",
				EndDate:         now.Add(12 * time.Hour),
				Active:          true,
				Closed:          false,
				OutcomeYesPrice: 0.90,
				OutcomeNoPrice:  0.10,
				Liquidity:       500.0,
			},
		},
	}

	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}

	scanner := NewScanner(params)
	eligible, err := scanner.Scan(mockPlatform)

	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	// All markets eligible by criteria but unparseable - should return 0
	if len(eligible) != 0 {
		t.Errorf("Expected 0 eligible markets (all unparseable), got %d", len(eligible))
	}
}
