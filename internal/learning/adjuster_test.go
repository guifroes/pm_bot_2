package learning

import (
	"testing"
	"time"
)

func TestSuggestAdjustment_MovesTowardsBestSegment(t *testing.T) {
	adj := NewAdjuster()

	// Current probability threshold is 0.80
	// Best performing segment is 0.85-0.90 (higher probability)
	current := 0.80
	segments := []SegmentStats{
		{ParamName: "probability", RangeStart: 0.80, RangeEnd: 0.85, TradeCount: 10, WinRate: 0.60, AvgPnL: 0.5},
		{ParamName: "probability", RangeStart: 0.85, RangeEnd: 0.90, TradeCount: 10, WinRate: 0.85, AvgPnL: 2.5},
		{ParamName: "probability", RangeStart: 0.90, RangeEnd: 0.95, TradeCount: 10, WinRate: 0.70, AvgPnL: 1.0},
	}
	bounds := AdjustmentBounds{Min: 0.75, Max: 0.95}

	newValue := adj.SuggestAdjustment(current, segments, bounds)

	// Should move toward the best segment (0.85-0.90 center = 0.875)
	// Max 10% adjustment means: 0.80 * 1.10 = 0.88 max
	// So new value should be > 0.80 but within 10%
	if newValue <= current {
		t.Errorf("expected increase from %v, got %v", current, newValue)
	}
	if newValue > 0.88 {
		t.Errorf("expected max 10%% adjustment, got %v (from %v)", newValue, current)
	}
	if newValue < bounds.Min || newValue > bounds.Max {
		t.Errorf("value %v outside bounds [%v, %v]", newValue, bounds.Min, bounds.Max)
	}
}

func TestSuggestAdjustment_DecreasesWhenLowerIsBetter(t *testing.T) {
	adj := NewAdjuster()

	// Current safety margin threshold is 1.5
	// Best performing segment is 0.8-1.2 (lower margin)
	current := 1.5
	segments := []SegmentStats{
		{ParamName: "safety_margin", RangeStart: 0.8, RangeEnd: 1.2, TradeCount: 10, WinRate: 0.90, AvgPnL: 3.0},
		{ParamName: "safety_margin", RangeStart: 1.2, RangeEnd: 1.5, TradeCount: 10, WinRate: 0.70, AvgPnL: 1.5},
		{ParamName: "safety_margin", RangeStart: 1.5, RangeEnd: 2.0, TradeCount: 10, WinRate: 0.60, AvgPnL: 0.5},
	}
	bounds := AdjustmentBounds{Min: 1.0, Max: 3.0}

	newValue := adj.SuggestAdjustment(current, segments, bounds)

	// Should move toward the best segment (0.8-1.2 center = 1.0)
	// But within 10% decrease: 1.5 * 0.90 = 1.35 min
	if newValue >= current {
		t.Errorf("expected decrease from %v, got %v", current, newValue)
	}
	if newValue < 1.35 {
		t.Errorf("expected max 10%% adjustment, got %v (from %v)", newValue, current)
	}
}

func TestSuggestAdjustment_RespectsMaxAdjustmentLimit(t *testing.T) {
	adj := NewAdjuster()

	// Large gap between current and best segment
	current := 0.80
	segments := []SegmentStats{
		{ParamName: "probability", RangeStart: 0.95, RangeEnd: 1.00, TradeCount: 10, WinRate: 0.95, AvgPnL: 5.0},
	}
	bounds := AdjustmentBounds{Min: 0.75, Max: 0.99}

	newValue := adj.SuggestAdjustment(current, segments, bounds)

	// Max 10% adjustment: 0.80 * 1.10 = 0.88
	if newValue > 0.88 {
		t.Errorf("exceeded 10%% adjustment limit: got %v from %v", newValue, current)
	}
}

func TestSuggestAdjustment_RespectsBounds(t *testing.T) {
	adj := NewAdjuster()

	current := 0.94
	segments := []SegmentStats{
		{ParamName: "probability", RangeStart: 0.95, RangeEnd: 1.00, TradeCount: 10, WinRate: 0.90, AvgPnL: 3.0},
	}
	bounds := AdjustmentBounds{Min: 0.75, Max: 0.95}

	newValue := adj.SuggestAdjustment(current, segments, bounds)

	// Cannot exceed bounds.Max
	if newValue > bounds.Max {
		t.Errorf("exceeded max bound: got %v, max %v", newValue, bounds.Max)
	}
}

func TestSuggestAdjustment_NoChangeWithEmptySegments(t *testing.T) {
	adj := NewAdjuster()

	current := 0.80
	segments := []SegmentStats{}
	bounds := AdjustmentBounds{Min: 0.75, Max: 0.95}

	newValue := adj.SuggestAdjustment(current, segments, bounds)

	// No data = no change
	if newValue != current {
		t.Errorf("expected no change with empty segments, got %v from %v", newValue, current)
	}
}

func TestSuggestAdjustment_NoChangeWithInsufficientData(t *testing.T) {
	adj := NewAdjuster()

	// All segments have fewer than 5 trades
	current := 0.80
	segments := []SegmentStats{
		{ParamName: "probability", RangeStart: 0.85, RangeEnd: 0.90, TradeCount: 3, WinRate: 0.90, AvgPnL: 5.0},
	}
	bounds := AdjustmentBounds{Min: 0.75, Max: 0.95}

	newValue := adj.SuggestAdjustment(current, segments, bounds)

	// Insufficient data = no change
	if newValue != current {
		t.Errorf("expected no change with insufficient data, got %v from %v", newValue, current)
	}
}

func TestSuggestAdjustment_CurrentInBestSegment(t *testing.T) {
	adj := NewAdjuster()

	// Current is already in the best segment
	current := 0.87
	segments := []SegmentStats{
		{ParamName: "probability", RangeStart: 0.80, RangeEnd: 0.85, TradeCount: 10, WinRate: 0.60, AvgPnL: 0.5},
		{ParamName: "probability", RangeStart: 0.85, RangeEnd: 0.90, TradeCount: 10, WinRate: 0.85, AvgPnL: 2.5},
	}
	bounds := AdjustmentBounds{Min: 0.75, Max: 0.95}

	newValue := adj.SuggestAdjustment(current, segments, bounds)

	// Already optimal = no change
	if newValue != current {
		t.Errorf("expected no change when already in best segment, got %v from %v", newValue, current)
	}
}

// Guardrails tests

func TestGuardrails_CheckCanAdjust_InsufficientTrades(t *testing.T) {
	g := NewGuardrails()

	canAdjust, reason := g.CheckCanAdjust(10, time.Time{})

	if canAdjust {
		t.Error("expected false for insufficient trades")
	}
	if reason != "insufficient_trades" {
		t.Errorf("expected reason 'insufficient_trades', got %v", reason)
	}
}

func TestGuardrails_CheckCanAdjust_EnoughTrades(t *testing.T) {
	g := NewGuardrails()

	canAdjust, reason := g.CheckCanAdjust(25, time.Time{})

	if !canAdjust {
		t.Error("expected true for enough trades")
	}
	if reason != "" {
		t.Errorf("expected empty reason, got %v", reason)
	}
}

func TestGuardrails_CheckCanAdjust_CooldownActive(t *testing.T) {
	g := NewGuardrails()

	// Last adjustment was 1 hour ago (cooldown is 24 hours)
	lastAdjustment := time.Now().Add(-1 * time.Hour)
	canAdjust, reason := g.CheckCanAdjust(25, lastAdjustment)

	if canAdjust {
		t.Error("expected false during cooldown")
	}
	if reason != "cooldown_active" {
		t.Errorf("expected reason 'cooldown_active', got %v", reason)
	}
}

func TestGuardrails_CheckCanAdjust_CooldownExpired(t *testing.T) {
	g := NewGuardrails()

	// Last adjustment was 25 hours ago (cooldown is 24 hours)
	lastAdjustment := time.Now().Add(-25 * time.Hour)
	canAdjust, reason := g.CheckCanAdjust(25, lastAdjustment)

	if !canAdjust {
		t.Error("expected true after cooldown expired")
	}
	if reason != "" {
		t.Errorf("expected empty reason, got %v", reason)
	}
}

func TestGuardrails_CheckDrawdown_NoDrawdown(t *testing.T) {
	g := NewGuardrails()

	// Current $100, Peak $100 = no drawdown
	shouldRevert := g.CheckDrawdown(100, 100)

	if shouldRevert {
		t.Error("expected false for no drawdown")
	}
}

func TestGuardrails_CheckDrawdown_SmallDrawdown(t *testing.T) {
	g := NewGuardrails()

	// Current $90, Peak $100 = 10% drawdown (below 20% threshold)
	shouldRevert := g.CheckDrawdown(90, 100)

	if shouldRevert {
		t.Error("expected false for small drawdown")
	}
}

func TestGuardrails_CheckDrawdown_LargeDrawdown(t *testing.T) {
	g := NewGuardrails()

	// Current $75, Peak $100 = 25% drawdown (above 20% threshold)
	shouldRevert := g.CheckDrawdown(75, 100)

	if !shouldRevert {
		t.Error("expected true for large drawdown")
	}
}

func TestGuardrails_CheckDrawdown_ExactThreshold(t *testing.T) {
	g := NewGuardrails()

	// Current $80, Peak $100 = exactly 20% drawdown
	shouldRevert := g.CheckDrawdown(80, 100)

	if !shouldRevert {
		t.Error("expected true at exact threshold (20%)")
	}
}

func TestGuardrails_CheckDrawdown_ZeroPeak(t *testing.T) {
	g := NewGuardrails()

	// Edge case: zero peak
	shouldRevert := g.CheckDrawdown(80, 0)

	if shouldRevert {
		t.Error("expected false for zero peak")
	}
}

func TestDefaultParameters(t *testing.T) {
	defaults := DefaultParameters()

	expected := map[string]float64{
		"probability_threshold":    0.80,
		"volatility_safety_margin": 1.5,
		"stop_loss_percent":        0.15,
		"kelly_fraction":           0.25,
	}

	for name, expectedVal := range expected {
		if val, ok := defaults[name]; !ok {
			t.Errorf("missing parameter %s", name)
		} else if val != expectedVal {
			t.Errorf("parameter %s: expected %v, got %v", name, expectedVal, val)
		}
	}
}
