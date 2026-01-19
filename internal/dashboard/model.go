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
}

// DataProvider defines the interface for fetching dashboard data.
type DataProvider interface {
	GetBankrolls() ([]views.BankrollData, error)
	GetPositions() ([]views.PositionData, error)
}

// Model represents the dashboard state
type Model struct {
	lastUpdate    time.Time
	quitting      bool
	width         int
	height        int
	dryRun        bool
	bankrolls     []views.BankrollData
	positions     []views.PositionData
	bankrollView  *views.BankrollView
	positionsView *views.PositionsView
	dataProvider  DataProvider
	err           error
}

// NewModel creates a new dashboard model
func NewModel() Model {
	return Model{
		lastUpdate:    time.Now(),
		quitting:      false,
		width:         80,
		height:        24,
		dryRun:        true,
		bankrollView:  views.NewBankrollView(),
		positionsView: views.NewPositionsView(),
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
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.lastUpdate = time.Time(msg)
		return m, tea.Batch(tickCmd(), m.fetchDataCmd())

	case dataUpdateMsg:
		m.bankrolls = msg.bankrolls
		m.positions = msg.positions
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

	// Status indicator
	var statusText string
	if m.dryRun {
		statusText = statusStyle.Render("[DRY-RUN]")
	} else {
		statusText = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Render("[LIVE]")
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

	// Help text
	help := helpStyle.Render("Press 'q' to quit, 'r' to refresh")

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n\n%s\n",
		header, bankrollSection, positionsSection, help)
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

		return dataUpdateMsg{
			bankrolls: bankrolls,
			positions: positions,
		}
	}
}
