package dashboard

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"prediction-bot/internal/dashboard/views"
)

// tickMsg is sent on each tick to update the timestamp
type tickMsg time.Time

// quitMsg is sent when the user wants to quit
type quitMsg struct{}

// dataUpdateMsg is sent when data is refreshed
type dataUpdateMsg struct {
	bankrolls []views.BankrollData
	positions []views.PositionData
	stats     views.StatsData
}

// DataProvider defines the interface for fetching dashboard data.
type DataProvider interface {
	GetBankrolls() ([]views.BankrollData, error)
	GetPositions() ([]views.PositionData, error)
	GetStats() (views.StatsData, error)
}

// Model represents the dashboard state
type Model struct {
	lastUpdate    time.Time
	quitting      bool
	paused        bool
	width         int
	height        int
	dryRun        bool
	bankrolls     []views.BankrollData
	positions     []views.PositionData
	stats         views.StatsData
	bankrollView  *views.BankrollView
	positionsView *views.PositionsView
	statsView     *views.StatsView
	keyMap        KeyMap
	dataProvider  DataProvider
	err           error
}

// NewModel creates a new dashboard model
func NewModel() Model {
	return Model{
		lastUpdate:    time.Now(),
		quitting:      false,
		paused:        false,
		width:         80,
		height:        24,
		dryRun:        true,
		bankrollView:  views.NewBankrollView(),
		positionsView: views.NewPositionsView(),
		statsView:     views.NewStatsView(),
		keyMap:        DefaultKeyMap(),
	}
}

// NewModelWithProvider creates a new dashboard model with a data provider.
func NewModelWithProvider(provider DataProvider, dryRun bool) Model {
	m := NewModel()
	m.dataProvider = provider
	m.dryRun = dryRun
	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.fetchDataCmd())
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "r":
			// Manual refresh
			return m, m.fetchDataCmd()
		case "p":
			// Toggle pause
			m.paused = !m.paused
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.lastUpdate = time.Time(msg)
		if m.paused {
			// Only tick, don't fetch data when paused
			return m, tickCmd()
		}
		return m, tea.Batch(tickCmd(), m.fetchDataCmd())

	case dataUpdateMsg:
		m.bankrolls = msg.bankrolls
		m.positions = msg.positions
		m.stats = msg.stats
		m.err = nil
		return m, nil

	case quitMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginBottom(1)

	timestampStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	// Header
	title := titleStyle.Render("Prediction Market Bot")
	timestamp := timestampStyle.Render(fmt.Sprintf("Last Update: %s", m.lastUpdate.Format("15:04:05")))

	// Status indicators
	var statusParts []string

	// Mode indicator
	if m.dryRun {
		statusParts = append(statusParts, statusStyle.Render("[DRY-RUN]"))
	} else {
		statusParts = append(statusParts, lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Render("[LIVE]"))
	}

	// Paused indicator
	if m.paused {
		pausedStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")). // Orange
			Blink(true)
		statusParts = append(statusParts, pausedStyle.Render("[PAUSED]"))
	}

	statusText := ""
	for i, part := range statusParts {
		if i > 0 {
			statusText += " "
		}
		statusText += part
	}

	header := fmt.Sprintf("%s %s\n%s", title, statusText, timestamp)

	// Calculate available width for sections
	sectionWidth := m.width - 2
	if sectionWidth < 40 {
		sectionWidth = 40
	}

	// Bankroll section
	bankrollSection := m.bankrollView.Render(m.bankrolls, sectionWidth)

	// Positions section
	positionsSection := m.positionsView.Render(m.positions, sectionWidth)

	// Stats section
	statsSection := m.statsView.Render(m.stats, sectionWidth)

	// Help text using keymap
	help := helpStyle.Render(m.keyMap.HelpView())

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n\n%s\n\n%s\n",
		header, bankrollSection, positionsSection, statsSection, help)
}

// tickCmd returns a command that sends a tick message after 1 second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchDataCmd returns a command that fetches data from the provider.
func (m Model) fetchDataCmd() tea.Cmd {
	if m.dataProvider == nil {
		return nil
	}

	return func() tea.Msg {
		bankrolls, _ := m.dataProvider.GetBankrolls()
		positions, _ := m.dataProvider.GetPositions()
		stats, _ := m.dataProvider.GetStats()

		return dataUpdateMsg{
			bankrolls: bankrolls,
			positions: positions,
			stats:     stats,
		}
	}
}
