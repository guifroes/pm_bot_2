package sizing

import (
	"math"
	"testing"
)

func TestSizer_Calculate_AppliesConstraints(t *testing.T) {
	sizer := NewSizer(SizerConfig{
		KellyFraction:    0.25,
		MinPosition:      1.0,  // minimum $1
		MaxBankrollPct:   0.20, // max 20% of bankroll
	})

	tests := []struct {
		name       string
		input      SizingInput
		wantMin    float64
		wantMax    float64
		wantReason string
	}{
		{
			name: "normal case - returns position within constraints",
			input: SizingInput{
				EntryPrice:   0.90,
				WinProb:      0.92,
				Bankroll:     50.0,
				SafetyMargin: 1.5,
			},
			wantMin: 1.0,  // at least minimum
			wantMax: 10.0, // max 20% of $50
		},
		{
			name: "kelly too small - returns minimum",
			input: SizingInput{
				EntryPrice:   0.90,
				WinProb:      0.901, // barely any edge
				Bankroll:     50.0,
				SafetyMargin: 1.5,
			},
			wantMin: 0.0, // kelly returns ~0, below minimum, so 0
			wantMax: 1.0,
		},
		{
			name: "kelly too large - capped at max",
			input: SizingInput{
				EntryPrice:   0.80,
				WinProb:      0.95,
				Bankroll:     100.0,
				SafetyMargin: 2.0,
			},
			wantMin: 1.0,
			wantMax: 20.0, // max 20% of $100
		},
		{
			name: "small bankroll - kelly below minimum returns 0",
			input: SizingInput{
				EntryPrice:   0.85,
				WinProb:      0.90,
				Bankroll:     10.0,
				SafetyMargin: 1.5,
			},
			wantMin: 0.0, // kelly ~$0.83 is below $1 minimum, so returns 0
			wantMax: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sizer.Calculate(tt.input)

			if result.PositionSize < tt.wantMin || result.PositionSize > tt.wantMax {
				t.Errorf("Calculate() position = %v, want between %v and %v",
					result.PositionSize, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestSizer_Calculate_RoundsDown(t *testing.T) {
	sizer := NewSizer(SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	})

	input := SizingInput{
		EntryPrice:   0.90,
		WinProb:      0.92,
		Bankroll:     50.0,
		SafetyMargin: 1.5,
	}

	result := sizer.Calculate(input)

	// Position should be rounded down to 2 decimal places (cents)
	rounded := math.Floor(result.PositionSize*100) / 100
	if result.PositionSize != rounded {
		t.Errorf("Calculate() position = %v should be rounded down to %v", result.PositionSize, rounded)
	}
}

func TestSizer_Calculate_ReturnsMetadata(t *testing.T) {
	sizer := NewSizer(SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	})

	input := SizingInput{
		EntryPrice:   0.90,
		WinProb:      0.92,
		Bankroll:     50.0,
		SafetyMargin: 1.5,
	}

	result := sizer.Calculate(input)

	// Should return metadata about the calculation
	if result.RawKelly <= 0 {
		t.Errorf("Calculate() should return positive RawKelly, got %v", result.RawKelly)
	}
	if result.BankrollPct <= 0 || result.BankrollPct > 0.20 {
		t.Errorf("Calculate() BankrollPct should be between 0 and 0.20, got %v", result.BankrollPct)
	}
}

func TestEstimateWinProbability(t *testing.T) {
	tests := []struct {
		name         string
		marketPrice  float64
		safetyMargin float64
		wantMin      float64
		wantMax      float64
	}{
		{
			name:         "high safety margin - boosts probability",
			marketPrice:  0.90,
			safetyMargin: 2.0, // very safe
			wantMin:      0.92,
			wantMax:      0.96,
		},
		{
			name:         "low safety margin - same as market price",
			marketPrice:  0.90,
			safetyMargin: 0.8, // risky
			wantMin:      0.88,
			wantMax:      0.92,
		},
		{
			name:         "moderate safety margin",
			marketPrice:  0.85,
			safetyMargin: 1.5, // threshold
			wantMin:      0.86,
			wantMax:      0.92,
		},
		{
			name:         "very low probability market",
			marketPrice:  0.80,
			safetyMargin: 1.5,
			wantMin:      0.82,
			wantMax:      0.90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateWinProbability(tt.marketPrice, tt.safetyMargin)

			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("EstimateWinProbability(%v, %v) = %v, want between %v and %v",
					tt.marketPrice, tt.safetyMargin, result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestEstimateWinProbability_NeverExceedsOne(t *testing.T) {
	// Even with very high market price and safety margin
	result := EstimateWinProbability(0.98, 5.0)

	if result > 1.0 {
		t.Errorf("EstimateWinProbability should never return > 1.0, got %v", result)
	}
	if result < 0.98 {
		t.Errorf("EstimateWinProbability should not reduce probability, got %v (market: 0.98)", result)
	}
}

func TestEstimateWinProbability_BoundedByMarketPrice(t *testing.T) {
	// With low safety margin, should not boost probability above market
	// but also should not reduce it significantly
	result := EstimateWinProbability(0.85, 0.5) // very risky

	if result > 0.90 {
		t.Errorf("EstimateWinProbability with low safety margin should not boost too much, got %v", result)
	}
	// Should stay close to market price
	if result < 0.80 {
		t.Errorf("EstimateWinProbability should not reduce too much below market price, got %v", result)
	}
}

func TestSizer_Calculate_NoEdge(t *testing.T) {
	sizer := NewSizer(SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	})

	input := SizingInput{
		EntryPrice:   0.95,
		WinProb:      0.90, // negative edge: paying 0.95 for 0.90 win prob
		Bankroll:     50.0,
		SafetyMargin: 1.5,
	}

	result := sizer.Calculate(input)

	if result.PositionSize != 0 {
		t.Errorf("Calculate() with no edge should return 0, got %v", result.PositionSize)
	}
	if result.Reason != "no_edge" {
		t.Errorf("Calculate() with no edge should have reason 'no_edge', got %v", result.Reason)
	}
}
