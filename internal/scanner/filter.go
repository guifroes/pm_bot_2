package scanner

import (
	"fmt"
	"time"

	"prediction-bot/internal/config"
	"prediction-bot/pkg/types"
)

const (
	// MaxTimeToResolution is the maximum time allowed until market closes (48 hours)
	MaxTimeToResolution = 48 * time.Hour

	// MinLiquidity is the minimum liquidity required in dollars
	MinLiquidity = 100.0
)

// EligibilityResult contains the result of eligibility check
type EligibilityResult struct {
	Eligible    bool
	Reasons     []string
	Probability float64
	BetSide     string // "YES" or "NO"
}

// EligibilityFilter checks if markets meet the eligibility criteria
type EligibilityFilter struct {
	params config.Parameters
}

// NewEligibilityFilter creates a new eligibility filter with the given parameters
func NewEligibilityFilter(params config.Parameters) *EligibilityFilter {
	return &EligibilityFilter{
		params: params,
	}
}

// IsEligible checks if a market meets all eligibility criteria
func (f *EligibilityFilter) IsEligible(market types.Market) EligibilityResult {
	result := EligibilityResult{
		Eligible: true,
		Reasons:  []string{},
	}

	// Determine best probability and bet side
	yesProbability := market.OutcomeYesPrice
	noProbability := market.OutcomeNoPrice

	if yesProbability >= noProbability {
		result.Probability = yesProbability
		result.BetSide = "YES"
	} else {
		result.Probability = noProbability
		result.BetSide = "NO"
	}

	// Check if market is active
	if !market.Active {
		result.Eligible = false
		result.Reasons = append(result.Reasons, "market is not active")
	}

	// Check if market is already closed
	if market.Closed {
		result.Eligible = false
		result.Reasons = append(result.Reasons, "market is already closed")
	}

	// Check probability threshold
	if result.Probability < f.params.ProbabilityThreshold {
		result.Eligible = false
		result.Reasons = append(result.Reasons,
			fmt.Sprintf("probability %.2f%% is below threshold %.2f%%",
				result.Probability*100, f.params.ProbabilityThreshold*100))
	}

	// Check time to resolution
	timeToResolution := time.Until(market.EndDate)
	if timeToResolution > MaxTimeToResolution {
		result.Eligible = false
		result.Reasons = append(result.Reasons,
			fmt.Sprintf("time to resolution %.1fh exceeds max %.1fh",
				timeToResolution.Hours(), MaxTimeToResolution.Hours()))
	}

	// Check if market has already ended
	if timeToResolution < 0 {
		result.Eligible = false
		result.Reasons = append(result.Reasons, "market has already ended")
	}

	// Check liquidity
	if market.Liquidity < MinLiquidity {
		result.Eligible = false
		result.Reasons = append(result.Reasons,
			fmt.Sprintf("liquidity $%.2f is below minimum $%.2f",
				market.Liquidity, MinLiquidity))
	}

	return result
}
