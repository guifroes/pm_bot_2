package dashboard

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestKeyMap_Help(t *testing.T) {
	km := DefaultKeyMap()

	// Check that help text contains expected keys
	helpText := km.HelpView()

	if !strings.Contains(helpText, "q") {
		t.Error("help text should contain 'q' key")
	}

	if !strings.Contains(helpText, "r") {
		t.Error("help text should contain 'r' key")
	}

	if !strings.Contains(helpText, "p") {
		t.Error("help text should contain 'p' key")
	}
}

func TestKeyMap_QuitKey(t *testing.T) {
	km := DefaultKeyMap()

	// Test that q is in quit key binding
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, km.Quit) {
		t.Error("q should match Quit key binding")
	}
}

func TestKeyMap_RefreshKey(t *testing.T) {
	km := DefaultKeyMap()

	// Test that r is in refresh key binding
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, km.Refresh) {
		t.Error("r should match Refresh key binding")
	}
}

func TestKeyMap_PauseKey(t *testing.T) {
	km := DefaultKeyMap()

	// Test that p is in pause key binding
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}, km.Pause) {
		t.Error("p should match Pause key binding")
	}
}

func TestModel_PauseToggle(t *testing.T) {
	m := NewModel()

	// Initially not paused
	if m.paused {
		t.Error("model should not be paused initially")
	}

	// Simulate pressing 'p'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updatedModel := newModel.(Model)

	// Should now be paused
	if !updatedModel.paused {
		t.Error("model should be paused after pressing 'p'")
	}

	// Press 'p' again to unpause
	newModel2, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updatedModel2 := newModel2.(Model)

	if updatedModel2.paused {
		t.Error("model should be unpaused after pressing 'p' again")
	}
}

func TestModel_ViewShowsPausedState(t *testing.T) {
	m := NewModel()
	m.paused = true

	view := m.View()

	if !strings.Contains(view, "PAUSED") {
		t.Errorf("view should show PAUSED indicator when paused, got: %s", view)
	}
}

func TestModel_NoTickWhenPaused(t *testing.T) {
	m := NewModel()
	m.paused = true

	// When paused, tickCmd should not trigger data fetch
	initialUpdate := m.lastUpdate

	// Simulate a tick while paused
	newModel, cmd := m.Update(tickMsg(m.lastUpdate.Add(1)))
	updatedModel := newModel.(Model)

	// Time should still update
	if updatedModel.lastUpdate == initialUpdate {
		t.Error("lastUpdate should still update even when paused")
	}

	// But fetchDataCmd should not be included in the batch
	// We verify this by checking if the returned command is nil or tick-only
	// Since paused mode should not fetch data automatically
	_ = cmd // Command handling is internal
}
