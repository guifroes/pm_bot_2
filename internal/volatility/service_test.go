package volatility

import (
	"os"
	"testing"
	"time"
)

func TestVolatilityService_AnalyzeAsset(t *testing.T) {
	// Skip if no network access
	if os.Getenv("SKIP_NETWORK_TESTS") == "1" {
		t.Skip("Skipping network test")
	}

	// Create service with real aggregator (no Alpha Vantage key needed for crypto)
	service := NewService("")

	// Test: AnalyzeAsset("BTC", $100000, "above", 24h) busca dados reais e retorna análise
	result, err := service.AnalyzeAsset("BTC", 100000, DirectionAbove, 24*time.Hour)
	if err != nil {
		t.Fatalf("AnalyzeAsset failed: %v", err)
	}

	// Verify result has valid data
	if result.CurrentPrice <= 0 {
		t.Errorf("Expected CurrentPrice > 0, got %f", result.CurrentPrice)
	}

	if result.Volatility <= 0 {
		t.Errorf("Expected Volatility > 0, got %f", result.Volatility)
	}

	// Safety margin can be negative if price is on wrong side of strike
	// but should have a reasonable value (not zero from missing data)
	if result.ExpectedMove <= 0 {
		t.Errorf("Expected ExpectedMove > 0, got %f", result.ExpectedMove)
	}

	// Recommendation should be set
	if result.Recommendation != RecommendationValid &&
		result.Recommendation != RecommendationRisky &&
		result.Recommendation != RecommendationReject {
		t.Errorf("Expected valid recommendation, got %s", result.Recommendation)
	}

	t.Logf("BTC Analysis Result:")
	t.Logf("  Current Price: $%.2f", result.CurrentPrice)
	t.Logf("  Strike Price: $%.2f", result.StrikePrice)
	t.Logf("  Volatility: %.4f (%.2f%%)", result.Volatility, result.Volatility*100)
	t.Logf("  Distance to Strike: %.4f (%.2f%%)", result.DistanceToStrike, result.DistanceToStrike*100)
	t.Logf("  Expected Move: %.4f (%.2f%%)", result.ExpectedMove, result.ExpectedMove*100)
	t.Logf("  Safety Margin: %.2f", result.SafetyMargin)
	t.Logf("  Recommendation: %s", result.Recommendation)
}

func TestVolatilityService_AnalyzeAsset_UnknownAsset(t *testing.T) {
	service := NewService("")

	_, err := service.AnalyzeAsset("UNKNOWNASSET", 100, DirectionAbove, 24*time.Hour)
	if err == nil {
		t.Error("Expected error for unknown asset, got nil")
	}
}

func TestVolatilityService_AnalyzeAsset_ETH(t *testing.T) {
	if os.Getenv("SKIP_NETWORK_TESTS") == "1" {
		t.Skip("Skipping network test")
	}

	service := NewService("")

	// Test with ETH to verify other crypto assets work
	result, err := service.AnalyzeAsset("ETH", 3000, DirectionBelow, 48*time.Hour)
	if err != nil {
		t.Fatalf("AnalyzeAsset for ETH failed: %v", err)
	}

	if result.CurrentPrice <= 0 {
		t.Errorf("Expected CurrentPrice > 0 for ETH, got %f", result.CurrentPrice)
	}

	t.Logf("ETH Analysis Result:")
	t.Logf("  Current Price: $%.2f", result.CurrentPrice)
	t.Logf("  Safety Margin: %.2f", result.SafetyMargin)
	t.Logf("  Recommendation: %s", result.Recommendation)
}
