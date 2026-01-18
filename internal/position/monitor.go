package position

import (
	"prediction-bot/internal/persistence"
)

// Monitor handles position monitoring for stop loss and volatility exits.
type Monitor struct {
	stopLossPercent float64
}

// NewMonitor creates a new position monitor with the given stop loss percentage.
func NewMonitor(stopLossPercent float64) *Monitor {
	return &Monitor{
		stopLossPercent: stopLossPercent,
	}
}

// CheckStopLoss checks if a position should exit due to stop loss.
// Returns true if the current price is strictly below the stop loss threshold.
// Threshold = entry_price * (1 - stop_loss_percent)
func (m *Monitor) CheckStopLoss(position *persistence.Position, currentPrice float64) bool {
	threshold := position.EntryPrice * (1 - m.stopLossPercent)
	return currentPrice < threshold
}
