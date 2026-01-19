package learning

import (
	"testing"
	"time"
)

func TestAnalyzeBySegment_ProbabilitySegments(t *testing.T) {
	// Create test outcomes with various entry prices (probability)
	outcomes := []TradeOutcome{
		// Segment 80-85: 2 wins, 1 loss = 66.67 percent win rate
		{PositionID: 1, EntryPrice: 0.82, RealizedPnL: 10.0, Platform: "poly"},
		{PositionID: 2, EntryPrice: 0.84, RealizedPnL: 5.0, Platform: "poly"},
		{PositionID: 3, EntryPrice: 0.81, RealizedPnL: -8.0, Platform: "poly"},

		// Segment 85-90: 3 wins, 0 losses = 100 percent win rate
		{PositionID: 4, EntryPrice: 0.87, RealizedPnL: 12.0, Platform: "poly"},
		{PositionID: 5, EntryPrice: 0.88, RealizedPnL: 7.0, Platform: "poly"},
		{PositionID: 6, EntryPrice: 0.86, RealizedPnL: 3.0, Platform: "poly"},

		// Segment 90-95: 1 win, 2 losses = 33.33 percent win rate
		{PositionID: 7, EntryPrice: 0.92, RealizedPnL: 20.0, Platform: "poly"},
		{PositionID: 8, EntryPrice: 0.91, RealizedPnL: -15.0, Platform: "poly"},
		{PositionID: 9, EntryPrice: 0.93, RealizedPnL: -10.0, Platform: "poly"},
	}

	analyzer := NewAnalyzer()
	segments := analyzer.AnalyzeBySegment(outcomes, "probability")

	if len(segments) == 0 {
		t.Fatal("expected segments, got empty slice")
	}

	// Find segment 85-90 (should have best win rate)
	var seg8590 *SegmentStats
	for i := range segments {
		if segments[i].RangeStart == 0.85 && segments[i].RangeEnd == 0.90 {
			seg8590 = &segments[i]
			break
		}
	}

	if seg8590 == nil {
		t.Fatal("segment 85-90 not found")
	}

	if seg8590.TradeCount != 3 {
		t.Errorf("segment 85-90 trade count: got %d, want 3", seg8590.TradeCount)
	}

	if seg8590.WinRate != 1.0 {
		t.Errorf("segment 85-90 win rate: got %.2f, want 1.0", seg8590.WinRate)
	}

	expectedAvgPnL := (12.0 + 7.0 + 3.0) / 3.0
	if seg8590.AvgPnL != expectedAvgPnL {
		t.Errorf("segment 85-90 avg PnL: got %.2f, want %.2f", seg8590.AvgPnL, expectedAvgPnL)
	}
}

func TestAnalyzeBySegment_SafetyMarginSegments(t *testing.T) {
	outcomes := []TradeOutcome{
		// Low safety margin (0.8-1.2): 1 win, 2 losses
		{PositionID: 1, SafetyMargin: 0.9, RealizedPnL: 5.0, Platform: "poly"},
		{PositionID: 2, SafetyMargin: 1.0, RealizedPnL: -10.0, Platform: "poly"},
		{PositionID: 3, SafetyMargin: 1.1, RealizedPnL: -5.0, Platform: "poly"},

		// Medium safety margin (1.5-2.0): 2 wins, 1 loss
		{PositionID: 4, SafetyMargin: 1.6, RealizedPnL: 15.0, Platform: "poly"},
		{PositionID: 5, SafetyMargin: 1.8, RealizedPnL: 8.0, Platform: "poly"},
		{PositionID: 6, SafetyMargin: 1.7, RealizedPnL: -3.0, Platform: "poly"},

		// High safety margin (2.0-2.5): 3 wins, 0 losses
		{PositionID: 7, SafetyMargin: 2.1, RealizedPnL: 20.0, Platform: "poly"},
		{PositionID: 8, SafetyMargin: 2.3, RealizedPnL: 12.0, Platform: "poly"},
		{PositionID: 9, SafetyMargin: 2.2, RealizedPnL: 7.0, Platform: "poly"},
	}

	analyzer := NewAnalyzer()
	segments := analyzer.AnalyzeBySegment(outcomes, "safety_margin")

	if len(segments) == 0 {
		t.Fatal("expected segments, got empty slice")
	}

	// High safety margin segment should have best win rate
	var highMarginSeg *SegmentStats
	for i := range segments {
		if segments[i].RangeStart >= 2.0 && segments[i].RangeEnd <= 2.5 {
			highMarginSeg = &segments[i]
			break
		}
	}

	if highMarginSeg == nil {
		t.Fatal("high safety margin segment not found")
	}

	if highMarginSeg.WinRate != 1.0 {
		t.Errorf("high margin segment win rate: got %.2f, want 1.0", highMarginSeg.WinRate)
	}
}

func TestAnalyzeBySegment_EmptyOutcomes(t *testing.T) {
	analyzer := NewAnalyzer()
	segments := analyzer.AnalyzeBySegment([]TradeOutcome{}, "probability")

	if len(segments) != 0 {
		t.Errorf("expected 0 segments for empty outcomes, got %d", len(segments))
	}
}

func TestAnalyzeBySegment_InvalidParam(t *testing.T) {
	outcomes := []TradeOutcome{
		{PositionID: 1, EntryPrice: 0.85, RealizedPnL: 10.0},
	}

	analyzer := NewAnalyzer()
	segments := analyzer.AnalyzeBySegment(outcomes, "unknown_param")

	// Should return empty for unknown parameter
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for unknown param, got %d", len(segments))
	}
}

func TestAnalyzeBySegment_SingleTrade(t *testing.T) {
	outcomes := []TradeOutcome{
		{PositionID: 1, EntryPrice: 0.87, RealizedPnL: 10.0},
	}

	analyzer := NewAnalyzer()
	segments := analyzer.AnalyzeBySegment(outcomes, "probability")

	// Should still work with a single trade
	found := false
	for _, seg := range segments {
		if seg.TradeCount > 0 {
			found = true
			if seg.WinRate != 1.0 {
				t.Errorf("single winning trade should have 100 pct win rate, got %.2f", seg.WinRate)
			}
		}
	}

	if !found {
		t.Error("expected at least one non-empty segment")
	}
}

func TestAnalyzeBySegment_TotalPnL(t *testing.T) {
	outcomes := []TradeOutcome{
		{PositionID: 1, EntryPrice: 0.85, RealizedPnL: 10.0},
		{PositionID: 2, EntryPrice: 0.86, RealizedPnL: -5.0},
		{PositionID: 3, EntryPrice: 0.87, RealizedPnL: 15.0},
	}

	analyzer := NewAnalyzer()
	segments := analyzer.AnalyzeBySegment(outcomes, "probability")

	// Find the segment containing these trades (85-90)
	for _, seg := range segments {
		if seg.TradeCount == 3 {
			expectedTotal := 10.0 + (-5.0) + 15.0
			if seg.TotalPnL != expectedTotal {
				t.Errorf("total PnL: got %.2f, want %.2f", seg.TotalPnL, expectedTotal)
			}
			break
		}
	}
}

func TestSegmentStats_Fields(t *testing.T) {
	now := time.Now()
	outcomes := []TradeOutcome{
		{PositionID: 1, EntryPrice: 0.88, RealizedPnL: 10.0, ExitTime: now},
		{PositionID: 2, EntryPrice: 0.89, RealizedPnL: -2.0, ExitTime: now.Add(-time.Hour)},
	}

	analyzer := NewAnalyzer()
	segments := analyzer.AnalyzeBySegment(outcomes, "probability")

	for _, seg := range segments {
		if seg.TradeCount == 2 {
			// Check all fields are populated
			if seg.ParamName != "probability" {
				t.Errorf("param name: got %s, want probability", seg.ParamName)
			}
			if seg.WinCount != 1 {
				t.Errorf("win count: got %d, want 1", seg.WinCount)
			}
			if seg.LossCount != 1 {
				t.Errorf("loss count: got %d, want 1", seg.LossCount)
			}
			break
		}
	}
}
