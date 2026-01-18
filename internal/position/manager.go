package position

import (
	"fmt"
	"time"

	"prediction-bot/internal/persistence"
	"prediction-bot/internal/scanner"
	"prediction-bot/internal/sizing"
	"prediction-bot/internal/volatility"
)

// Skip reasons for position entry.
const (
	SkipReasonDuplicate        = "duplicate_position"
	SkipReasonVolatilityReject = "volatility_reject"
	SkipReasonVolatilityRisky  = "volatility_risky"
	SkipReasonSizingNoEdge     = "sizing_no_edge"
	SkipReasonSizingTooSmall   = "sizing_below_minimum"
	SkipReasonInsufficientFunds = "insufficient_funds"
)

// VolatilityAnalyzer defines the interface for volatility analysis.
type VolatilityAnalyzer interface {
	AnalyzeAsset(asset string, strikePrice float64, direction volatility.Direction, timeToClose time.Duration) (volatility.ServiceResult, error)
}

// EntryResult contains the result of processing a position entry.
type EntryResult struct {
	// Skipped is true if the position was not opened.
	Skipped bool
	// SkipReason explains why the position was skipped.
	SkipReason string
	// PositionID is the database ID of the created position (0 if skipped).
	PositionID int64
	// PositionSize is the dollar amount of the position.
	PositionSize float64
	// Quantity is the number of contracts/shares.
	Quantity float64
	// EntryPrice is the price per contract.
	EntryPrice float64
	// SafetyMargin is the volatility safety margin at entry.
	SafetyMargin float64
	// Volatility is the calculated volatility at entry.
	Volatility float64
	// WinProbability is the estimated win probability.
	WinProbability float64
}

// Manager handles position entry and management logic.
type Manager struct {
	positionRepo *persistence.PositionRepository
	bankrollRepo *persistence.BankrollRepository
	volatility   VolatilityAnalyzer
	sizer        *sizing.Sizer
	allowRisky   bool
}

// NewManager creates a new position manager with the given dependencies.
func NewManager(
	positionRepo *persistence.PositionRepository,
	bankrollRepo *persistence.BankrollRepository,
	volatilityService VolatilityAnalyzer,
	sizer *sizing.Sizer,
) *Manager {
	return &Manager{
		positionRepo: positionRepo,
		bankrollRepo: bankrollRepo,
		volatility:   volatilityService,
		sizer:        sizer,
		allowRisky:   false,
	}
}

// SetAllowRisky configures whether to allow risky positions (safety margin between 0.8 and 1.5).
func (m *Manager) SetAllowRisky(allow bool) {
	m.allowRisky = allow
}

// ProcessEntry processes an eligible market for potential position entry.
// If dryRun is true, the position is recorded but no actual order is placed.
//
// Flow:
// 1. Check for duplicate position
// 2. Analyze volatility
// 3. Calculate position size
// 4. Persist position to database
// 5. Deduct from bankroll
func (m *Manager) ProcessEntry(market scanner.EligibleMarket, dryRun bool) (EntryResult, error) {
	result := EntryResult{}

	// Step 1: Check for duplicate position
	existing, err := m.positionRepo.GetByMarket(market.Market.Platform, market.Market.ID)
	if err != nil {
		return result, fmt.Errorf("check duplicate position: %w", err)
	}
	if existing != nil {
		result.Skipped = true
		result.SkipReason = SkipReasonDuplicate
		return result, nil
	}

	// Step 2: Get bankroll for this platform
	bankroll, err := m.bankrollRepo.Get(market.Market.Platform)
	if err != nil {
		return result, fmt.Errorf("get bankroll: %w", err)
	}
	if bankroll == nil || bankroll.CurrentAmount <= 0 {
		result.Skipped = true
		result.SkipReason = SkipReasonInsufficientFunds
		return result, nil
	}

	// Step 3: Analyze volatility
	direction := volatility.DirectionAbove
	if market.Parsed.Direction == "below" {
		direction = volatility.DirectionBelow
	}

	timeToClose := time.Until(market.Market.EndDate)
	if timeToClose < 0 {
		timeToClose = 0
	}

	volResult, err := m.volatility.AnalyzeAsset(
		market.Parsed.Asset,
		market.Parsed.Strike,
		direction,
		timeToClose,
	)
	if err != nil {
		return result, fmt.Errorf("analyze volatility: %w", err)
	}

	// Check volatility recommendation
	if volResult.Recommendation == volatility.RecommendationReject {
		result.Skipped = true
		result.SkipReason = SkipReasonVolatilityReject
		result.SafetyMargin = volResult.SafetyMargin
		result.Volatility = volResult.Volatility
		return result, nil
	}

	if volResult.Recommendation == volatility.RecommendationRisky && !m.allowRisky {
		result.Skipped = true
		result.SkipReason = SkipReasonVolatilityRisky
		result.SafetyMargin = volResult.SafetyMargin
		result.Volatility = volResult.Volatility
		return result, nil
	}

	// Step 4: Calculate position size
	entryPrice := market.Probability
	if market.BetSide == "NO" {
		entryPrice = 1.0 - market.Probability
	}

	// Estimate win probability based on safety margin
	winProb := sizing.EstimateWinProbability(entryPrice, volResult.SafetyMargin)

	sizingInput := sizing.SizingInput{
		EntryPrice:   entryPrice,
		WinProb:      winProb,
		Bankroll:     bankroll.CurrentAmount,
		SafetyMargin: volResult.SafetyMargin,
	}

	sizingOutput := m.sizer.Calculate(sizingInput)

	if sizingOutput.PositionSize <= 0 {
		result.Skipped = true
		if sizingOutput.Reason == "no_edge" {
			result.SkipReason = SkipReasonSizingNoEdge
		} else {
			result.SkipReason = SkipReasonSizingTooSmall
		}
		result.SafetyMargin = volResult.SafetyMargin
		result.Volatility = volResult.Volatility
		return result, nil
	}

	// Calculate quantity (number of contracts)
	quantity := sizingOutput.PositionSize / entryPrice

	// Step 5: Persist position to database
	position := &persistence.Position{
		Platform:            market.Market.Platform,
		MarketID:            market.Market.ID,
		MarketTitle:         market.Market.Title,
		Asset:               market.Parsed.Asset,
		Strike:              market.Parsed.Strike,
		Direction:           market.Parsed.Direction,
		EntryPrice:          entryPrice,
		Quantity:            quantity,
		Side:                market.BetSide,
		Status:              "open",
		SafetyMarginAtEntry: volResult.SafetyMargin,
		VolatilityAtEntry:   volResult.Volatility,
	}

	positionID, err := m.positionRepo.Create(position)
	if err != nil {
		return result, fmt.Errorf("create position: %w", err)
	}

	// Step 6: Deduct from bankroll
	err = m.bankrollRepo.AddToBalance(market.Market.Platform, -sizingOutput.PositionSize)
	if err != nil {
		return result, fmt.Errorf("deduct from bankroll: %w", err)
	}

	// Populate result
	result.PositionID = positionID
	result.PositionSize = sizingOutput.PositionSize
	result.Quantity = quantity
	result.EntryPrice = entryPrice
	result.SafetyMargin = volResult.SafetyMargin
	result.Volatility = volResult.Volatility
	result.WinProbability = winProb

	return result, nil
}
