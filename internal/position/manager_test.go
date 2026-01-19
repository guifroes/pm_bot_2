package position

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"prediction-bot/internal/persistence"
	"prediction-bot/internal/scanner"
	"prediction-bot/internal/sizing"
	"prediction-bot/internal/volatility"
	"prediction-bot/pkg/types"
)

// setupTestDB creates a temporary test database with migrations.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create temp file for test database
	tmpFile, err := os.CreateTemp("", "test_position_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	// Open database
	db, err := persistence.OpenDB(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

// MockVolatilityService mocks the volatility service for testing.
type MockVolatilityService struct {
	result volatility.ServiceResult
	err    error
}

func (m *MockVolatilityService) AnalyzeAsset(asset string, strikePrice float64, direction volatility.Direction, timeToClose time.Duration) (volatility.ServiceResult, error) {
	if m.err != nil {
		return volatility.ServiceResult{}, m.err
	}
	return m.result, nil
}

// TestProcessEntryDryRun tests the full entry flow in dry-run mode.
func TestProcessEntryDryRun(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Initialize bankroll
	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	// Create mock volatility service that returns a "valid" trade
	mockVolatility := &MockVolatilityService{
		result: volatility.ServiceResult{
			Asset:            "BTC",
			CurrentPrice:     100000.0,
			StrikePrice:      95000.0,
			Direction:        volatility.DirectionAbove,
			TimeToClose:      24 * time.Hour,
			IsCrypto:         true,
			Volatility:       0.5,
			DistanceToStrike: 0.05,
			ExpectedMove:     0.026,
			SafetyMargin:     1.91,
			Recommendation:   volatility.RecommendationValid,
			Timestamp:        time.Now(),
		},
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create manager
	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	// Create eligible market
	market := scanner.EligibleMarket{
		Market: types.Market{
			ID:              "test-market-1",
			Platform:        "polymarket",
			Title:           "Will Bitcoin be above $95,000 on Jan 20?",
			EndDate:         time.Now().Add(24 * time.Hour),
			OutcomeYesPrice: 0.90,
			Liquidity:       1000.0,
		},
		Parsed: &scanner.ParsedMarket{
			Asset:     "BTC",
			Strike:    95000.0,
			Direction: "above",
		},
		Probability: 0.90,
		BetSide:     "YES",
	}

	// Process entry in dry-run mode
	result, err := manager.ProcessEntry(market, true)
	if err != nil {
		t.Fatalf("ProcessEntry failed: %v", err)
	}

	// Verify result
	if result.Skipped {
		t.Fatalf("Expected trade to be processed, got skipped: %s", result.SkipReason)
	}

	if result.PositionID == 0 {
		t.Fatal("Expected position ID to be set")
	}

	// Verify position was created in database
	pos, err := positionRepo.GetByID(result.PositionID)
	if err != nil {
		t.Fatalf("Failed to get position: %v", err)
	}
	if pos == nil {
		t.Fatal("Expected position to exist in database")
	}

	// Verify position details
	if pos.Platform != "polymarket" {
		t.Errorf("Expected platform 'polymarket', got '%s'", pos.Platform)
	}
	if pos.MarketID != "test-market-1" {
		t.Errorf("Expected market ID 'test-market-1', got '%s'", pos.MarketID)
	}
	if pos.Asset != "BTC" {
		t.Errorf("Expected asset 'BTC', got '%s'", pos.Asset)
	}
	if pos.Strike != 95000.0 {
		t.Errorf("Expected strike 95000.0, got %f", pos.Strike)
	}
	if pos.Direction != "above" {
		t.Errorf("Expected direction 'above', got '%s'", pos.Direction)
	}
	if pos.Status != "open" {
		t.Errorf("Expected status 'open', got '%s'", pos.Status)
	}
	if pos.Side != "YES" {
		t.Errorf("Expected side 'YES', got '%s'", pos.Side)
	}
	if pos.Quantity <= 0 {
		t.Errorf("Expected positive quantity, got %f", pos.Quantity)
	}
	if pos.SafetyMarginAtEntry <= 0 {
		t.Errorf("Expected positive safety margin, got %f", pos.SafetyMarginAtEntry)
	}
}

// TestProcessEntryDuplicatePosition tests that duplicate positions are skipped.
func TestProcessEntryDuplicatePosition(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Initialize bankroll
	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	// Create existing position for this market
	_, err = positionRepo.Create(&persistence.Position{
		Platform:  "polymarket",
		MarketID:  "test-market-1",
		Asset:     "BTC",
		Strike:    95000.0,
		Direction: "above",
		EntryPrice: 0.90,
		Quantity:  5.0,
		Side:      "YES",
		Status:    "open",
	})
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	mockVolatility := &MockVolatilityService{
		result: volatility.ServiceResult{
			SafetyMargin:   1.91,
			Recommendation: volatility.RecommendationValid,
		},
	}

	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	market := scanner.EligibleMarket{
		Market: types.Market{
			ID:              "test-market-1",
			Platform:        "polymarket",
			OutcomeYesPrice: 0.90,
		},
		Parsed: &scanner.ParsedMarket{
			Asset:     "BTC",
			Strike:    95000.0,
			Direction: "above",
		},
		Probability: 0.90,
		BetSide:     "YES",
	}

	result, err := manager.ProcessEntry(market, true)
	if err != nil {
		t.Fatalf("ProcessEntry failed: %v", err)
	}

	if !result.Skipped {
		t.Fatal("Expected duplicate position to be skipped")
	}
	if result.SkipReason != SkipReasonDuplicate {
		t.Errorf("Expected skip reason '%s', got '%s'", SkipReasonDuplicate, result.SkipReason)
	}
}

// TestProcessEntryVolatilityReject tests that positions with poor volatility are rejected.
func TestProcessEntryVolatilityReject(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	// Mock volatility service that returns "reject" recommendation
	mockVolatility := &MockVolatilityService{
		result: volatility.ServiceResult{
			SafetyMargin:   0.5, // Below risky threshold
			Recommendation: volatility.RecommendationReject,
		},
	}

	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	market := scanner.EligibleMarket{
		Market: types.Market{
			ID:              "test-market-2",
			Platform:        "polymarket",
			EndDate:         time.Now().Add(24 * time.Hour),
			OutcomeYesPrice: 0.90,
		},
		Parsed: &scanner.ParsedMarket{
			Asset:     "BTC",
			Strike:    100000.0,
			Direction: "above",
		},
		Probability: 0.90,
		BetSide:     "YES",
	}

	result, err := manager.ProcessEntry(market, true)
	if err != nil {
		t.Fatalf("ProcessEntry failed: %v", err)
	}

	if !result.Skipped {
		t.Fatal("Expected volatility reject to skip position")
	}
	if result.SkipReason != SkipReasonVolatilityReject {
		t.Errorf("Expected skip reason '%s', got '%s'", SkipReasonVolatilityReject, result.SkipReason)
	}
}

// TestProcessEntrySizingTooSmall tests that positions too small are skipped.
func TestProcessEntrySizingTooSmall(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Very small bankroll to trigger minimum position constraint
	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 5.0) // Only $5
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	mockVolatility := &MockVolatilityService{
		result: volatility.ServiceResult{
			SafetyMargin:   1.91,
			Recommendation: volatility.RecommendationValid,
		},
	}

	// Set high minimum position to trigger "below_minimum"
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    10.0, // $10 minimum
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	market := scanner.EligibleMarket{
		Market: types.Market{
			ID:              "test-market-3",
			Platform:        "polymarket",
			EndDate:         time.Now().Add(24 * time.Hour),
			OutcomeYesPrice: 0.90,
		},
		Parsed: &scanner.ParsedMarket{
			Asset:     "BTC",
			Strike:    95000.0,
			Direction: "above",
		},
		Probability: 0.90,
		BetSide:     "YES",
	}

	result, err := manager.ProcessEntry(market, true)
	if err != nil {
		t.Fatalf("ProcessEntry failed: %v", err)
	}

	if !result.Skipped {
		t.Fatal("Expected small position to be skipped")
	}
	if result.SkipReason != SkipReasonSizingTooSmall {
		t.Errorf("Expected skip reason '%s', got '%s'", SkipReasonSizingTooSmall, result.SkipReason)
	}
}

// TestProcessEntryBankrollDeducted tests that bankroll is deducted after position entry.
func TestProcessEntryBankrollDeducted(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	mockVolatility := &MockVolatilityService{
		result: volatility.ServiceResult{
			SafetyMargin:   1.91,
			Volatility:     0.5,
			Recommendation: volatility.RecommendationValid,
		},
	}

	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	market := scanner.EligibleMarket{
		Market: types.Market{
			ID:              "test-market-4",
			Platform:        "polymarket",
			EndDate:         time.Now().Add(24 * time.Hour),
			OutcomeYesPrice: 0.90,
		},
		Parsed: &scanner.ParsedMarket{
			Asset:     "BTC",
			Strike:    95000.0,
			Direction: "above",
		},
		Probability: 0.90,
		BetSide:     "YES",
	}

	result, err := manager.ProcessEntry(market, true)
	if err != nil {
		t.Fatalf("ProcessEntry failed: %v", err)
	}

	if result.Skipped {
		t.Fatalf("Expected trade to be processed, got skipped: %s", result.SkipReason)
	}

	// Verify bankroll was deducted
	bankroll, err := bankrollRepo.Get("polymarket")
	if err != nil {
		t.Fatalf("Failed to get bankroll: %v", err)
	}

	pos, _ := positionRepo.GetByID(result.PositionID)
	expectedBankroll := 50.0 - (pos.Quantity * pos.EntryPrice)

	if bankroll.CurrentAmount > expectedBankroll+0.01 {
		t.Errorf("Expected bankroll ~%.2f, got %.2f", expectedBankroll, bankroll.CurrentAmount)
	}
}

// TestProcessEntryAllowsRisky tests that risky positions can be allowed with config.
func TestProcessEntryAllowsRisky(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 100.0) // Larger bankroll
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	// Mock volatility service that returns "risky" recommendation
	// Use safety margin of 1.3 which is between risky (0.8) and valid (1.5) thresholds
	mockVolatility := &MockVolatilityService{
		result: volatility.ServiceResult{
			SafetyMargin:   1.3, // Between risky and valid thresholds
			Volatility:     0.5,
			Recommendation: volatility.RecommendationRisky,
		},
	}

	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)
	manager.SetAllowRisky(true) // Allow risky positions

	// Use a lower market probability (0.82) to ensure edge with risky safety margin
	// At safety margin 1.3, EstimateWinProbability gives a boost
	// With entry 0.82 and boost, we should get enough edge for a position
	market := scanner.EligibleMarket{
		Market: types.Market{
			ID:              "test-market-5",
			Platform:        "polymarket",
			EndDate:         time.Now().Add(24 * time.Hour),
			OutcomeYesPrice: 0.82,
		},
		Parsed: &scanner.ParsedMarket{
			Asset:     "BTC",
			Strike:    95000.0,
			Direction: "above",
		},
		Probability: 0.82,
		BetSide:     "YES",
	}

	result, err := manager.ProcessEntry(market, true)
	if err != nil {
		t.Fatalf("ProcessEntry failed: %v", err)
	}

	if result.Skipped {
		t.Fatalf("Expected risky trade to be allowed, got skipped: %s", result.SkipReason)
	}
}

// TestExecuteExitDryRunStopLoss tests exiting a position due to stop loss in dry-run mode.
func TestExecuteExitDryRunStopLoss(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Initialize bankroll
	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	// Deduct position cost from bankroll (simulating entry)
	err = bankrollRepo.AddToBalance("polymarket", -9.0) // $9 position (10 contracts * $0.90)
	if err != nil {
		t.Fatalf("Failed to deduct from bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	// Create an open position
	positionID, err := positionRepo.Create(&persistence.Position{
		Platform:            "polymarket",
		MarketID:            "test-market-exit-1",
		MarketTitle:         "Will Bitcoin be above $95,000 on Jan 20?",
		Asset:               "BTC",
		Strike:              95000.0,
		Direction:           "above",
		EntryPrice:          0.90,
		Quantity:            10.0,
		Side:                "YES",
		Status:              "open",
		SafetyMarginAtEntry: 1.91,
		VolatilityAtEntry:   0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	mockVolatility := &MockVolatilityService{}
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	// Execute exit at a loss (stop loss triggered at $0.75)
	exitPrice := 0.75
	result, err := manager.ExecuteExit(positionID, exitPrice, ExitReasonStopLoss, true)
	if err != nil {
		t.Fatalf("ExecuteExit failed: %v", err)
	}

	// Verify exit result
	if result.PositionID != positionID {
		t.Errorf("Expected position ID %d, got %d", positionID, result.PositionID)
	}
	if result.ExitPrice != exitPrice {
		t.Errorf("Expected exit price %f, got %f", exitPrice, result.ExitPrice)
	}
	if result.ExitReason != ExitReasonStopLoss {
		t.Errorf("Expected exit reason '%s', got '%s'", ExitReasonStopLoss, result.ExitReason)
	}

	// Calculate expected PnL: (exitPrice - entryPrice) * quantity = (0.75 - 0.90) * 10 = -1.50
	expectedPnL := (exitPrice - 0.90) * 10.0
	if result.RealizedPnL < expectedPnL-0.01 || result.RealizedPnL > expectedPnL+0.01 {
		t.Errorf("Expected PnL ~%.2f, got %.2f", expectedPnL, result.RealizedPnL)
	}

	// Verify position is closed in database
	pos, err := positionRepo.GetByID(positionID)
	if err != nil {
		t.Fatalf("Failed to get position: %v", err)
	}
	if pos.Status != "closed" {
		t.Errorf("Expected status 'closed', got '%s'", pos.Status)
	}
	if pos.ExitPrice == nil || *pos.ExitPrice != exitPrice {
		t.Errorf("Expected exit price %f, got %v", exitPrice, pos.ExitPrice)
	}
	if pos.ExitReason == nil || *pos.ExitReason != ExitReasonStopLoss {
		t.Errorf("Expected exit reason '%s', got %v", ExitReasonStopLoss, pos.ExitReason)
	}

	// Verify bankroll was updated: original 41 + exit value (0.75 * 10) = 41 + 7.5 = 48.5
	// PnL is negative so total bankroll decreases
	bankroll, err := bankrollRepo.Get("polymarket")
	if err != nil {
		t.Fatalf("Failed to get bankroll: %v", err)
	}
	// 50 - 9 (entry) + 7.5 (exit proceeds) = 48.5
	expectedBankroll := 50.0 - 9.0 + (exitPrice * 10.0)
	if bankroll.CurrentAmount < expectedBankroll-0.01 || bankroll.CurrentAmount > expectedBankroll+0.01 {
		t.Errorf("Expected bankroll ~%.2f, got %.2f", expectedBankroll, bankroll.CurrentAmount)
	}
}

// TestExecuteExitDryRunVolatility tests exiting a position due to volatility in dry-run mode.
func TestExecuteExitDryRunVolatility(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	// Deduct position cost from bankroll
	err = bankrollRepo.AddToBalance("polymarket", -9.0)
	if err != nil {
		t.Fatalf("Failed to deduct from bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	positionID, err := positionRepo.Create(&persistence.Position{
		Platform:            "polymarket",
		MarketID:            "test-market-exit-2",
		MarketTitle:         "Will Bitcoin be above $95,000 on Jan 20?",
		Asset:               "BTC",
		Strike:              95000.0,
		Direction:           "above",
		EntryPrice:          0.90,
		Quantity:            10.0,
		Side:                "YES",
		Status:              "open",
		SafetyMarginAtEntry: 1.91,
		VolatilityAtEntry:   0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	mockVolatility := &MockVolatilityService{}
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	// Exit at current price (slight loss due to volatility concerns)
	exitPrice := 0.88
	result, err := manager.ExecuteExit(positionID, exitPrice, ExitReasonVolatility, true)
	if err != nil {
		t.Fatalf("ExecuteExit failed: %v", err)
	}

	if result.ExitReason != ExitReasonVolatility {
		t.Errorf("Expected exit reason '%s', got '%s'", ExitReasonVolatility, result.ExitReason)
	}

	// Verify position is closed
	pos, err := positionRepo.GetByID(positionID)
	if err != nil {
		t.Fatalf("Failed to get position: %v", err)
	}
	if pos.Status != "closed" {
		t.Errorf("Expected status 'closed', got '%s'", pos.Status)
	}
}

// TestExecuteExitDryRunWin tests exiting a winning position (market resolved).
func TestExecuteExitDryRunWin(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	// Deduct position cost from bankroll
	err = bankrollRepo.AddToBalance("polymarket", -9.0) // Entry: 10 * 0.90 = $9
	if err != nil {
		t.Fatalf("Failed to deduct from bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	positionID, err := positionRepo.Create(&persistence.Position{
		Platform:            "polymarket",
		MarketID:            "test-market-exit-3",
		MarketTitle:         "Will Bitcoin be above $95,000 on Jan 20?",
		Asset:               "BTC",
		Strike:              95000.0,
		Direction:           "above",
		EntryPrice:          0.90,
		Quantity:            10.0,
		Side:                "YES",
		Status:              "open",
		SafetyMarginAtEntry: 1.91,
		VolatilityAtEntry:   0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	mockVolatility := &MockVolatilityService{}
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	// Market resolved YES, exit at $1.00
	exitPrice := 1.0
	result, err := manager.ExecuteExit(positionID, exitPrice, ExitReasonResolved, true)
	if err != nil {
		t.Fatalf("ExecuteExit failed: %v", err)
	}

	// Calculate expected PnL: (1.00 - 0.90) * 10 = $1.00 profit
	expectedPnL := (exitPrice - 0.90) * 10.0
	if result.RealizedPnL < expectedPnL-0.01 || result.RealizedPnL > expectedPnL+0.01 {
		t.Errorf("Expected PnL ~%.2f, got %.2f", expectedPnL, result.RealizedPnL)
	}

	// Verify bankroll was updated: 41 + 10 (exit proceeds) = 51
	bankroll, err := bankrollRepo.Get("polymarket")
	if err != nil {
		t.Fatalf("Failed to get bankroll: %v", err)
	}
	expectedBankroll := 50.0 - 9.0 + (exitPrice * 10.0)
	if bankroll.CurrentAmount < expectedBankroll-0.01 || bankroll.CurrentAmount > expectedBankroll+0.01 {
		t.Errorf("Expected bankroll ~%.2f, got %.2f", expectedBankroll, bankroll.CurrentAmount)
	}
}

// TestExecuteExitPositionNotFound tests exiting a non-existent position.
func TestExecuteExitPositionNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bankrollRepo := persistence.NewBankrollRepository(db)
	positionRepo := persistence.NewPositionRepository(db)

	mockVolatility := &MockVolatilityService{}
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	// Try to exit a position that doesn't exist
	_, err := manager.ExecuteExit(99999, 0.50, ExitReasonStopLoss, true)
	if err == nil {
		t.Fatal("Expected error for non-existent position")
	}
}

// TestExecuteExitAlreadyClosed tests that closing an already closed position returns an error.
func TestExecuteExitAlreadyClosed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bankrollRepo := persistence.NewBankrollRepository(db)
	err := bankrollRepo.Initialize("polymarket", 50.0)
	if err != nil {
		t.Fatalf("Failed to initialize bankroll: %v", err)
	}

	positionRepo := persistence.NewPositionRepository(db)

	// Create a position and immediately close it
	positionID, err := positionRepo.Create(&persistence.Position{
		Platform:   "polymarket",
		MarketID:   "test-market-exit-4",
		EntryPrice: 0.90,
		Quantity:   10.0,
		Side:       "YES",
		Status:     "open",
	})
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Close it via repository directly
	err = positionRepo.Close(positionID, 0.95, "test_close", 0.5)
	if err != nil {
		t.Fatalf("Failed to close position: %v", err)
	}

	mockVolatility := &MockVolatilityService{}
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	manager := NewManager(positionRepo, bankrollRepo, mockVolatility, sizer)

	// Try to exit the already closed position
	_, err = manager.ExecuteExit(positionID, 0.50, ExitReasonStopLoss, true)
	if err == nil {
		t.Fatal("Expected error for already closed position")
	}
}
