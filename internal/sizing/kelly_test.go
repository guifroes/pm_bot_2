package sizing

import (
	"math"
	"testing"
)

func TestCalculateKelly(t *testing.T) {
	tests := []struct {
		name       string
		entryPrice float64
		winProb    float64
		bankroll   float64
		fraction   float64
		wantMin    float64
		wantMax    float64
	}{
		{
			name:       "standard case - high probability market",
			entryPrice: 0.90,
			winProb:    0.92,
			bankroll:   50.0,
			fraction:   0.25,
			wantMin:    2.0,
			wantMax:    3.0,
		},
		{
			name:       "lower probability market",
			entryPrice: 0.85,
			winProb:    0.88,
			bankroll:   100.0,
			fraction:   0.25,
			wantMin:    3.0,
			wantMax:    6.0,
		},
		{
			name:       "edge case - entry equals probability (no edge)",
			entryPrice: 0.90,
			winProb:    0.90,
			bankroll:   50.0,
			fraction:   0.25,
			wantMin:    0.0,
			wantMax:    0.5,
		},
		{
			name:       "negative edge - should return 0",
			entryPrice: 0.95,
			winProb:    0.90,
			bankroll:   50.0,
			fraction:   0.25,
			wantMin:    0.0,
			wantMax:    0.0,
		},
		{
			name:       "full Kelly (fraction=1.0)",
			entryPrice: 0.80,
			winProb:    0.90,
			bankroll:   100.0,
			fraction:   1.0,
			wantMin:    40.0,
			wantMax:    60.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateKelly(tt.entryPrice, tt.winProb, tt.bankroll, tt.fraction)

			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateKelly(%v, %v, %v, %v) = %v, want between %v and %v",
					tt.entryPrice, tt.winProb, tt.bankroll, tt.fraction, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateKelly_Formula(t *testing.T) {
	// Test the exact formula: f = (p*b - q) / b
	// where b = (1-price)/price (odds)
	// p = win probability, q = 1-p

	entryPrice := 0.90
	winProb := 0.92
	bankroll := 50.0
	fraction := 0.25

	// Calculate expected value manually
	// b = (1 - 0.90) / 0.90 = 0.1111...
	// p = 0.92, q = 0.08
	// f = (0.92 * 0.1111 - 0.08) / 0.1111 = (0.1022 - 0.08) / 0.1111 = 0.2
	// fractional kelly = 0.2 * 0.25 = 0.05 (5% of bankroll)
	// position = 50 * 0.05 = $2.50

	result := CalculateKelly(entryPrice, winProb, bankroll, fraction)
	expected := 2.5 // approximately
	tolerance := 0.5

	if math.Abs(result-expected) > tolerance {
		t.Errorf("CalculateKelly formula check: got %v, expected approximately %v (tolerance %v)",
			result, expected, tolerance)
	}
}

func TestCalculateKelly_ZeroInputs(t *testing.T) {
	// Zero bankroll should return 0
	result := CalculateKelly(0.90, 0.92, 0.0, 0.25)
	if result != 0 {
		t.Errorf("CalculateKelly with zero bankroll should return 0, got %v", result)
	}

	// Zero fraction should return 0
	result = CalculateKelly(0.90, 0.92, 50.0, 0.0)
	if result != 0 {
		t.Errorf("CalculateKelly with zero fraction should return 0, got %v", result)
	}
}

func TestCalculateKelly_InvalidInputs(t *testing.T) {
	// Entry price >= 1 should return 0 (no profit possible)
	result := CalculateKelly(1.0, 0.92, 50.0, 0.25)
	if result != 0 {
		t.Errorf("CalculateKelly with entry price 1.0 should return 0, got %v", result)
	}

	// Entry price <= 0 should return 0
	result = CalculateKelly(0.0, 0.92, 50.0, 0.25)
	if result != 0 {
		t.Errorf("CalculateKelly with entry price 0 should return 0, got %v", result)
	}

	// Win probability <= 0 should return 0
	result = CalculateKelly(0.90, 0.0, 50.0, 0.25)
	if result != 0 {
		t.Errorf("CalculateKelly with win prob 0 should return 0, got %v", result)
	}
}
