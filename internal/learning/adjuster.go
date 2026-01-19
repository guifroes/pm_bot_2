package learning

import (
	"math"
	"time"
)

// MaxAdjustmentPercent is the maximum percentage change allowed per adjustment.
const MaxAdjustmentPercent = 0.10

// MinTradesPerSegment is the minimum number of trades required in a segment
// for it to be considered in the adjustment decision.
const MinTradesPerSegment = 5

// MinTradesForAdjustment is the minimum number of closed trades required
// before the learning system will make any adjustments.
const MinTradesForAdjustment = 20

// AdjustmentCooldown is the minimum time between parameter adjustments.
const AdjustmentCooldown = 24 * time.Hour

// DrawdownRevertThreshold is the percentage loss from peak that triggers
// a revert to default parameters.
const DrawdownRevertThreshold = 0.20

// AdjustmentBounds defines the valid range for a parameter.
type AdjustmentBounds struct {
	Min float64
	Max float64
}

// Adjuster suggests parameter adjustments based on segment performance.
type Adjuster struct{}

// NewAdjuster creates a new Adjuster.
func NewAdjuster() *Adjuster {
	return &Adjuster{}
}

// SuggestAdjustment analyzes segment statistics and suggests a new parameter value.
// The adjustment moves toward the best performing segment but is limited to
// MaxAdjustmentPercent (10%) change per call.
//
// Returns the current value unchanged if:
// - segments is empty
// - no segment has enough trades (MinTradesPerSegment)
// - current value is already in the best segment
func (a *Adjuster) SuggestAdjustment(current float64, segments []SegmentStats, bounds AdjustmentBounds) float64 {
	if len(segments) == 0 {
		return current
	}

	// Find the best performing segment with sufficient trades
	bestSegment := findBestSegment(segments)
	if bestSegment == nil {
		return current
	}

	// Check if current is already in the best segment
	if current >= bestSegment.RangeStart && current < bestSegment.RangeEnd {
		return current
	}

	// Calculate target (center of best segment)
	target := (bestSegment.RangeStart + bestSegment.RangeEnd) / 2

	// Calculate direction and magnitude
	delta := target - current
	maxDelta := current * MaxAdjustmentPercent

	// Limit adjustment to 10%
	var newValue float64
	if math.Abs(delta) > maxDelta {
		if delta > 0 {
			newValue = current + maxDelta
		} else {
			newValue = current - maxDelta
		}
	} else {
		newValue = target
	}

	// Apply bounds
	if newValue < bounds.Min {
		newValue = bounds.Min
	}
	if newValue > bounds.Max {
		newValue = bounds.Max
	}

	return newValue
}

// findBestSegment returns the segment with the best performance,
// considering win rate and average PnL. Returns nil if no segment
// has enough trades.
func findBestSegment(segments []SegmentStats) *SegmentStats {
	var best *SegmentStats
	var bestScore float64

	for i := range segments {
		seg := &segments[i]

		// Skip segments with insufficient data
		if seg.TradeCount < MinTradesPerSegment {
			continue
		}

		// Score = win rate * (1 + normalized avg pnl)
		// This balances consistency (win rate) with profitability (avg pnl)
		score := seg.WinRate
		if seg.AvgPnL > 0 {
			score *= (1 + seg.AvgPnL/10) // Normalize PnL contribution
		}

		if best == nil || score > bestScore {
			best = seg
			bestScore = score
		}
	}

	return best
}

// Guardrails provides safety checks for parameter adjustments.
type Guardrails struct {
	minTrades   int
	cooldown    time.Duration
	revertPct   float64
}

// NewGuardrails creates a new Guardrails with default settings.
func NewGuardrails() *Guardrails {
	return &Guardrails{
		minTrades:   MinTradesForAdjustment,
		cooldown:    AdjustmentCooldown,
		revertPct:   DrawdownRevertThreshold,
	}
}

// CheckCanAdjust verifies all conditions for making an adjustment.
// Returns (canAdjust, reason).
func (g *Guardrails) CheckCanAdjust(tradeCount int, lastAdjustment time.Time) (bool, string) {
	// Check minimum trades
	if tradeCount < g.minTrades {
		return false, "insufficient_trades"
	}

	// Check cooldown period
	if !lastAdjustment.IsZero() && time.Since(lastAdjustment) < g.cooldown {
		return false, "cooldown_active"
	}

	return true, ""
}

// CheckDrawdown determines if current drawdown exceeds the revert threshold.
// Returns true if we should revert to defaults.
func (g *Guardrails) CheckDrawdown(currentBankroll, peakBankroll float64) bool {
	if peakBankroll <= 0 {
		return false
	}

	drawdown := (peakBankroll - currentBankroll) / peakBankroll
	return drawdown >= g.revertPct
}

// DefaultParameters returns the default parameter values for reversion.
func DefaultParameters() map[string]float64 {
	return map[string]float64{
		"probability_threshold":    0.80,
		"volatility_safety_margin": 1.5,
		"stop_loss_percent":        0.15,
		"kelly_fraction":           0.25,
	}
}
