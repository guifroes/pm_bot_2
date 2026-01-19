package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BankrollData represents bankroll information for display.
type BankrollData struct {
	Platform      string
	InitialAmount float64
	CurrentAmount float64
}

// Delta returns the change from initial to current amount.
func (b BankrollData) Delta() float64 {
	return b.CurrentAmount - b.InitialAmount
}

// DeltaPercent returns the percentage change from initial.
func (b BankrollData) DeltaPercent() float64 {
	if b.InitialAmount == 0 {
		return 0
	}
	return (b.Delta() / b.InitialAmount) * 100
}

// BankrollView renders bankroll information.
type BankrollView struct {
	titleStyle   lipgloss.Style
	boxStyle     lipgloss.Style
	labelStyle   lipgloss.Style
	valueStyle   lipgloss.Style
	positiveStyle lipgloss.Style
	negativeStyle lipgloss.Style
	neutralStyle  lipgloss.Style
}

// NewBankrollView creates a new BankrollView with default styles.
func NewBankrollView() *BankrollView {
	return &BankrollView{
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
			Width(12),
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
	}
}

// Render renders the bankroll view with the given data.
func (v *BankrollView) Render(data []BankrollData, width int) string {
	title := v.titleStyle.Render("Bankroll")

	if len(data) == 0 {
		content := v.neutralStyle.Render("No bankroll data available")
		return fmt.Sprintf("%s\n%s", title, v.boxStyle.Width(width-4).Render(content))
	}

	var lines []string
	var totalInitial, totalCurrent float64

	for _, b := range data {
		totalInitial += b.InitialAmount
		totalCurrent += b.CurrentAmount

		line := v.renderPlatformLine(b)
		lines = append(lines, line)
	}

	// Add separator and total if multiple platforms
	if len(data) > 1 {
		lines = append(lines, strings.Repeat("─", width-6))
		totalData := BankrollData{
			Platform:      "Total",
			InitialAmount: totalInitial,
			CurrentAmount: totalCurrent,
		}
		lines = append(lines, v.renderPlatformLine(totalData))
	}

	content := strings.Join(lines, "\n")
	return fmt.Sprintf("%s\n%s", title, v.boxStyle.Width(width-4).Render(content))
}

// renderPlatformLine renders a single platform line with current amount and delta.
func (v *BankrollView) renderPlatformLine(b BankrollData) string {
	// Format platform name with capitalization
	platform := strings.Title(strings.ToLower(b.Platform))
	label := v.labelStyle.Render(platform)

	// Format current amount
	amount := v.valueStyle.Render(fmt.Sprintf("$%.2f", b.CurrentAmount))

	// Format delta
	delta := b.Delta()
	var deltaStr string
	if delta > 0 {
		deltaStr = v.positiveStyle.Render(fmt.Sprintf("+$%.2f", delta))
	} else if delta < 0 {
		deltaStr = v.negativeStyle.Render(fmt.Sprintf("-$%.2f", -delta))
	} else {
		deltaStr = v.neutralStyle.Render("$0.00")
	}

	// Format percent change
	pct := b.DeltaPercent()
	var pctStr string
	if pct > 0 {
		pctStr = v.positiveStyle.Render(fmt.Sprintf("(+%.1f%%)", pct))
	} else if pct < 0 {
		pctStr = v.negativeStyle.Render(fmt.Sprintf("(%.1f%%)", pct))
	} else {
		pctStr = v.neutralStyle.Render("(0.0%)")
	}

	return fmt.Sprintf("%s %s  %s %s", label, amount, deltaStr, pctStr)
}
