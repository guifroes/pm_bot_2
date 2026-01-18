package scanner

import (
	"prediction-bot/internal/config"
	"prediction-bot/internal/platform"
	"prediction-bot/pkg/types"
)

// EligibleMarket represents a market that passed all eligibility criteria
// and was successfully parsed for asset, strike, and direction information.
type EligibleMarket struct {
	Market      types.Market
	Parsed      *ParsedMarket
	Probability float64
	BetSide     string // "YES" or "NO"
}

// Scanner scans prediction market platforms for eligible markets
type Scanner struct {
	filter *EligibilityFilter
}

// NewScanner creates a new scanner with the given parameters
func NewScanner(params config.Parameters) *Scanner {
	return &Scanner{
		filter: NewEligibilityFilter(params),
	}
}

// Scan scans a single platform for eligible markets.
// It lists all active markets, filters by eligibility criteria,
// and parses market titles to extract asset, strike, and direction.
// Returns only markets that are both eligible and parseable.
func (s *Scanner) Scan(p platform.Platform) ([]EligibleMarket, error) {
	// List active markets from platform
	isActive := true
	filter := types.MarketFilter{
		IsActive: &isActive,
		Limit:    500, // Reasonable limit for single scan
	}

	markets, err := p.ListMarkets(filter)
	if err != nil {
		return nil, err
	}

	var eligible []EligibleMarket

	for _, market := range markets {
		// Check eligibility
		result := s.filter.IsEligible(market)
		if !result.Eligible {
			continue
		}

		// Parse market title to extract asset, strike, direction
		parsed, err := ParseMarketTitle(market.Title)
		if err != nil {
			// Market is eligible but title is not parseable
			// (e.g., political markets, sports, etc.)
			// Skip without error
			continue
		}

		eligible = append(eligible, EligibleMarket{
			Market:      market,
			Parsed:      parsed,
			Probability: result.Probability,
			BetSide:     result.BetSide,
		})
	}

	return eligible, nil
}
