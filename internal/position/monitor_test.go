package position

import (
	"fmt"
	"testing"
	"time"

	"prediction-bot/internal/persistence"
	"prediction-bot/internal/volatility"
)

func TestCheckStopLoss_TriggerExit(t *testing.T) {
	// Position entry at 0.90, current price at 0.76 (>15% loss) → trigger exit
	// Stop loss threshold: entry_price * (1 - stop_loss_percent)
	// 0.90 * (1 - 0.15) = 0.765
	// Current price 0.76 < 0.765 → should trigger

	monitor := NewMonitor(0.15) // 15% stop loss

	position := &persistence.Position{
		EntryPrice: 0.90,
		Status:     "open",
	}

	triggered := monitor.CheckStopLoss(position, 0.76)
	if !triggered {
		t.Errorf("CheckStopLoss: expected true (trigger exit) for entry=0.90, current=0.76, got false")
	}
}

func TestCheckStopLoss_NoTrigger(t *testing.T) {
	// Position entry at 0.90, current price at 0.85 (only 5.5% loss) → no trigger
	// Threshold: 0.90 * 0.85 = 0.765
	// Current 0.85 > 0.765 → should not trigger

	monitor := NewMonitor(0.15)

	position := &persistence.Position{
		EntryPrice: 0.90,
		Status:     "open",
	}

	triggered := monitor.CheckStopLoss(position, 0.85)
	if triggered {
		t.Errorf("CheckStopLoss: expected false (no trigger) for entry=0.90, current=0.85, got true")
	}
}

func TestCheckStopLoss_ExactlyAtThreshold(t *testing.T) {
	// Position entry at 0.90, threshold = 0.765
	// Current price exactly at threshold → should NOT trigger (use strict less-than)

	monitor := NewMonitor(0.15)

	position := &persistence.Position{
		EntryPrice: 0.90,
		Status:     "open",
	}

	// Threshold = 0.90 * (1 - 0.15) = 0.765
	triggered := monitor.CheckStopLoss(position, 0.765)
	if triggered {
		t.Errorf("CheckStopLoss: expected false at exact threshold (0.765), got true")
	}
}

func TestCheckStopLoss_JustBelowThreshold(t *testing.T) {
	// Position entry at 0.90, current price just below threshold → trigger

	monitor := NewMonitor(0.15)

	position := &persistence.Position{
		EntryPrice: 0.90,
		Status:     "open",
	}

	// Threshold = 0.765, price = 0.764 → trigger
	triggered := monitor.CheckStopLoss(position, 0.764)
	if !triggered {
		t.Errorf("CheckStopLoss: expected true for price just below threshold (0.764), got false")
	}
}

func TestCheckStopLoss_VariousStopLossPercents(t *testing.T) {
	tests := []struct {
		name           string
		stopLossPercent float64
		entryPrice     float64
		currentPrice   float64
		expectTrigger  bool
	}{
		{
			name:           "10% stop loss, 12% drop",
			stopLossPercent: 0.10,
			entryPrice:     0.80,
			currentPrice:   0.70, // threshold = 0.72, 0.70 < 0.72 → trigger
			expectTrigger:  true,
		},
		{
			name:           "10% stop loss, 8% drop",
			stopLossPercent: 0.10,
			entryPrice:     0.80,
			currentPrice:   0.74, // threshold = 0.72, 0.74 > 0.72 → no trigger
			expectTrigger:  false,
		},
		{
			name:           "20% stop loss, 25% drop",
			stopLossPercent: 0.20,
			entryPrice:     0.50,
			currentPrice:   0.375, // threshold = 0.40, 0.375 < 0.40 → trigger
			expectTrigger:  true,
		},
		{
			name:           "5% stop loss, 3% drop",
			stopLossPercent: 0.05,
			entryPrice:     0.95,
			currentPrice:   0.92, // threshold = 0.9025, 0.92 > 0.9025 → no trigger
			expectTrigger:  false,
		},
		{
			name:           "price went up",
			stopLossPercent: 0.15,
			entryPrice:     0.85,
			currentPrice:   0.92, // price went UP → definitely no trigger
			expectTrigger:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewMonitor(tt.stopLossPercent)
			position := &persistence.Position{
				EntryPrice: tt.entryPrice,
				Status:     "open",
			}

			triggered := monitor.CheckStopLoss(position, tt.currentPrice)
			if triggered != tt.expectTrigger {
				threshold := tt.entryPrice * (1 - tt.stopLossPercent)
				t.Errorf("CheckStopLoss: entry=%.4f, current=%.4f, threshold=%.4f, expected trigger=%v, got %v",
					tt.entryPrice, tt.currentPrice, threshold, tt.expectTrigger, triggered)
			}
		})
	}
}

// MockVolatilityAnalyzer implements the VolatilityAnalyzer interface for testing.
type MockVolatilityAnalyzer struct {
	safetyMargin float64
	err          error
}

func (m *MockVolatilityAnalyzer) AnalyzeAsset(asset string, strikePrice float64, direction volatility.Direction, timeToClose time.Duration) (volatility.ServiceResult, error) {
	if m.err != nil {
		return volatility.ServiceResult{}, m.err
	}
	return volatility.ServiceResult{
		SafetyMargin: m.safetyMargin,
	}, nil
}

func TestCheckVolatilityExit_TriggerOnLowSafetyMargin(t *testing.T) {
	// Safety margin atual < 0.8 → trigger volatility exit
	// Position has strike $100k, direction "above", but volatility increased
	// making current safety margin 0.6 < 0.8 → trigger exit

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: 0.6, err: nil}

	position := &persistence.Position{
		ID:        1,
		Asset:     "BTC",
		Strike:    100000,
		Direction: "above",
		Status:    "open",
	}

	triggered, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 24*time.Hour)
	if err != nil {
		t.Fatalf("CheckVolatilityExit returned error: %v", err)
	}
	if !triggered {
		t.Errorf("CheckVolatilityExit: expected true (trigger exit) for safety_margin=0.6, got false")
	}
}

func TestCheckVolatilityExit_NoTriggerOnGoodSafetyMargin(t *testing.T) {
	// Safety margin atual >= 0.8 → no trigger
	// Position has safety margin 1.2 (risky but acceptable) → no exit

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: 1.2, err: nil}

	position := &persistence.Position{
		ID:        1,
		Asset:     "BTC",
		Strike:    100000,
		Direction: "above",
		Status:    "open",
	}

	triggered, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 24*time.Hour)
	if err != nil {
		t.Fatalf("CheckVolatilityExit returned error: %v", err)
	}
	if triggered {
		t.Errorf("CheckVolatilityExit: expected false (no trigger) for safety_margin=1.2, got true")
	}
}

func TestCheckVolatilityExit_NoTriggerOnValidSafetyMargin(t *testing.T) {
	// Safety margin atual >= 1.5 → definitely no trigger

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: 2.5, err: nil}

	position := &persistence.Position{
		ID:        1,
		Asset:     "ETH",
		Strike:    3000,
		Direction: "below",
		Status:    "open",
	}

	triggered, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 12*time.Hour)
	if err != nil {
		t.Fatalf("CheckVolatilityExit returned error: %v", err)
	}
	if triggered {
		t.Errorf("CheckVolatilityExit: expected false (no trigger) for safety_margin=2.5, got true")
	}
}

func TestCheckVolatilityExit_ExactlyAtThreshold(t *testing.T) {
	// Safety margin exactly at 0.8 → should NOT trigger (use strict less-than)

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: 0.8, err: nil}

	position := &persistence.Position{
		ID:        1,
		Asset:     "BTC",
		Strike:    100000,
		Direction: "above",
		Status:    "open",
	}

	triggered, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 24*time.Hour)
	if err != nil {
		t.Fatalf("CheckVolatilityExit returned error: %v", err)
	}
	if triggered {
		t.Errorf("CheckVolatilityExit: expected false at exact threshold (0.8), got true")
	}
}

func TestCheckVolatilityExit_JustBelowThreshold(t *testing.T) {
	// Safety margin just below 0.8 → trigger

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: 0.79, err: nil}

	position := &persistence.Position{
		ID:        1,
		Asset:     "BTC",
		Strike:    100000,
		Direction: "above",
		Status:    "open",
	}

	triggered, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 24*time.Hour)
	if err != nil {
		t.Fatalf("CheckVolatilityExit returned error: %v", err)
	}
	if !triggered {
		t.Errorf("CheckVolatilityExit: expected true for safety_margin=0.79, got false")
	}
}

func TestCheckVolatilityExit_NegativeSafetyMargin(t *testing.T) {
	// Negative safety margin (on wrong side of strike) → trigger

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: -0.5, err: nil}

	position := &persistence.Position{
		ID:        1,
		Asset:     "BTC",
		Strike:    100000,
		Direction: "above",
		Status:    "open",
	}

	triggered, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 24*time.Hour)
	if err != nil {
		t.Fatalf("CheckVolatilityExit returned error: %v", err)
	}
	if !triggered {
		t.Errorf("CheckVolatilityExit: expected true for negative safety_margin=-0.5, got false")
	}
}

func TestCheckVolatilityExit_ErrorFromAnalyzer(t *testing.T) {
	// Error from analyzer should be propagated

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: 0, err: fmt.Errorf("failed to fetch price data")}

	position := &persistence.Position{
		ID:        1,
		Asset:     "INVALID",
		Strike:    100,
		Direction: "above",
		Status:    "open",
	}

	_, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 24*time.Hour)
	if err == nil {
		t.Errorf("CheckVolatilityExit: expected error from analyzer, got nil")
	}
}

func TestCheckVolatilityExit_DirectionBelow(t *testing.T) {
	// Test that direction "below" is correctly converted

	monitor := NewMonitor(0.15)
	mockAnalyzer := &MockVolatilityAnalyzer{safetyMargin: 0.5, err: nil}

	position := &persistence.Position{
		ID:        1,
		Asset:     "ETH",
		Strike:    3000,
		Direction: "below", // testing "below" direction
		Status:    "open",
	}

	triggered, err := monitor.CheckVolatilityExit(position, mockAnalyzer, 12*time.Hour)
	if err != nil {
		t.Fatalf("CheckVolatilityExit returned error: %v", err)
	}
	if !triggered {
		t.Errorf("CheckVolatilityExit: expected true for safety_margin=0.5, got false")
	}
}
