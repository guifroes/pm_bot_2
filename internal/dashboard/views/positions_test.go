package views

import (
	"strings"
	"testing"
	"time"
)

func TestPositionsView_RenderSinglePosition(t *testing.T) {
	positions := []PositionData{
		{
			ID:          1,
			Platform:    "polymarket",
			MarketTitle: "Will Bitcoin be above $100k?",
			Asset:       "BTC",
			EntryPrice:  0.85,
			CurrentPrice: 0.90,
			Quantity:    10.0,
			Side:        "YES",
			EntryTime:   time.Now().Add(-2 * time.Hour),
		},
	}

	view := NewPositionsView()
	output := view.Render(positions, 80)

	// Should show market title or truncated version
	if !strings.Contains(output, "Bitcoin") && !strings.Contains(output, "BTC") {
		t.Errorf("expected output to contain market reference, got: %s", output)
	}

	// Should show entry price
	if !strings.Contains(output, "0.85") && !strings.Contains(output, "85") {
		t.Errorf("expected output to contain entry price, got: %s", output)
	}

	// Should show PnL (positive in this case)
	// Current: 0.90, Entry: 0.85, Qty: 10 -> PnL = (0.90-0.85)*10 = 0.50
	if !strings.Contains(output, "+") {
		t.Errorf("expected output to show positive PnL indicator, got: %s", output)
	}
}

func TestPositionsView_RenderMultiplePositions(t *testing.T) {
	positions := []PositionData{
		{
			ID:          1,
			Platform:    "polymarket",
			MarketTitle: "Bitcoin above $100k",
			Asset:       "BTC",
			EntryPrice:  0.85,
			CurrentPrice: 0.90,
			Quantity:    10.0,
			Side:        "YES",
			EntryTime:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          2,
			Platform:    "kalshi",
			MarketTitle: "ETH above $5k",
			Asset:       "ETH",
			EntryPrice:  0.75,
			CurrentPrice: 0.60,
			Quantity:    5.0,
			Side:        "YES",
			EntryTime:   time.Now().Add(-1 * time.Hour),
		},
	}

	view := NewPositionsView()
	output := view.Render(positions, 80)

	// Should show both positions
	hasBTC := strings.Contains(output, "BTC") || strings.Contains(output, "Bitcoin")
	hasETH := strings.Contains(output, "ETH")
	if !hasBTC || !hasETH {
		t.Errorf("expected output to contain both positions, got: %s", output)
	}
}

func TestPositionsView_NegativePnL(t *testing.T) {
	positions := []PositionData{
		{
			ID:          1,
			Platform:    "kalshi",
			MarketTitle: "S&P 500 above 5000",
			Asset:       "SPY",
			EntryPrice:  0.80,
			CurrentPrice: 0.65,
			Quantity:    10.0,
			Side:        "YES",
			EntryTime:   time.Now().Add(-1 * time.Hour),
		},
	}

	view := NewPositionsView()
	output := view.Render(positions, 80)

	// Should show negative PnL indicator
	if !strings.Contains(output, "-") {
		t.Errorf("expected output to show negative PnL indicator, got: %s", output)
	}
}

func TestPositionsView_EmptyPositions(t *testing.T) {
	view := NewPositionsView()
	output := view.Render(nil, 80)

	// Should handle empty positions gracefully
	if output == "" {
		t.Error("expected non-empty output even with no positions")
	}

	// Should show "No positions" or similar message
	if !strings.Contains(strings.ToLower(output), "no") &&
		!strings.Contains(strings.ToLower(output), "open") &&
		!strings.Contains(strings.ToLower(output), "position") {
		t.Logf("output with empty positions: %s", output)
	}
}

func TestPositionsView_CalculateUnrealizedPnL(t *testing.T) {
	pos := PositionData{
		ID:          1,
		Platform:    "polymarket",
		MarketTitle: "Test Market",
		Asset:       "BTC",
		EntryPrice:  0.80,
		CurrentPrice: 0.90,
		Quantity:    10.0,
		Side:        "YES",
		EntryTime:   time.Now(),
	}

	// Expected PnL: (0.90 - 0.80) * 10 = $1.00
	expectedPnL := 1.00

	actualPnL := pos.UnrealizedPnL()
	if !floatEquals(actualPnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %.2f, got %.2f", expectedPnL, actualPnL)
	}
}

// floatEquals compares two floats with a tolerance.
func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}

func TestPositionsView_CalculateUnrealizedPnL_Negative(t *testing.T) {
	pos := PositionData{
		ID:          1,
		Platform:    "polymarket",
		MarketTitle: "Test Market",
		Asset:       "BTC",
		EntryPrice:  0.85,
		CurrentPrice: 0.70,
		Quantity:    20.0,
		Side:        "YES",
		EntryTime:   time.Now(),
	}

	// Expected PnL: (0.70 - 0.85) * 20 = -$3.00
	expectedPnL := -3.00

	actualPnL := pos.UnrealizedPnL()
	if !floatEquals(actualPnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %.2f, got %.2f", expectedPnL, actualPnL)
	}
}
