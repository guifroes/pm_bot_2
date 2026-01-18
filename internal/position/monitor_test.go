package position

import (
	"testing"

	"prediction-bot/internal/persistence"
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
