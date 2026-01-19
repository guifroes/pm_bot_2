package dashboard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// App represents the dashboard application
type App struct {
	program *tea.Program
}

// NewApp creates a new dashboard application
func NewApp() *App {
	model := NewModel()
	program := tea.NewProgram(model, tea.WithAltScreen())

	return &App{
		program: program,
	}
}

// Run starts the dashboard application
func (a *App) Run() error {
	if _, err := a.program.Run(); err != nil {
		return fmt.Errorf("running dashboard: %w", err)
	}
	return nil
}
