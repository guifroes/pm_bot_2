package volatility

import (
	"math"

	"prediction-bot/pkg/types"
)

const (
	// TradingDaysCrypto is the number of trading days per year for crypto (24/7 market)
	TradingDaysCrypto = 365
	// TradingDaysStock is the number of trading days per year for stocks
	TradingDaysStock = 252
)

// CalculateVolatility calculates the annualized volatility from a series of prices.
// It uses the standard deviation of daily log returns, annualized by the appropriate factor.
// For crypto assets (isCrypto=true), it uses 365 trading days.
// For stocks (isCrypto=false), it uses 252 trading days.
// Returns 0 if there are insufficient data points (less than 2 prices).
func CalculateVolatility(prices []types.Price, isCrypto bool) float64 {
	if len(prices) < 2 {
		return 0
	}

	// Calculate daily log returns
	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1].Price <= 0 || prices[i].Price <= 0 {
			continue
		}
		returns[i-1] = math.Log(prices[i].Price / prices[i-1].Price)
	}

	if len(returns) == 0 {
		return 0
	}

	// Calculate mean of returns
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// Calculate variance (using sample variance with n-1 denominator)
	var sumSquaredDiff float64
	for _, r := range returns {
		diff := r - mean
		sumSquaredDiff += diff * diff
	}

	// Use sample standard deviation (n-1)
	variance := sumSquaredDiff / float64(len(returns)-1)
	dailyVol := math.Sqrt(variance)

	// Annualize the volatility
	var tradingDays float64
	if isCrypto {
		tradingDays = TradingDaysCrypto
	} else {
		tradingDays = TradingDaysStock
	}

	annualizedVol := dailyVol * math.Sqrt(tradingDays)

	return annualizedVol
}
