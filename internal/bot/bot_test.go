package bot

import (
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
