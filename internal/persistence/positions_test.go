package persistence

import (
	"os"
	"testing"
)

func TestPositionRepository_Create(t *testing.T) {
	// Setup: create temp database
	tmpFile, err := os.CreateTemp("", "test_positions_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create repository
	repo := NewPositionRepository(db)

	// Test: Create a position
	pos := &Position{
		Platform:            "polymarket",
		MarketID:            "0x123abc",
		MarketTitle:         "Will Bitcoin be above $100k?",
		Asset:               "BTC",
		Strike:              100000,
		Direction:           "above",
		EntryPrice:          0.85,
		Quantity:            10.0,
		Side:                "YES",
		Status:              "open",
		SafetyMarginAtEntry: 1.8,
		VolatilityAtEntry:   0.5,
	}

	id, err := repo.Create(pos)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestPositionRepository_GetOpen(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_positions_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewPositionRepository(db)

	// Create two open positions and one closed
	openPos1 := &Position{
		Platform:   "polymarket",
		MarketID:   "0x111",
		EntryPrice: 0.80,
		Quantity:   5.0,
		Side:       "YES",
		Status:     "open",
	}
	openPos2 := &Position{
		Platform:   "kalshi",
		MarketID:   "KXBTC-123",
		EntryPrice: 0.90,
		Quantity:   3.0,
		Side:       "YES",
		Status:     "open",
	}
	closedPos := &Position{
		Platform:   "polymarket",
		MarketID:   "0x222",
		EntryPrice: 0.75,
		Quantity:   2.0,
		Side:       "NO",
		Status:     "closed",
	}

	repo.Create(openPos1)
	repo.Create(openPos2)
	repo.Create(closedPos)

	// Test: Get only open positions
	openPositions, err := repo.GetOpen()
	if err != nil {
		t.Fatalf("failed to get open positions: %v", err)
	}

	if len(openPositions) != 2 {
		t.Errorf("expected 2 open positions, got %d", len(openPositions))
	}
}

func TestPositionRepository_GetByMarket(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_positions_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewPositionRepository(db)

	// Create a position
	pos := &Position{
		Platform:   "polymarket",
		MarketID:   "0xABC123",
		EntryPrice: 0.85,
		Quantity:   10.0,
		Side:       "YES",
		Status:     "open",
	}
	repo.Create(pos)

	// Test: Get by platform and market ID
	found, err := repo.GetByMarket("polymarket", "0xABC123")
	if err != nil {
		t.Fatalf("failed to get by market: %v", err)
	}

	if found == nil {
		t.Fatal("expected position, got nil")
	}

	if found.MarketID != "0xABC123" {
		t.Errorf("expected market ID 0xABC123, got %s", found.MarketID)
	}
}

func TestPositionRepository_Update(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_positions_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewPositionRepository(db)

	// Create a position
	pos := &Position{
		Platform:   "polymarket",
		MarketID:   "0xUPDATE",
		EntryPrice: 0.80,
		Quantity:   5.0,
		Side:       "YES",
		Status:     "open",
	}
	id, _ := repo.Create(pos)

	// Test: Update quantity
	pos.ID = id
	pos.Quantity = 10.0
	if err := repo.Update(pos); err != nil {
		t.Fatalf("failed to update position: %v", err)
	}

	// Verify update
	updated, _ := repo.GetByID(id)
	if updated.Quantity != 10.0 {
		t.Errorf("expected quantity 10.0, got %f", updated.Quantity)
	}
}

func TestPositionRepository_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_positions_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewPositionRepository(db)

	// Create an open position
	pos := &Position{
		Platform:   "kalshi",
		MarketID:   "KXCLOSE",
		EntryPrice: 0.90,
		Quantity:   3.0,
		Side:       "YES",
		Status:     "open",
	}
	id, _ := repo.Create(pos)

	// Test: Close the position
	exitPrice := 1.0
	reason := "resolution_win"
	pnl := 0.30 // (1.0 - 0.90) * 3.0

	if err := repo.Close(id, exitPrice, reason, pnl); err != nil {
		t.Fatalf("failed to close position: %v", err)
	}

	// Verify closed
	closed, _ := repo.GetByID(id)
	if closed.Status != "closed" {
		t.Errorf("expected status closed, got %s", closed.Status)
	}
	if closed.ExitPrice == nil || *closed.ExitPrice != exitPrice {
		t.Errorf("expected exit price %f, got %v", exitPrice, closed.ExitPrice)
	}
	if closed.ExitReason == nil || *closed.ExitReason != reason {
		t.Errorf("expected exit reason %s, got %v", reason, closed.ExitReason)
	}
	if closed.RealizedPnL == nil || *closed.RealizedPnL != pnl {
		t.Errorf("expected realized pnl %f, got %v", pnl, closed.RealizedPnL)
	}
}

func TestPositionRepository_GetOpenByPlatform(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_positions_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewPositionRepository(db)

	// Create positions on different platforms
	repo.Create(&Position{Platform: "polymarket", MarketID: "0x1", EntryPrice: 0.80, Quantity: 1, Side: "YES", Status: "open"})
	repo.Create(&Position{Platform: "polymarket", MarketID: "0x2", EntryPrice: 0.85, Quantity: 1, Side: "YES", Status: "open"})
	repo.Create(&Position{Platform: "kalshi", MarketID: "KX1", EntryPrice: 0.90, Quantity: 1, Side: "YES", Status: "open"})

	// Test: Get open positions by platform
	polyPositions, err := repo.GetOpenByPlatform("polymarket")
	if err != nil {
		t.Fatalf("failed to get open by platform: %v", err)
	}

	if len(polyPositions) != 2 {
		t.Errorf("expected 2 polymarket positions, got %d", len(polyPositions))
	}
}

