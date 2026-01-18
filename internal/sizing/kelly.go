package sizing

// CalculateKelly calculates the position size using the Kelly Criterion.
//
// The Kelly Criterion formula: f = (p*b - q) / b
// where:
//   - p = win probability
//   - q = lose probability (1-p)
//   - b = odds (profit per dollar risked) = (1-entryPrice)/entryPrice
//
// Parameters:
//   - entryPrice: the price to buy at (0 < price < 1)
//   - winProb: estimated probability of winning
//   - bankroll: total capital available
//   - fraction: Kelly fraction to use (0.25 = quarter Kelly, 1.0 = full Kelly)
//
// Returns the recommended position size in dollars.
func CalculateKelly(entryPrice, winProb, bankroll, fraction float64) float64 {
	// Validate inputs
	if entryPrice <= 0 || entryPrice >= 1 {
		return 0
	}
	if winProb <= 0 || winProb > 1 {
		return 0
	}
	if bankroll <= 0 {
		return 0
	}
	if fraction <= 0 {
		return 0
	}

	// Calculate odds (profit per dollar risked)
	// If you buy at $0.90, you can win $0.10 for every $0.90 risked
	// b = (1 - price) / price
	b := (1 - entryPrice) / entryPrice

	// Kelly formula: f = (p*b - q) / b
	// where p = win probability, q = 1-p
	p := winProb
	q := 1 - winProb

	// Calculate optimal fraction of bankroll
	kellyFraction := (p*b - q) / b

	// If kelly fraction is negative or zero, we have no edge
	if kellyFraction <= 0 {
		return 0
	}

	// Apply fractional Kelly (e.g., quarter Kelly for safety)
	adjustedFraction := kellyFraction * fraction

	// Calculate position size
	positionSize := bankroll * adjustedFraction

	return positionSize
}
