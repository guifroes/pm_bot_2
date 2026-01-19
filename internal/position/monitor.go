package position

import (
	"fmt"
	"time"

	"prediction-bot/internal/persistence"
	"prediction-bot/internal/volatility"
)

// VolatilityExitThreshold is the minimum safety margin before triggering a volatility exit.
// If the current safety margin falls below this threshold, the position should be closed.
const VolatilityExitThreshold = 0.8

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

// CheckVolatilityExit checks if a position should exit due to volatility changes.
// Returns true if the current safety margin is strictly below the exit threshold (0.8).
//
// The safety margin is recalculated using current market data:
//   - Current price of the underlying asset
//   - Current volatility (from recent history)
//   - Remaining time to market close
//
// A safety margin below 0.8 indicates that volatility has increased or price has moved
// unfavorably, making the position too risky to hold.
func (m *Monitor) CheckVolatilityExit(position *persistence.Position, analyzer VolatilityAnalyzer, timeToClose time.Duration) (bool, error) {
	// Convert direction string to volatility.Direction
	direction := volatility.DirectionAbove
	if position.Direction == "below" {
		direction = volatility.DirectionBelow
	}

	// Re-analyze the asset with current data
	result, err := analyzer.AnalyzeAsset(
		position.Asset,
		position.Strike,
		direction,
		timeToClose,
	)
	if err != nil {
		return false, fmt.Errorf("check volatility exit: %w", err)
	}

	// Trigger exit if safety margin is strictly below the threshold
	return result.SafetyMargin < VolatilityExitThreshold, nil
}
