package volatility

import (
	"testing"
	"time"

	"prediction-bot/pkg/types"
)

func TestCalculateVolatility_CryptoAsset(t *testing.T) {
	// Generate synthetic price data with known volatility characteristics
	// Using daily prices over 14 days to calculate volatility
	basePrices := []float64{
		100000, 101500, 99800, 102000, 100500,
		103000, 101000, 104000, 102500, 105000,
		103500, 106000, 104500, 107000,
	}

	prices := make([]types.Price, len(basePrices))
	baseTime := time.Now()
	for i, p := range basePrices {
		prices[i] = types.Price{
			Symbol:    "BTCUSDT",
			Price:     p,
			Timestamp: baseTime.Add(time.Duration(-len(basePrices)+i+1) * 24 * time.Hour),
			Source:    "binance",
		}
	}

	vol := CalculateVolatility(prices, true)

	// Volatility should be annualized and between 0 and 2 for crypto
	if vol <= 0 || vol > 2 {
		t.Errorf("Expected volatility between 0 and 2, got %f", vol)
	}

	t.Logf("Calculated crypto volatility: %.4f (%.2f%%)", vol, vol*100)
}

func TestCalculateVolatility_StockAsset(t *testing.T) {
	// Generate synthetic stock price data
	basePrices := []float64{
		450.0, 452.5, 448.0, 455.0, 453.0,
		458.0, 456.0, 460.0, 457.5, 462.0,
		459.0, 465.0, 463.0, 468.0,
	}

	prices := make([]types.Price, len(basePrices))
	baseTime := time.Now()
	for i, p := range basePrices {
		prices[i] = types.Price{
			Symbol:    "SPY",
			Price:     p,
			Timestamp: baseTime.Add(time.Duration(-len(basePrices)+i+1) * 24 * time.Hour),
			Source:    "alphavantage",
		}
	}

	vol := CalculateVolatility(prices, false)

	// Stock volatility should be positive and reasonable
	if vol <= 0 || vol > 2 {
		t.Errorf("Expected volatility between 0 and 2, got %f", vol)
	}

	t.Logf("Calculated stock volatility: %.4f (%.2f%%)", vol, vol*100)
}

func TestCalculateVolatility_InsufficientData(t *testing.T) {
	// With only 1 price, cannot calculate volatility
	prices := []types.Price{
		{Symbol: "BTCUSDT", Price: 100000, Timestamp: time.Now(), Source: "binance"},
	}

	vol := CalculateVolatility(prices, true)

	// Should return 0 or some indicator of insufficient data
	if vol != 0 {
		t.Errorf("Expected 0 volatility for insufficient data, got %f", vol)
	}
}

func TestCalculateVolatility_EmptyData(t *testing.T) {
	var prices []types.Price

	vol := CalculateVolatility(prices, true)

	if vol != 0 {
		t.Errorf("Expected 0 volatility for empty data, got %f", vol)
	}
}

func TestCalculateVolatility_CryptoUsesMoreTradingDays(t *testing.T) {
	// Create identical price movements for crypto and stock
	basePrices := []float64{100, 102, 101, 103, 102, 104, 103, 105, 104, 106}

	cryptoPrices := make([]types.Price, len(basePrices))
	stockPrices := make([]types.Price, len(basePrices))
	baseTime := time.Now()

	for i, p := range basePrices {
		cryptoPrices[i] = types.Price{
			Symbol:    "BTCUSDT",
			Price:     p,
			Timestamp: baseTime.Add(time.Duration(-len(basePrices)+i+1) * 24 * time.Hour),
			Source:    "binance",
		}
		stockPrices[i] = types.Price{
			Symbol:    "SPY",
			Price:     p,
			Timestamp: baseTime.Add(time.Duration(-len(basePrices)+i+1) * 24 * time.Hour),
			Source:    "alphavantage",
		}
	}

	cryptoVol := CalculateVolatility(cryptoPrices, true)  // 365 days
	stockVol := CalculateVolatility(stockPrices, false)   // 252 days

	// Crypto volatility should be higher because it's annualized with more days
	// sqrt(365) > sqrt(252), so crypto vol should be ~20% higher
	ratio := cryptoVol / stockVol
	expectedRatio := 1.203 // sqrt(365/252) ≈ 1.203

	if ratio < 1.1 || ratio > 1.3 {
		t.Errorf("Expected crypto/stock ratio around %.3f, got %.3f", expectedRatio, ratio)
	}

	t.Logf("Crypto vol: %.4f, Stock vol: %.4f, Ratio: %.4f", cryptoVol, stockVol, ratio)
}
