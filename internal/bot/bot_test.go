package bot

import (
	"context"
	"testing"
	"time"

	"prediction-bot/internal/config"
	"prediction-bot/internal/persistence"
	"prediction-bot/internal/platform"
	"prediction-bot/internal/position"
	"prediction-bot/internal/scanner"
	"prediction-bot/internal/sizing"
	"prediction-bot/internal/volatility"
	"prediction-bot/pkg/types"
)

// MockPlatform implements platform.Platform for testing.
type MockPlatform struct {
	name     string
	markets  []types.Market
	balance  float64
	listErr  error
	scanTime time.Time
}

func (m *MockPlatform) Name() string {
	return m.name
}

func (m *MockPlatform) ListMarkets(filter types.MarketFilter) ([]types.Market, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.markets, nil
}

func (m *MockPlatform) GetOrderBook(tokenID string) (*types.OrderBook, error) {
	return &types.OrderBook{}, nil
}

func (m *MockPlatform) GetBalance() (float64, error) {
	return m.balance, nil
}

func (m *MockPlatform) GetPositions() ([]types.Position, error) {
	return []types.Position{}, nil
}

// MockVolatilityAnalyzer implements position.VolatilityAnalyzer for testing.
type MockVolatilityAnalyzer struct {
	safetyMargin   float64
	vol            float64
	recommendation volatility.Recommendation
}

func (m *MockVolatilityAnalyzer) AnalyzeAsset(
	asset string,
	strikePrice float64,
	direction volatility.Direction,
	timeToClose time.Duration,
) (volatility.ServiceResult, error) {
	return volatility.ServiceResult{
		SafetyMargin:   m.safetyMargin,
		Volatility:     m.vol,
		Recommendation: m.recommendation,
	}, nil
}

// TestRunScanCycle_ExecutesWithoutError tests that a scan cycle runs successfully
// with mock platforms and processes eligible markets.
func TestRunScanCycle_ExecutesWithoutError(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll for test platform
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create mock platform with eligible market
	endDate := time.Now().Add(24 * time.Hour)
	mockPlatform := &MockPlatform{
		name:    "mock",
		balance: 100.0,
		markets: []types.Market{
			{
				ID:              "test-market-1",
				Platform:        "mock",
				Title:           "Will Bitcoin be above $100,000 on Jan 20?",
				OutcomeYesPrice: 0.85,
				OutcomeNoPrice:  0.15,
				Volume:          10000.0,
				Liquidity:       5000.0,
				Active:          true,
				EndDate:         endDate,
			},
		},
	}

	// Create mock volatility analyzer with valid analysis
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0, // Valid margin
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create scanner with params
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Run single scan cycle
	err = bot.RunScanCycle()
	if err != nil {
		t.Fatalf("RunScanCycle failed: %v", err)
	}

	// Verify that the eligible market was processed
	// Check if position was created
	positions, err := posRepo.GetOpen()
	if err != nil {
		t.Fatalf("failed to get open positions: %v", err)
	}

	// Should have created one position for the eligible market
	if len(positions) != 1 {
		t.Errorf("expected 1 position, got %d", len(positions))
	}

	if len(positions) > 0 {
		pos := positions[0]
		if pos.MarketID != "test-market-1" {
			t.Errorf("expected market ID 'test-market-1', got %s", pos.MarketID)
		}
		if pos.Platform != "mock" {
			t.Errorf("expected platform 'mock', got %s", pos.Platform)
		}
		if pos.Status != "open" {
			t.Errorf("expected status 'open', got %s", pos.Status)
		}
	}
}

// TestRunScanCycle_MultiplePlatforms tests scanning across multiple platforms.
func TestRunScanCycle_MultiplePlatforms(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll for both platforms
	err = bankRepo.Initialize("platform1", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}
	err = bankRepo.Initialize("platform2", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	endDate := time.Now().Add(24 * time.Hour)

	// Create two mock platforms
	platform1 := &MockPlatform{
		name:    "platform1",
		balance: 100.0,
		markets: []types.Market{
			{
				ID:              "market-p1",
				Platform:        "platform1",
				Title:           "Will Ethereum be above $5,000 on Jan 20?",
				OutcomeYesPrice: 0.82,
				OutcomeNoPrice:  0.18,
				Volume:          5000.0,
				Liquidity:       2000.0,
				Active:          true,
				EndDate:         endDate,
			},
		},
	}

	platform2 := &MockPlatform{
		name:    "platform2",
		balance: 100.0,
		markets: []types.Market{
			{
				ID:              "market-p2",
				Platform:        "platform2",
				Title:           "Will Bitcoin be below $80,000 on Jan 20?",
				OutcomeYesPrice: 0.88,
				OutcomeNoPrice:  0.12,
				Volume:          8000.0,
				Liquidity:       3000.0,
				Active:          true,
				EndDate:         endDate,
			},
		},
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot with multiple platforms
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{platform1, platform2}, sc, manager)

	// Run scan cycle
	err = bot.RunScanCycle()
	if err != nil {
		t.Fatalf("RunScanCycle failed: %v", err)
	}

	// Verify positions were created for both platforms
	positions, err := posRepo.GetOpen()
	if err != nil {
		t.Fatalf("failed to get open positions: %v", err)
	}

	if len(positions) != 2 {
		t.Errorf("expected 2 positions, got %d", len(positions))
	}

	// Verify we have one from each platform
	platformsSeen := make(map[string]bool)
	for _, pos := range positions {
		platformsSeen[pos.Platform] = true
	}

	if !platformsSeen["platform1"] {
		t.Error("expected position from platform1")
	}
	if !platformsSeen["platform2"] {
		t.Error("expected position from platform2")
	}
}

// TestRunScanCycle_NoEligibleMarkets tests that scan cycle handles empty results gracefully.
func TestRunScanCycle_NoEligibleMarkets(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create mock platform with no markets
	mockPlatform := &MockPlatform{
		name:    "mock",
		balance: 100.0,
		markets: []types.Market{}, // Empty
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Run scan cycle - should succeed without error
	err = bot.RunScanCycle()
	if err != nil {
		t.Fatalf("RunScanCycle failed: %v", err)
	}

	// Verify no positions were created
	positions, err := posRepo.GetOpen()
	if err != nil {
		t.Fatalf("failed to get open positions: %v", err)
	}

	if len(positions) != 0 {
		t.Errorf("expected 0 positions, got %d", len(positions))
	}
}

// TestRunMonitorCycle_ChecksAllOpenPositions tests that the monitor cycle
// checks all open positions for stop loss and volatility exits.
func TestRunMonitorCycle_ChecksAllOpenPositions(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create an open position
	pos := &persistence.Position{
		Platform:            "mock",
		MarketID:            "test-market-1",
		MarketTitle:         "Will Bitcoin be above $100,000?",
		Asset:               "BTC",
		Strike:              100000,
		Direction:           "above",
		EntryPrice:          0.85,
		Quantity:            10.0,
		Side:                "YES",
		Status:              "open",
		SafetyMarginAtEntry: 2.0,
		VolatilityAtEntry:   0.5,
	}
	_, err = posRepo.Create(pos)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Create mock platform that returns current price for the position
	mockPlatform := &MockPlatformWithPrice{
		name:         "mock",
		balance:      100.0,
		markets:      []types.Market{},
		currentPrice: 0.80, // Below entry price but above stop loss
	}

	// Create mock volatility analyzer that returns safe margin
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0, // Safe - no volatility exit
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create monitor
	monitor := position.NewMonitor(0.15) // 15% stop loss

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Set monitor on bot
	bot.SetMonitor(monitor)
	bot.SetVolatilityAnalyzer(mockVolatility)
	bot.SetPositionRepo(posRepo)

	// Run monitor cycle - should complete without error
	err = bot.RunMonitorCycle()
	if err != nil {
		t.Fatalf("RunMonitorCycle failed: %v", err)
	}

	// Position should still be open (no stop loss or volatility exit triggered)
	positions, err := posRepo.GetOpen()
	if err != nil {
		t.Fatalf("failed to get open positions: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("expected 1 open position, got %d", len(positions))
	}
}

// TestRunMonitorCycle_TriggersStopLoss tests that stop loss exits are triggered.
func TestRunMonitorCycle_TriggersStopLoss(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create an open position
	pos := &persistence.Position{
		Platform:            "mock",
		MarketID:            "test-market-stop-loss",
		MarketTitle:         "Will Bitcoin be above $100,000?",
		Asset:               "BTC",
		Strike:              100000,
		Direction:           "above",
		EntryPrice:          0.90,
		Quantity:            10.0,
		Side:                "YES",
		Status:              "open",
		SafetyMarginAtEntry: 2.0,
		VolatilityAtEntry:   0.5,
	}
	posID, err := posRepo.Create(pos)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Create mock platform with price below stop loss threshold
	// Stop loss at 15%: 0.90 * 0.85 = 0.765
	// Current price 0.70 is below threshold
	mockPlatform := &MockPlatformWithPrice{
		name:         "mock",
		balance:      100.0,
		markets:      []types.Market{},
		currentPrice: 0.70, // Below stop loss threshold
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create monitor with 15% stop loss
	monitor := position.NewMonitor(0.15)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Set monitor and dependencies
	bot.SetMonitor(monitor)
	bot.SetVolatilityAnalyzer(mockVolatility)
	bot.SetPositionRepo(posRepo)

	// Run monitor cycle
	err = bot.RunMonitorCycle()
	if err != nil {
		t.Fatalf("RunMonitorCycle failed: %v", err)
	}

	// Position should be closed due to stop loss
	closedPos, err := posRepo.GetByID(posID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}

	if closedPos.Status != "closed" {
		t.Errorf("expected position to be closed, got status %s", closedPos.Status)
	}

	if closedPos.ExitReason == nil || *closedPos.ExitReason != "stop_loss" {
		exitReason := "nil"
		if closedPos.ExitReason != nil {
			exitReason = *closedPos.ExitReason
		}
		t.Errorf("expected exit reason 'stop_loss', got %s", exitReason)
	}
}

// TestRunMonitorCycle_TriggersVolatilityExit tests that volatility exits are triggered.
func TestRunMonitorCycle_TriggersVolatilityExit(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create an open position
	pos := &persistence.Position{
		Platform:            "mock",
		MarketID:            "test-market-vol-exit",
		MarketTitle:         "Will Bitcoin be above $100,000?",
		Asset:               "BTC",
		Strike:              100000,
		Direction:           "above",
		EntryPrice:          0.90,
		Quantity:            10.0,
		Side:                "YES",
		Status:              "open",
		SafetyMarginAtEntry: 2.0,
		VolatilityAtEntry:   0.5,
	}
	posID, err := posRepo.Create(pos)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Create mock platform with price above stop loss
	mockPlatform := &MockPlatformWithPrice{
		name:         "mock",
		balance:      100.0,
		markets:      []types.Market{},
		currentPrice: 0.85, // Above stop loss threshold
	}

	// Create mock volatility analyzer with LOW safety margin (below 0.8 threshold)
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   0.5, // Below 0.8 threshold - triggers volatility exit
		vol:            0.8,
		recommendation: volatility.RecommendationReject,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create monitor with 15% stop loss
	monitor := position.NewMonitor(0.15)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Set monitor and dependencies
	bot.SetMonitor(monitor)
	bot.SetVolatilityAnalyzer(mockVolatility)
	bot.SetPositionRepo(posRepo)

	// Run monitor cycle
	err = bot.RunMonitorCycle()
	if err != nil {
		t.Fatalf("RunMonitorCycle failed: %v", err)
	}

	// Position should be closed due to volatility exit
	closedPos, err := posRepo.GetByID(posID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}

	if closedPos.Status != "closed" {
		t.Errorf("expected position to be closed, got status %s", closedPos.Status)
	}

	if closedPos.ExitReason == nil || *closedPos.ExitReason != "volatility_exit" {
		exitReason := "nil"
		if closedPos.ExitReason != nil {
			exitReason = *closedPos.ExitReason
		}
		t.Errorf("expected exit reason 'volatility_exit', got %s", exitReason)
	}
}

// TestRunMonitorCycle_NoOpenPositions tests that monitor cycle handles empty positions gracefully.
func TestRunMonitorCycle_NoOpenPositions(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create mock platform
	mockPlatform := &MockPlatformWithPrice{
		name:         "mock",
		balance:      100.0,
		markets:      []types.Market{},
		currentPrice: 0.85,
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create monitor
	monitor := position.NewMonitor(0.15)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Set monitor and dependencies
	bot.SetMonitor(monitor)
	bot.SetVolatilityAnalyzer(mockVolatility)
	bot.SetPositionRepo(posRepo)

	// Run monitor cycle - should succeed without error
	err = bot.RunMonitorCycle()
	if err != nil {
		t.Fatalf("RunMonitorCycle failed: %v", err)
	}
}

// MockPlatformWithPrice extends MockPlatform with current price support.
type MockPlatformWithPrice struct {
	name         string
	markets      []types.Market
	balance      float64
	listErr      error
	currentPrice float64
}

func (m *MockPlatformWithPrice) Name() string {
	return m.name
}

func (m *MockPlatformWithPrice) ListMarkets(filter types.MarketFilter) ([]types.Market, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.markets, nil
}

func (m *MockPlatformWithPrice) GetOrderBook(tokenID string) (*types.OrderBook, error) {
	// Return order book with the current price
	return &types.OrderBook{
		Bids: []types.Level{{Price: m.currentPrice, Size: 100}},
		Asks: []types.Level{{Price: m.currentPrice + 0.01, Size: 100}},
	}, nil
}

func (m *MockPlatformWithPrice) GetBalance() (float64, error) {
	return m.balance, nil
}

func (m *MockPlatformWithPrice) GetPositions() ([]types.Position, error) {
	return []types.Position{}, nil
}

func (m *MockPlatformWithPrice) GetCurrentPrice(marketID string) (float64, error) {
	return m.currentPrice, nil
}

// TestRunScanCycle_SkipsIneligibleMarkets tests that ineligible markets are skipped.
func TestRunScanCycle_SkipsIneligibleMarkets(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	endDate := time.Now().Add(24 * time.Hour)

	// Create mock platform with mix of eligible and ineligible markets
	mockPlatform := &MockPlatform{
		name:    "mock",
		balance: 100.0,
		markets: []types.Market{
			{
				ID:              "eligible-market",
				Platform:        "mock",
				Title:           "Will Bitcoin be above $100,000 on Jan 20?",
				OutcomeYesPrice: 0.85, // 85% - eligible
				OutcomeNoPrice:  0.15,
				Volume:          10000.0,
				Liquidity:       5000.0,
				Active:          true,
				EndDate:         endDate,
			},
			{
				ID:              "low-prob-market",
				Platform:        "mock",
				Title:           "Will Ethereum be above $10,000 on Jan 20?",
				OutcomeYesPrice: 0.50, // 50% - below threshold
				OutcomeNoPrice:  0.50,
				Volume:          10000.0,
				Liquidity:       5000.0,
				Active:          true,
				EndDate:         endDate,
			},
			{
				ID:              "political-market",
				Platform:        "mock",
				Title:           "Will candidate X win the election?", // Not parseable
				OutcomeYesPrice: 0.90,
				OutcomeNoPrice:  0.10,
				Volume:          10000.0,
				Liquidity:       5000.0,
				Active:          true,
				EndDate:         endDate,
			},
		},
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second,
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Run scan cycle
	err = bot.RunScanCycle()
	if err != nil {
		t.Fatalf("RunScanCycle failed: %v", err)
	}

	// Verify only eligible market was processed
	positions, err := posRepo.GetOpen()
	if err != nil {
		t.Fatalf("failed to get open positions: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("expected 1 position, got %d", len(positions))
	}

	if len(positions) > 0 && positions[0].MarketID != "eligible-market" {
		t.Errorf("expected market ID 'eligible-market', got %s", positions[0].MarketID)
	}
}

// TestRun_ExecutesCyclesWithTicker tests that Run executes scan and monitor cycles
// based on configured intervals until context is cancelled.
func TestRun_ExecutesCyclesWithTicker(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create mock platform with no markets (to keep cycles fast)
	mockPlatform := &MockPlatformWithPrice{
		name:         "mock",
		balance:      100.0,
		markets:      []types.Market{},
		currentPrice: 0.85,
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create monitor
	monitor := position.NewMonitor(0.15)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot with very short intervals for testing
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    100 * time.Millisecond,
		MonitorInterval: 50 * time.Millisecond,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Set dependencies
	bot.SetMonitor(monitor)
	bot.SetVolatilityAnalyzer(mockVolatility)
	bot.SetPositionRepo(posRepo)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Run bot - should complete without error when context expires
	err = bot.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run failed with unexpected error: %v", err)
	}
}

// TestRun_GracefulShutdownOnContextCancel tests that Run shuts down gracefully
// when the context is cancelled.
func TestRun_GracefulShutdownOnContextCancel(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Create mock platform
	mockPlatform := &MockPlatformWithPrice{
		name:         "mock",
		balance:      100.0,
		markets:      []types.Market{},
		currentPrice: 0.85,
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create monitor
	monitor := position.NewMonitor(0.15)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot with longer intervals
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    1 * time.Second,
		MonitorInterval: 500 * time.Millisecond,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Set dependencies
	bot.SetMonitor(monitor)
	bot.SetVolatilityAnalyzer(mockVolatility)
	bot.SetPositionRepo(posRepo)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Track when Run completes
	done := make(chan error, 1)

	go func() {
		done <- bot.Run(ctx)
	}()

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for Run to complete with timeout
	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Errorf("Run returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not shut down within timeout after context cancellation")
	}
}

// TestRun_RunsImmediateScanOnStart tests that Run executes an immediate scan cycle
// when started, before waiting for the first ticker interval.
func TestRun_RunsImmediateScanOnStart(t *testing.T) {
	// Create temporary database
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = persistence.RunMigrations(db, "../../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Setup repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll
	err = bankRepo.Initialize("mock", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	endDate := time.Now().Add(24 * time.Hour)

	// Create mock platform with eligible market
	mockPlatform := &MockPlatformWithPrice{
		name:    "mock",
		balance: 100.0,
		markets: []types.Market{
			{
				ID:              "immediate-scan-market",
				Platform:        "mock",
				Title:           "Will Bitcoin be above $100,000 on Jan 20?",
				OutcomeYesPrice: 0.85,
				OutcomeNoPrice:  0.15,
				Volume:          10000.0,
				Liquidity:       5000.0,
				Active:          true,
				EndDate:         endDate,
			},
		},
		currentPrice: 0.85,
	}

	// Create mock volatility analyzer
	mockVolatility := &MockVolatilityAnalyzer{
		safetyMargin:   2.0,
		vol:            0.5,
		recommendation: volatility.RecommendationValid,
	}

	// Create sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  0.25,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Create position manager
	manager := position.NewManager(posRepo, bankRepo, mockVolatility, sizer)

	// Create monitor
	monitor := position.NewMonitor(0.15)

	// Create scanner
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}
	sc := scanner.NewScanner(params)

	// Create bot with long scan interval (we want to test immediate scan)
	bot := NewBot(BotConfig{
		DryRun:          true,
		ScanInterval:    10 * time.Second, // Long interval
		MonitorInterval: 5 * time.Second,
	}, []platform.Platform{mockPlatform}, sc, manager)

	// Set dependencies
	bot.SetMonitor(monitor)
	bot.SetVolatilityAnalyzer(mockVolatility)
	bot.SetPositionRepo(posRepo)

	// Create context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Run bot
	bot.Run(ctx)

	// Check that position was created (scan happened immediately)
	positions, err := posRepo.GetOpen()
	if err != nil {
		t.Fatalf("failed to get open positions: %v", err)
	}

	// Should have created position from immediate scan
	if len(positions) != 1 {
		t.Errorf("expected 1 position from immediate scan, got %d", len(positions))
	}

	if len(positions) > 0 && positions[0].MarketID != "immediate-scan-market" {
		t.Errorf("expected market ID 'immediate-scan-market', got %s", positions[0].MarketID)
	}
}
