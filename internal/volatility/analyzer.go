package volatility

import (
	"math"
	"time"
)

// Direction represents the bet direction in a prediction market
type Direction string

const (
	// DirectionAbove means betting the price will be above strike
	DirectionAbove Direction = "above"
	// DirectionBelow means betting the price will be below strike
	DirectionBelow Direction = "below"
)

// Recommendation represents the trading recommendation based on safety margin
type Recommendation string

const (
	// RecommendationValid indicates the trade has sufficient safety margin
	RecommendationValid Recommendation = "valid"
	// RecommendationRisky indicates the trade is borderline and should be monitored
	RecommendationRisky Recommendation = "risky"
	// RecommendationReject indicates the trade should be avoided
	RecommendationReject Recommendation = "reject"
)

// Safety margin thresholds for recommendations
const (
	// SafetyMarginValidThreshold is the minimum margin for a valid trade
	SafetyMarginValidThreshold = 1.5
	// SafetyMarginRiskyThreshold is the minimum margin before rejection
	SafetyMarginRiskyThreshold = 0.8
	// MaxSafetyMargin caps the safety margin to avoid infinity
	MaxSafetyMargin = 1000.0
)

// AnalysisInput contains the parameters needed to analyze a market
type AnalysisInput struct {
	// CurrentPrice is the current price of the underlying asset
	CurrentPrice float64
	// StrikePrice is the strike price in the market condition
	StrikePrice float64
	// Direction is whether betting above or below strike
	Direction Direction
	// Volatility is the annualized volatility of the asset
	Volatility float64
	// TimeToCloseHours is the time until market closes in hours
	TimeToCloseHours float64
	// IsCrypto indicates if this is a crypto asset (affects annualization)
	IsCrypto bool
}

// AnalysisResult contains the output of volatility analysis
type AnalysisResult struct {
	// DistanceToStrike is the relative distance from current to strike
	// Positive means favorable direction
	DistanceToStrike float64
	// ExpectedMove is the expected price movement based on volatility
	ExpectedMove float64
	// SafetyMargin is the ratio of distance to expected move
	SafetyMargin float64
	// Recommendation is the trade recommendation
	Recommendation Recommendation
	// Timestamp when the analysis was performed
	Timestamp time.Time
}

// Analyze performs volatility analysis to determine if a trade is safe.
//
// The safety margin formula is:
//
//	safety_margin = distance_to_strike / (2 * expected_move)
//
// where:
//
//	distance_to_strike = |current_price - strike| / current_price
//	expected_move = volatility * sqrt(time_to_close_in_years)
//
// A higher safety margin indicates a safer trade.
func Analyze(input AnalysisInput) AnalysisResult {
	result := AnalysisResult{
		Timestamp: time.Now(),
	}

	// Validate input
	if input.CurrentPrice <= 0 {
		result.Recommendation = RecommendationReject
		return result
	}

	// Calculate distance to strike (relative to current price)
	// The sign depends on direction:
	// - For "above": positive when current > strike (we want price above)
	// - For "below": positive when current < strike (we want price below)
	var rawDistance float64
	if input.Direction == DirectionAbove {
		rawDistance = (input.CurrentPrice - input.StrikePrice) / input.CurrentPrice
	} else {
		rawDistance = (input.StrikePrice - input.CurrentPrice) / input.CurrentPrice
	}

	// Distance to strike is the absolute value for safety margin calculation
	// but we track if it's favorable or not
	result.DistanceToStrike = math.Abs(rawDistance)

	// If we're on the wrong side of the strike, it's unfavorable
	if rawDistance < 0 {
		// Currently on wrong side of strike
		result.DistanceToStrike = -result.DistanceToStrike
	}

	// Calculate expected move
	// expected_move = volatility * sqrt(time_in_years)
	var tradingDays float64
	if input.IsCrypto {
		tradingDays = TradingDaysCrypto
	} else {
		tradingDays = TradingDaysStock
	}

	timeInYears := input.TimeToCloseHours / 24.0 / tradingDays
	result.ExpectedMove = input.Volatility * math.Sqrt(timeInYears)

	// Calculate safety margin
	// safety_margin = distance_to_strike / (2 * expected_move)
	if result.ExpectedMove <= 0 {
		// Zero volatility or zero time means no expected move
		// This is extremely safe if we're on the right side
		if result.DistanceToStrike > 0 {
			result.SafetyMargin = MaxSafetyMargin
		} else {
			result.SafetyMargin = 0
		}
	} else {
		result.SafetyMargin = result.DistanceToStrike / (2 * result.ExpectedMove)
	}

	// Cap safety margin to avoid extreme values
	if result.SafetyMargin > MaxSafetyMargin {
		result.SafetyMargin = MaxSafetyMargin
	}

	// Determine recommendation based on safety margin
	result.Recommendation = determineRecommendation(result.SafetyMargin)

	return result
}

// determineRecommendation returns the trade recommendation based on safety margin
func determineRecommendation(safetyMargin float64) Recommendation {
	switch {
	case safetyMargin >= SafetyMarginValidThreshold:
		return RecommendationValid
	case safetyMargin >= SafetyMarginRiskyThreshold:
		return RecommendationRisky
	default:
		return RecommendationReject
	}
}
