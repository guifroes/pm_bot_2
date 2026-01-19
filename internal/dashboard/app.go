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

// NewAppWithProvider creates a new dashboard application with a data provider.
func NewAppWithProvider(provider DataProvider, dryRun bool) *App {
	model := NewModelWithProvider(provider, dryRun)
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
