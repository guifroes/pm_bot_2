package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatsData represents trading statistics for display.
type StatsData struct {
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	TotalPnL      float64
	RealizedPnL   float64
	UnrealizedPnL float64
	MaxDrawdown   float64 // As a decimal (0.15 = 15%)
}

// WinRate calculates the win rate as a percentage.
func (s StatsData) WinRate() float64 {
	if s.TotalTrades == 0 {
		return 0
	}
	return (float64(s.WinningTrades) / float64(s.TotalTrades)) * 100
}

// StatsView renders trading statistics.
type StatsView struct {
	titleStyle    lipgloss.Style
	boxStyle      lipgloss.Style
	labelStyle    lipgloss.Style
	valueStyle    lipgloss.Style
	positiveStyle lipgloss.Style
	negativeStyle lipgloss.Style
	neutralStyle  lipgloss.Style
	warningStyle  lipgloss.Style
}

// NewStatsView creates a new StatsView with default styles.
func NewStatsView() *StatsView {
	return &StatsView{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1),
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		labelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(16),
		valueStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")),
		positiveStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42")), // Green
		negativeStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")), // Red
		neutralStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")), // Gray
		warningStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")), // Orange
	}
}

// Render renders the stats view with the given data.
func (v *StatsView) Render(stats StatsData, width int) string {
	title := v.titleStyle.Render("Statistics")

	var lines []string

	// Trades row
	lines = append(lines, v.renderTradesRow(stats))

	// Win rate row
	lines = append(lines, v.renderWinRateRow(stats))

	// Separator
	lines = append(lines, strings.Repeat("─", width-6))

	// PnL rows
	lines = append(lines, v.renderPnLRow("Total PnL", stats.TotalPnL))
	lines = append(lines, v.renderPnLRow("Realized", stats.RealizedPnL))
	lines = append(lines, v.renderPnLRow("Unrealized", stats.UnrealizedPnL))

	// Separator
	lines = append(lines, strings.Repeat("─", width-6))

	// Drawdown row
	lines = append(lines, v.renderDrawdownRow(stats))

	content := strings.Join(lines, "\n")
	return fmt.Sprintf("%s\n%s", title, v.boxStyle.Width(width-4).Render(content))
}

// renderTradesRow renders the trades count row.
func (v *StatsView) renderTradesRow(stats StatsData) string {
	label := v.labelStyle.Render("Trades")

	winsStr := v.positiveStyle.Render(fmt.Sprintf("%dW", stats.WinningTrades))
	lossesStr := v.negativeStyle.Render(fmt.Sprintf("%dL", stats.LosingTrades))
	totalStr := v.valueStyle.Render(fmt.Sprintf("%d", stats.TotalTrades))

	return fmt.Sprintf("%s %s (%s / %s)", label, totalStr, winsStr, lossesStr)
}

// renderWinRateRow renders the win rate row.
func (v *StatsView) renderWinRateRow(stats StatsData) string {
	label := v.labelStyle.Render("Win Rate")

	winRate := stats.WinRate()
	var rateStyle lipgloss.Style

	switch {
	case winRate >= 60:
		rateStyle = v.positiveStyle
	case winRate >= 40:
		rateStyle = v.warningStyle
	default:
		rateStyle = v.negativeStyle
	}

	rateStr := rateStyle.Render(fmt.Sprintf("%.1f%%", winRate))

	return fmt.Sprintf("%s %s", label, rateStr)
}

// renderPnLRow renders a P&L row with appropriate coloring.
func (v *StatsView) renderPnLRow(labelText string, pnl float64) string {
	label := v.labelStyle.Render(labelText)

	var pnlStr string
	if pnl > 0 {
		pnlStr = v.positiveStyle.Render(fmt.Sprintf("+$%.2f", pnl))
	} else if pnl < 0 {
		pnlStr = v.negativeStyle.Render(fmt.Sprintf("-$%.2f", -pnl))
	} else {
		pnlStr = v.neutralStyle.Render("$0.00")
	}

	return fmt.Sprintf("%s %s", label, pnlStr)
}

// renderDrawdownRow renders the max drawdown row.
func (v *StatsView) renderDrawdownRow(stats StatsData) string {
	label := v.labelStyle.Render("Max Drawdown")

	drawdownPct := stats.MaxDrawdown * 100

	var ddStyle lipgloss.Style
	switch {
	case drawdownPct >= 20:
		ddStyle = v.negativeStyle
	case drawdownPct >= 10:
		ddStyle = v.warningStyle
	default:
		ddStyle = v.neutralStyle
	}

	ddStr := ddStyle.Render(fmt.Sprintf("%.1f%%", drawdownPct))

	return fmt.Sprintf("%s %s", label, ddStr)
}
