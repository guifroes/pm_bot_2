package dashboard

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tickMsg is sent on each tick to update the timestamp
type tickMsg time.Time

// quitMsg is sent when the user wants to quit
type quitMsg struct{}

// Model represents the dashboard state
type Model struct {
	lastUpdate time.Time
	quitting   bool
	width      int
	height     int
}

// NewModel creates a new dashboard model
func NewModel() Model {
	return Model{
		lastUpdate: time.Now(),
		quitting:   false,
		width:      80,
		height:     24,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.lastUpdate = time.Time(msg)
		return m, tickCmd()

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

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	// Header
	title := titleStyle.Render("📊 Prediction Market Bot")
	timestamp := timestampStyle.Render(fmt.Sprintf("Last Update: %s", m.lastUpdate.Format("15:04:05")))

	header := fmt.Sprintf("%s\n%s", title, timestamp)

	// Status section placeholder
	statusContent := "Status: Running (DRY-RUN mode)\n\nPress 'q' to quit"
	status := boxStyle.Render(statusContent)

	return fmt.Sprintf("\n%s\n\n%s\n", header, status)
}

// tickCmd returns a command that sends a tick message after 1 second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
