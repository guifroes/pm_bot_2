package views

import (
	"strings"
	"testing"
)

func TestStatsView_Render_EmptyStats(t *testing.T) {
	view := NewStatsView()
	stats := StatsData{}

	result := view.Render(stats, 60)

	if !strings.Contains(result, "Statistics") {
		t.Error("expected title 'Statistics' in output")
	}

	// Should show zeros for empty stats
	if !strings.Contains(result, "0") {
		t.Error("expected zeros in output for empty stats")
	}
}

func TestStatsView_Render_WithData(t *testing.T) {
	view := NewStatsView()
	stats := StatsData{
		TotalTrades:   25,
		WinningTrades: 18,
		LosingTrades:  7,
		TotalPnL:      150.50,
		RealizedPnL:   120.00,
		UnrealizedPnL: 30.50,
		MaxDrawdown:   0.12,
	}

	result := view.Render(stats, 60)

	// Title
	if !strings.Contains(result, "Statistics") {
		t.Error("expected title 'Statistics' in output")
	}

	// Trades count
	if !strings.Contains(result, "25") {
		t.Errorf("expected total trades '25' in output, got: %s", result)
	}

	// Win rate should be 72% (18/25)
	if !strings.Contains(result, "72") {
		t.Errorf("expected win rate '72' in output, got: %s", result)
	}
}

func TestStatsData_WinRate(t *testing.T) {
	tests := []struct {
		name     string
		stats    StatsData
		expected float64
	}{
		{
			name:     "zero trades",
			stats:    StatsData{TotalTrades: 0},
			expected: 0,
		},
		{
			name: "all wins",
			stats: StatsData{
				TotalTrades:   10,
				WinningTrades: 10,
			},
			expected: 100.0,
		},
		{
			name: "all losses",
			stats: StatsData{
				TotalTrades:   10,
				WinningTrades: 0,
			},
			expected: 0.0,
		},
		{
			name: "mixed",
			stats: StatsData{
				TotalTrades:   20,
				WinningTrades: 15,
			},
			expected: 75.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.WinRate()
			if result != tt.expected {
				t.Errorf("WinRate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStatsView_Render_PositivePnL_Green(t *testing.T) {
	view := NewStatsView()
	stats := StatsData{
		TotalTrades:   10,
		WinningTrades: 8,
		LosingTrades:  2,
		TotalPnL:      250.75,
		RealizedPnL:   250.75,
	}

	result := view.Render(stats, 60)

	// Should contain positive PnL indicator
	if !strings.Contains(result, "+") || !strings.Contains(result, "250") {
		t.Errorf("expected positive PnL with '+' sign, got: %s", result)
	}
}

func TestStatsView_Render_NegativePnL_Red(t *testing.T) {
	view := NewStatsView()
	stats := StatsData{
		TotalTrades:   10,
		WinningTrades: 3,
		LosingTrades:  7,
		TotalPnL:      -150.00,
		RealizedPnL:   -150.00,
	}

	result := view.Render(stats, 60)

	// Should contain negative PnL indicator
	if !strings.Contains(result, "-") || !strings.Contains(result, "150") {
		t.Errorf("expected negative PnL with '-' sign, got: %s", result)
	}
}

func TestStatsView_Render_DrawdownDisplay(t *testing.T) {
	view := NewStatsView()
	stats := StatsData{
		TotalTrades: 10,
		MaxDrawdown: 0.15, // 15%
	}

	result := view.Render(stats, 60)

	// Should contain drawdown percentage
	if !strings.Contains(result, "15") {
		t.Errorf("expected drawdown '15' in output, got: %s", result)
	}
}
