package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// PositionData represents position information for display.
type PositionData struct {
	ID           int64
	Platform     string
	MarketTitle  string
	Asset        string
	EntryPrice   float64
	CurrentPrice float64
	Quantity     float64
	Side         string
	EntryTime    time.Time
}

// UnrealizedPnL calculates the unrealized profit/loss.
func (p PositionData) UnrealizedPnL() float64 {
	return (p.CurrentPrice - p.EntryPrice) * p.Quantity
}

// UnrealizedPnLPercent calculates the unrealized P&L as percentage.
func (p PositionData) UnrealizedPnLPercent() float64 {
	if p.EntryPrice == 0 {
		return 0
	}
	return ((p.CurrentPrice - p.EntryPrice) / p.EntryPrice) * 100
}

// HoldingTime returns the duration since entry.
func (p PositionData) HoldingTime() time.Duration {
	return time.Since(p.EntryTime)
}

// PositionsView renders positions information.
type PositionsView struct {
	titleStyle    lipgloss.Style
	boxStyle      lipgloss.Style
	headerStyle   lipgloss.Style
	rowStyle      lipgloss.Style
	positiveStyle lipgloss.Style
	negativeStyle lipgloss.Style
	neutralStyle  lipgloss.Style
	assetStyle    lipgloss.Style
	platformStyle lipgloss.Style
}

// NewPositionsView creates a new PositionsView with default styles.
func NewPositionsView() *PositionsView {
	return &PositionsView{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1),
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("241")),
		rowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")),
		positiveStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42")), // Green
		negativeStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")), // Red
		neutralStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")), // Gray
		assetStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")), // Orange
		platformStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")), // Blue
	}
}

// Render renders the positions view with the given data.
func (v *PositionsView) Render(positions []PositionData, width int) string {
	title := v.titleStyle.Render("Open Positions")

	if len(positions) == 0 {
		content := v.neutralStyle.Render("No open positions")
		return fmt.Sprintf("%s\n%s", title, v.boxStyle.Width(width-4).Render(content))
	}

	var lines []string

	// Header
	header := v.renderHeader()
	lines = append(lines, header)
	lines = append(lines, strings.Repeat("─", width-6))

	// Position rows
	var totalPnL float64
	for _, pos := range positions {
		line := v.renderPositionRow(pos, width)
		lines = append(lines, line)
		totalPnL += pos.UnrealizedPnL()
	}

	// Total PnL
	lines = append(lines, strings.Repeat("─", width-6))
	lines = append(lines, v.renderTotalPnL(totalPnL))

	content := strings.Join(lines, "\n")
	return fmt.Sprintf("%s\n%s", title, v.boxStyle.Width(width-4).Render(content))
}

// renderHeader renders the table header.
func (v *PositionsView) renderHeader() string {
	return v.headerStyle.Render(
		fmt.Sprintf("%-6s %-10s %-5s %-6s %-6s %-8s %-10s",
			"Plat", "Asset", "Side", "Entry", "Curr", "Qty", "PnL"))
}

// renderPositionRow renders a single position row.
func (v *PositionsView) renderPositionRow(pos PositionData, width int) string {
	// Platform (abbreviated)
	platform := abbreviatePlatform(pos.Platform)
	platformStr := v.platformStyle.Render(fmt.Sprintf("%-6s", platform))

	// Asset
	asset := pos.Asset
	if asset == "" {
		asset = truncateTitle(pos.MarketTitle, 10)
	}
	assetStr := v.assetStyle.Render(fmt.Sprintf("%-10s", truncateString(asset, 10)))

	// Side
	side := v.rowStyle.Render(fmt.Sprintf("%-5s", pos.Side))

	// Entry price
	entry := v.rowStyle.Render(fmt.Sprintf("$%.2f", pos.EntryPrice))

	// Current price
	current := v.rowStyle.Render(fmt.Sprintf("$%.2f", pos.CurrentPrice))

	// Quantity
	qty := v.rowStyle.Render(fmt.Sprintf("%-8.1f", pos.Quantity))

	// PnL with color
	pnl := pos.UnrealizedPnL()
	var pnlStr string
	if pnl > 0 {
		pnlStr = v.positiveStyle.Render(fmt.Sprintf("+$%.2f", pnl))
	} else if pnl < 0 {
		pnlStr = v.negativeStyle.Render(fmt.Sprintf("-$%.2f", -pnl))
	} else {
		pnlStr = v.neutralStyle.Render("$0.00")
	}

	return fmt.Sprintf("%s %s %s %-6s %-6s %s %s",
		platformStr, assetStr, side, entry, current, qty, pnlStr)
}

// renderTotalPnL renders the total P&L line.
func (v *PositionsView) renderTotalPnL(totalPnL float64) string {
	label := v.headerStyle.Render("Total Unrealized PnL:")

	var pnlStr string
	if totalPnL > 0 {
		pnlStr = v.positiveStyle.Render(fmt.Sprintf("+$%.2f", totalPnL))
	} else if totalPnL < 0 {
		pnlStr = v.negativeStyle.Render(fmt.Sprintf("-$%.2f", -totalPnL))
	} else {
		pnlStr = v.neutralStyle.Render("$0.00")
	}

	return fmt.Sprintf("%s %s", label, pnlStr)
}

// abbreviatePlatform returns an abbreviated platform name.
func abbreviatePlatform(platform string) string {
	switch strings.ToLower(platform) {
	case "polymarket":
		return "POLY"
	case "kalshi":
		return "KALSH"
	default:
		if len(platform) > 5 {
			return strings.ToUpper(platform[:5])
		}
		return strings.ToUpper(platform)
	}
}

// truncateString truncates a string to maxLen characters with ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// truncateTitle extracts a meaningful short form from a market title.
func truncateTitle(title string, maxLen int) string {
	// Try to extract key terms
	lowerTitle := strings.ToLower(title)

	// Check for common assets
	assets := []string{"bitcoin", "btc", "ethereum", "eth", "solana", "sol", "s&p", "spy"}
	for _, asset := range assets {
		if strings.Contains(lowerTitle, asset) {
			return strings.ToUpper(asset)[:min(len(asset), maxLen)]
		}
	}

	// Fallback to truncating the title
	return truncateString(title, maxLen)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
