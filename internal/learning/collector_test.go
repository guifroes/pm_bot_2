package learning

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"prediction-bot/internal/persistence"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test_learning_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := persistence.OpenDB(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to open database: %v", err)
	}

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

func TestCollector_CollectOutcomes_Returns20ClosedTrades(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	posRepo := persistence.NewPositionRepository(db)

	// Create 25 closed positions
	for i := 0; i < 25; i++ {
		entryPrice := 0.85 + float64(i)*0.001
		exitPrice := 0.92 + float64(i)*0.001
		pnl := (exitPrice - entryPrice) * 100

		pos := &persistence.Position{
			Platform:            "polymarket",
			MarketID:            "test-market-" + string(rune('a'+i)),
			MarketTitle:         "Test Market",
			Asset:               "BTC",
			Strike:              100000,
			Direction:           "above",
			EntryPrice:          entryPrice,
			Quantity:            100,
			Side:                "YES",
			Status:              "open",
			SafetyMarginAtEntry: 1.5 + float64(i)*0.1,
			VolatilityAtEntry:   0.5,
		}

		id, err := posRepo.Create(pos)
		if err != nil {
			t.Fatalf("failed to create position: %v", err)
		}

		// Close the position
		err = posRepo.Close(id, exitPrice, "market_resolved", pnl)
		if err != nil {
			t.Fatalf("failed to close position: %v", err)
		}
	}

	// Create collector and collect outcomes
	collector := NewCollector(db)
	outcomes, err := collector.CollectOutcomes(20)
	if err != nil {
		t.Fatalf("CollectOutcomes failed: %v", err)
	}

	// Should return exactly 20 trades (the minimum requested)
	if len(outcomes) != 20 {
		t.Errorf("expected 20 outcomes, got %d", len(outcomes))
	}

	// Verify outcome data
	for _, outcome := range outcomes {
		if outcome.Platform == "" {
			t.Error("expected platform to be set")
		}
		if outcome.Asset == "" {
			t.Error("expected asset to be set")
		}
		if outcome.EntryPrice == 0 {
			t.Error("expected entry price to be set")
		}
		if outcome.ExitPrice == 0 {
			t.Error("expected exit price to be set")
		}
		if outcome.SafetyMargin == 0 {
			t.Error("expected safety margin to be set")
		}
	}
}

func TestCollector_CollectOutcomes_ReturnsEmptyWhenNotEnoughTrades(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	posRepo := persistence.NewPositionRepository(db)

	// Create only 5 closed positions
	for i := 0; i < 5; i++ {
		pos := &persistence.Position{
			Platform:            "kalshi",
			MarketID:            "test-market-" + string(rune('a'+i)),
			MarketTitle:         "Test Market",
			Asset:               "ETH",
			Strike:              3000,
			Direction:           "below",
			EntryPrice:          0.80,
			Quantity:            50,
			Side:                "NO",
			Status:              "open",
			SafetyMarginAtEntry: 1.2,
			VolatilityAtEntry:   0.6,
		}

		id, err := posRepo.Create(pos)
		if err != nil {
			t.Fatalf("failed to create position: %v", err)
		}

		err = posRepo.Close(id, 0.75, "stop_loss", -2.50)
		if err != nil {
			t.Fatalf("failed to close position: %v", err)
		}
	}

	collector := NewCollector(db)
	outcomes, err := collector.CollectOutcomes(20)
	if err != nil {
		t.Fatalf("CollectOutcomes failed: %v", err)
	}

	// Should return empty when not enough trades
	if len(outcomes) != 0 {
		t.Errorf("expected 0 outcomes when less than minTrades, got %d", len(outcomes))
	}
}

func TestCollector_CollectOutcomes_ExcludesOpenPositions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	posRepo := persistence.NewPositionRepository(db)

	// Create 25 positions, but only close 10
	for i := 0; i < 25; i++ {
		pos := &persistence.Position{
			Platform:            "polymarket",
			MarketID:            "test-market-" + string(rune('a'+i)),
			MarketTitle:         "Test Market",
			Asset:               "BTC",
			Strike:              100000,
			Direction:           "above",
			EntryPrice:          0.85,
			Quantity:            100,
			Side:                "YES",
			Status:              "open",
			SafetyMarginAtEntry: 1.5,
			VolatilityAtEntry:   0.5,
		}

		id, err := posRepo.Create(pos)
		if err != nil {
			t.Fatalf("failed to create position: %v", err)
		}

		// Only close the first 10
		if i < 10 {
			err = posRepo.Close(id, 0.92, "market_resolved", 7.0)
			if err != nil {
				t.Fatalf("failed to close position: %v", err)
			}
		}
	}

	collector := NewCollector(db)
	outcomes, err := collector.CollectOutcomes(5)
	if err != nil {
		t.Fatalf("CollectOutcomes failed: %v", err)
	}

	// Should return exactly minTrades=5 closed trades (the most recent ones)
	if len(outcomes) != 5 {
		t.Errorf("expected 5 closed outcomes (limited by minTrades), got %d", len(outcomes))
	}
}

func TestTradeOutcome_IsWin(t *testing.T) {
	winningTrade := TradeOutcome{
		RealizedPnL: 5.50,
	}
	if !winningTrade.IsWin() {
		t.Error("expected positive PnL to be a win")
	}

	losingTrade := TradeOutcome{
		RealizedPnL: -2.00,
	}
	if losingTrade.IsWin() {
		t.Error("expected negative PnL to be a loss")
	}

	breakEvenTrade := TradeOutcome{
		RealizedPnL: 0,
	}
	if breakEvenTrade.IsWin() {
		t.Error("expected zero PnL to not be a win")
	}
}

func TestTradeOutcome_ReturnPercent(t *testing.T) {
	trade := TradeOutcome{
		EntryPrice: 0.80,
		ExitPrice:  0.92,
		Quantity:   100,
	}

	// Return % = (exit - entry) / entry * 100
	expected := (0.92 - 0.80) / 0.80 * 100
	got := trade.ReturnPercent()

	if got < expected-0.01 || got > expected+0.01 {
		t.Errorf("expected return percent %.2f%%, got %.2f%%", expected, got)
	}
}

func TestCollector_CollectOutcomes_IncludesParametersUsed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	posRepo := persistence.NewPositionRepository(db)

	// Create closed positions with varied parameters
	for i := 0; i < 25; i++ {
		safetyMargin := 1.0 + float64(i)*0.1
		volatility := 0.3 + float64(i)*0.02

		pos := &persistence.Position{
			Platform:            "polymarket",
			MarketID:            "test-market-" + string(rune('a'+i)),
			MarketTitle:         "Test Market",
			Asset:               "BTC",
			Strike:              100000 + float64(i)*1000,
			Direction:           "above",
			EntryPrice:          0.85,
			Quantity:            100,
			Side:                "YES",
			Status:              "open",
			SafetyMarginAtEntry: safetyMargin,
			VolatilityAtEntry:   volatility,
		}

		id, err := posRepo.Create(pos)
		if err != nil {
			t.Fatalf("failed to create position: %v", err)
		}

		err = posRepo.Close(id, 0.92, "market_resolved", 7.0)
		if err != nil {
			t.Fatalf("failed to close position: %v", err)
		}
	}

	collector := NewCollector(db)
	outcomes, err := collector.CollectOutcomes(20)
	if err != nil {
		t.Fatalf("CollectOutcomes failed: %v", err)
	}

	// Verify parameters were captured
	for _, outcome := range outcomes {
		if outcome.SafetyMargin == 0 {
			t.Error("expected safety margin to be captured")
		}
		if outcome.Volatility == 0 {
			t.Error("expected volatility to be captured")
		}
		if outcome.Strike == 0 {
			t.Error("expected strike to be captured")
		}
	}
}

func TestCollector_CollectOutcomes_OrderedByExitTime(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	posRepo := persistence.NewPositionRepository(db)

	// Create 20 closed positions with different exit times
	for i := 0; i < 20; i++ {
		pos := &persistence.Position{
			Platform:            "polymarket",
			MarketID:            "test-market-" + string(rune('a'+i)),
			MarketTitle:         "Test Market",
			Asset:               "BTC",
			Strike:              100000,
			Direction:           "above",
			EntryPrice:          0.85,
			Quantity:            100,
			Side:                "YES",
			Status:              "open",
			SafetyMarginAtEntry: 1.5,
			VolatilityAtEntry:   0.5,
		}

		id, err := posRepo.Create(pos)
		if err != nil {
			t.Fatalf("failed to create position: %v", err)
		}

		err = posRepo.Close(id, 0.92, "market_resolved", 7.0)
		if err != nil {
			t.Fatalf("failed to close position: %v", err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(5 * time.Millisecond)
	}

	collector := NewCollector(db)
	outcomes, err := collector.CollectOutcomes(10)
	if err != nil {
		t.Fatalf("CollectOutcomes failed: %v", err)
	}

	// Should be ordered by exit time descending (most recent first)
	for i := 1; i < len(outcomes); i++ {
		if outcomes[i].ExitTime.After(outcomes[i-1].ExitTime) {
			t.Errorf("expected outcomes to be ordered by exit time desc, but outcome %d (%v) is after outcome %d (%v)",
				i, outcomes[i].ExitTime, i-1, outcomes[i-1].ExitTime)
		}
	}
}
