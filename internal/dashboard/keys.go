package dashboard

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap defines the keybindings for the dashboard.
type KeyMap struct {
	Quit    key.Binding
	Refresh key.Binding
	Pause   key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause"),
		),
	}
}

// HelpView returns a formatted help view showing all keybindings.
func (k KeyMap) HelpView() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	separator := helpStyle.Render(" • ")

	items := []string{
		fmt.Sprintf("%s %s", keyStyle.Render("q"), helpStyle.Render("quit")),
		fmt.Sprintf("%s %s", keyStyle.Render("r"), helpStyle.Render("refresh")),
		fmt.Sprintf("%s %s", keyStyle.Render("p"), helpStyle.Render("pause")),
	}

	return fmt.Sprintf("%s%s%s%s%s", items[0], separator, items[1], separator, items[2])
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Refresh, k.Pause}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Refresh, k.Pause},
	}
}
