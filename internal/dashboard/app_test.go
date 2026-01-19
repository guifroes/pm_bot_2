package dashboard

import (
	"strings"
	"testing"
	"time"

	"prediction-bot/internal/dashboard/views"
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

// MockDataProvider implements DataProvider for testing.
type MockDataProvider struct {
	bankrolls []views.BankrollData
	positions []views.PositionData
}

func (m *MockDataProvider) GetBankrolls() ([]views.BankrollData, error) {
	return m.bankrolls, nil
}

func (m *MockDataProvider) GetPositions() ([]views.PositionData, error) {
	return m.positions, nil
}

func TestModelViewShowsBankroll(t *testing.T) {
	model := NewModel()
	model.bankrolls = []views.BankrollData{
		{
			Platform:      "polymarket",
			InitialAmount: 50.00,
			CurrentAmount: 55.00,
		},
	}

	view := model.View()

	// Should show bankroll section
	if !strings.Contains(view, "Bankroll") {
		t.Error("expected view to contain 'Bankroll' section")
	}

	// Should show platform
	if !strings.Contains(strings.ToLower(view), "polymarket") {
		t.Error("expected view to contain platform name")
	}

	// Should show amount
	if !strings.Contains(view, "55") {
		t.Errorf("expected view to contain amount, got: %s", view)
	}
}

func TestModelViewShowsPositions(t *testing.T) {
	model := NewModel()
	model.positions = []views.PositionData{
		{
			ID:           1,
			Platform:     "polymarket",
			MarketTitle:  "Will Bitcoin be above $100k?",
			Asset:        "BTC",
			EntryPrice:   0.85,
			CurrentPrice: 0.90,
			Quantity:     10.0,
			Side:         "YES",
			EntryTime:    time.Now(),
		},
	}

	view := model.View()

	// Should show positions section
	if !strings.Contains(view, "Position") {
		t.Error("expected view to contain 'Position' section")
	}

	// Should show asset
	if !strings.Contains(view, "BTC") {
		t.Error("expected view to contain asset 'BTC'")
	}
}

func TestModelViewShowsNoPositionsMessage(t *testing.T) {
	model := NewModel()
	// No positions set

	view := model.View()

	// Should show "No open positions" message
	if !strings.Contains(strings.ToLower(view), "no") ||
		!strings.Contains(strings.ToLower(view), "position") {
		t.Errorf("expected view to contain 'No open positions' message, got: %s", view)
	}
}

func TestNewModelWithProvider(t *testing.T) {
	provider := &MockDataProvider{
		bankrolls: []views.BankrollData{
			{Platform: "kalshi", InitialAmount: 100, CurrentAmount: 95},
		},
	}

	model := NewModelWithProvider(provider, true)

	if model.dataProvider == nil {
		t.Error("expected model to have data provider")
	}

	if !model.dryRun {
		t.Error("expected dryRun to be true")
	}
}

func TestNewAppWithProvider(t *testing.T) {
	provider := &MockDataProvider{}
	app := NewAppWithProvider(provider, false)

	if app == nil {
		t.Error("expected app to not be nil")
	}
}

func TestModelViewShowsDryRunStatus(t *testing.T) {
	model := NewModel()
	model.dryRun = true

	view := model.View()

	if !strings.Contains(view, "DRY-RUN") {
		t.Errorf("expected view to contain 'DRY-RUN' status, got: %s", view)
	}
}

func TestModelViewShowsLiveStatus(t *testing.T) {
	model := NewModel()
	model.dryRun = false

	view := model.View()

	if !strings.Contains(view, "LIVE") {
		t.Errorf("expected view to contain 'LIVE' status, got: %s", view)
	}
}
