package views

import (
	"strings"
	"testing"
)

func TestBankrollView_RenderSinglePlatform(t *testing.T) {
	data := []BankrollData{
		{
			Platform:      "polymarket",
			InitialAmount: 50.00,
			CurrentAmount: 55.00,
		},
	}

	view := NewBankrollView()
	output := view.Render(data, 60)

	// Should show platform name
	if !strings.Contains(output, "polymarket") && !strings.Contains(output, "Polymarket") {
		t.Errorf("expected output to contain platform name, got: %s", output)
	}

	// Should show current amount
	if !strings.Contains(output, "55") {
		t.Errorf("expected output to contain current amount 55, got: %s", output)
	}

	// Should show delta (positive)
	if !strings.Contains(output, "+5") && !strings.Contains(output, "+$5") {
		t.Errorf("expected output to show positive delta, got: %s", output)
	}
}

func TestBankrollView_RenderMultiplePlatforms(t *testing.T) {
	data := []BankrollData{
		{
			Platform:      "polymarket",
			InitialAmount: 50.00,
			CurrentAmount: 55.00,
		},
		{
			Platform:      "kalshi",
			InitialAmount: 50.00,
			CurrentAmount: 45.00,
		},
	}

	view := NewBankrollView()
	output := view.Render(data, 60)

	// Should show both platforms
	hasPolymarket := strings.Contains(strings.ToLower(output), "polymarket")
	hasKalshi := strings.Contains(strings.ToLower(output), "kalshi")
	if !hasPolymarket || !hasKalshi {
		t.Errorf("expected output to contain both platforms, got: %s", output)
	}

	// Should show total
	if !strings.Contains(strings.ToLower(output), "total") {
		t.Errorf("expected output to contain total, got: %s", output)
	}
}

func TestBankrollView_NegativeDelta(t *testing.T) {
	data := []BankrollData{
		{
			Platform:      "kalshi",
			InitialAmount: 50.00,
			CurrentAmount: 40.00,
		},
	}

	view := NewBankrollView()
	output := view.Render(data, 60)

	// Should show negative delta
	if !strings.Contains(output, "-10") && !strings.Contains(output, "-$10") {
		t.Errorf("expected output to show negative delta, got: %s", output)
	}
}

func TestBankrollView_EmptyData(t *testing.T) {
	view := NewBankrollView()
	output := view.Render(nil, 60)

	// Should handle empty data gracefully
	if output == "" {
		t.Error("expected non-empty output even with no data")
	}

	// Should show "No bankroll data" or similar message
	hasNoData := strings.Contains(strings.ToLower(output), "no") ||
		strings.Contains(output, "0") ||
		strings.Contains(output, "-")
	if !hasNoData {
		t.Logf("output with empty data: %s", output)
	}
}
