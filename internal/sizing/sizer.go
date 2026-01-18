package sizing

import "math"

// SizerConfig holds configuration for the Sizer.
type SizerConfig struct {
	KellyFraction  float64 // Fraction of Kelly to use (e.g., 0.25 for quarter Kelly)
	MinPosition    float64 // Minimum position size in dollars
	MaxBankrollPct float64 // Maximum percentage of bankroll per position
}

// SizingInput contains the inputs needed to calculate position size.
type SizingInput struct {
	EntryPrice   float64 // Price to buy at (0 < price < 1)
	WinProb      float64 // Estimated win probability
	Bankroll     float64 // Total available capital
	SafetyMargin float64 // Volatility safety margin
}

// SizingOutput contains the calculated position size and metadata.
type SizingOutput struct {
	PositionSize float64 // Final position size in dollars (rounded down to cents)
	RawKelly     float64 // Raw Kelly position before constraints
	BankrollPct  float64 // Percentage of bankroll for this position
	Reason       string  // Reason if position is 0 (e.g., "no_edge", "below_minimum")
}

// Sizer calculates position sizes with constraints.
type Sizer struct {
	config SizerConfig
}

// NewSizer creates a new Sizer with the given configuration.
func NewSizer(config SizerConfig) *Sizer {
	return &Sizer{config: config}
}

// Calculate determines the position size applying Kelly criterion and constraints.
func (s *Sizer) Calculate(input SizingInput) SizingOutput {
	// Calculate raw Kelly position
	rawKelly := CalculateKelly(input.EntryPrice, input.WinProb, input.Bankroll, s.config.KellyFraction)

	// If Kelly returns 0, we have no edge
	if rawKelly <= 0 {
		return SizingOutput{
			PositionSize: 0,
			RawKelly:     0,
			BankrollPct:  0,
			Reason:       "no_edge",
		}
	}

	// Apply maximum constraint (max % of bankroll)
	maxPosition := input.Bankroll * s.config.MaxBankrollPct
	position := math.Min(rawKelly, maxPosition)

	// Check if position is below minimum
	if position < s.config.MinPosition {
		// If raw kelly was positive but position is below minimum after constraints,
		// we return 0 (not worth the transaction cost)
		return SizingOutput{
			PositionSize: 0,
			RawKelly:     rawKelly,
			BankrollPct:  position / input.Bankroll,
			Reason:       "below_minimum",
		}
	}

	// Round down to cents (2 decimal places)
	position = math.Floor(position*100) / 100

	// Calculate final bankroll percentage
	bankrollPct := position / input.Bankroll

	return SizingOutput{
		PositionSize: position,
		RawKelly:     rawKelly,
		BankrollPct:  bankrollPct,
		Reason:       "",
	}
}

// EstimateWinProbability estimates the true win probability based on market price and safety margin.
//
// The idea is that if volatility analysis shows a high safety margin, the true probability
// of the market resolving YES (for "above" bets) is higher than what the market implies.
//
// This function provides a probability boost based on safety margin:
// - Safety margin >= 2.0: significant boost (the market is very safe)
// - Safety margin ~1.5: moderate boost (the market is reasonably safe)
// - Safety margin < 1.0: no boost or slight reduction (the market is risky)
//
// The boost is calculated as: boost = (safetyMargin - 1.0) * boostFactor * (1 - marketPrice)
// This ensures:
// - Boost scales with distance from 1.0 (higher safety = more boost)
// - Boost is proportional to the "room" available (closer to 1.0 = less room to boost)
// - Result never exceeds 1.0
func EstimateWinProbability(marketPrice, safetyMargin float64) float64 {
	// Validate inputs
	if marketPrice <= 0 || marketPrice > 1 {
		return marketPrice
	}

	// Base probability is the market price
	probability := marketPrice

	// Calculate boost based on safety margin
	// The boost factor determines how much to adjust per unit of safety margin
	const boostFactor = 0.03 // 3% per unit of safety margin above 1.0

	// Safety margin contribution to probability boost
	// Only boost if safety margin > 1.0
	if safetyMargin > 1.0 {
		// The "room" to boost (distance to 1.0)
		room := 1.0 - marketPrice

		// Boost proportional to safety margin excess and room available
		boost := (safetyMargin - 1.0) * boostFactor * room * 10

		// Apply diminishing returns for very high safety margins
		if safetyMargin > 2.0 {
			boost = boost * 0.7 // reduce boost for very high margins
		}

		probability = marketPrice + boost
	} else if safetyMargin < 1.0 {
		// Risky position - slight reduction
		// But we don't want to reduce too much, as the market price already reflects risk
		penalty := (1.0 - safetyMargin) * 0.02 // 2% penalty per unit below 1.0
		probability = marketPrice - penalty
	}

	// Ensure probability stays within bounds [marketPrice * 0.9, 1.0]
	// We don't want to reduce probability too much below market price
	minProb := marketPrice * 0.9
	probability = math.Max(probability, minProb)
	probability = math.Min(probability, 1.0)

	return probability
}
