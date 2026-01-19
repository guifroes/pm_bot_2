package dashboard

import (
	"strings"
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("expected app to not be nil")
	}
}

func TestModelInit(t *testing.T) {
	model := NewModel()
	cmd := model.Init()

	// Init should return a tick command for timestamp updates
	if cmd == nil {
		t.Error("expected Init to return a command for timestamp updates")
	}
}

func TestModelView(t *testing.T) {
	model := NewModel()
	view := model.View()

	// View should contain header with title
	if !strings.Contains(view, "Prediction Market Bot") {
		t.Error("expected view to contain 'Prediction Market Bot' title")
	}

	// View should contain timestamp section
	if !strings.Contains(view, "Last Update:") {
		t.Error("expected view to contain 'Last Update:' timestamp")
	}
}

func TestModelViewShowsTimestamp(t *testing.T) {
	model := NewModel()
	model.lastUpdate = time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)

	view := model.View()

	// Should show the timestamp in some format
	if !strings.Contains(view, "14:30:45") {
		t.Errorf("expected view to contain timestamp '14:30:45', got: %s", view)
	}
}

func TestModelUpdate_TickMessage(t *testing.T) {
	model := NewModel()
	initialTime := model.lastUpdate

	// Simulate a tick message
	newModel, _ := model.Update(tickMsg(time.Now()))

	updatedModel := newModel.(Model)
	if !updatedModel.lastUpdate.After(initialTime) {
		t.Error("expected lastUpdate to be updated after tick message")
	}
}

func TestModelUpdate_QuitMessage(t *testing.T) {
	model := NewModel()

	// Simulate pressing 'q'
	newModel, cmd := model.Update(quitMsg{})

	if cmd == nil {
		t.Error("expected quit command to be returned")
	}

	updatedModel := newModel.(Model)
	if !updatedModel.quitting {
		t.Error("expected quitting flag to be true")
	}
}
