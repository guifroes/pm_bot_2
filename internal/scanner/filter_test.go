package scanner

import (
	"testing"
	"time"

	"prediction-bot/internal/config"
	"prediction-bot/pkg/types"
)

func TestIsEligible_MeetsAllCriteria(t *testing.T) {
	// Market: prob=85%, closes=24h, liquidity=$500 → eligible=true
	params := config.Parameters{
		ProbabilityThreshold:   0.80,
		VolatilitySafetyMargin: 1.5,
		StopLossPercent:        0.15,
		KellyFraction:          0.25,
	}

	market := types.Market{
		ID:              "test-market-1",
		Platform:        "polymarket",
		Title:           "Will Bitcoin be above $100,000 on Jan 20?",
		EndDate:         time.Now().Add(24 * time.Hour),
		Liquidity:       500.0,
		Active:          true,
		OutcomeYesPrice: 0.85, // 85% probability
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if !result.Eligible {
		t.Errorf("Expected market to be eligible, got ineligible. Reasons: %v", result.Reasons)
	}
}

func TestIsEligible_ProbabilityTooLow(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-2",
		EndDate:         time.Now().Add(24 * time.Hour),
		Liquidity:       500.0,
		Active:          true,
		OutcomeYesPrice: 0.75, // 75% < 80% threshold
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if result.Eligible {
		t.Errorf("Expected market to be ineligible due to low probability")
	}

	if !containsReason(result.Reasons, "probability") {
		t.Errorf("Expected reason to mention probability, got: %v", result.Reasons)
	}
}

func TestIsEligible_TimeTooLong(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-3",
		EndDate:         time.Now().Add(72 * time.Hour), // 72h > 48h max
		Liquidity:       500.0,
		Active:          true,
		OutcomeYesPrice: 0.85,
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if result.Eligible {
		t.Errorf("Expected market to be ineligible due to time too long")
	}

	if !containsReason(result.Reasons, "time") {
		t.Errorf("Expected reason to mention time, got: %v", result.Reasons)
	}
}

func TestIsEligible_LiquidityTooLow(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-4",
		EndDate:         time.Now().Add(24 * time.Hour),
		Liquidity:       50.0, // $50 < $100 min
		Active:          true,
		OutcomeYesPrice: 0.85,
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if result.Eligible {
		t.Errorf("Expected market to be ineligible due to low liquidity")
	}

	if !containsReason(result.Reasons, "liquidity") {
		t.Errorf("Expected reason to mention liquidity, got: %v", result.Reasons)
	}
}

func TestIsEligible_MarketNotActive(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-5",
		EndDate:         time.Now().Add(24 * time.Hour),
		Liquidity:       500.0,
		Active:          false, // Not active
		OutcomeYesPrice: 0.85,
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if result.Eligible {
		t.Errorf("Expected market to be ineligible due to not being active")
	}

	if !containsReason(result.Reasons, "active") {
		t.Errorf("Expected reason to mention active, got: %v", result.Reasons)
	}
}

func TestIsEligible_MarketAlreadyClosed(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-6",
		EndDate:         time.Now().Add(24 * time.Hour),
		Liquidity:       500.0,
		Active:          true,
		Closed:          true, // Already closed
		OutcomeYesPrice: 0.85,
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if result.Eligible {
		t.Errorf("Expected market to be ineligible due to being closed")
	}

	if !containsReason(result.Reasons, "closed") {
		t.Errorf("Expected reason to mention closed, got: %v", result.Reasons)
	}
}

func TestIsEligible_MultipleFailures(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-7",
		EndDate:         time.Now().Add(72 * time.Hour), // Too long
		Liquidity:       50.0,                           // Too low
		Active:          true,
		OutcomeYesPrice: 0.75, // Too low
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if result.Eligible {
		t.Errorf("Expected market to be ineligible")
	}

	// Should have at least 3 reasons
	if len(result.Reasons) < 3 {
		t.Errorf("Expected at least 3 reasons, got %d: %v", len(result.Reasons), result.Reasons)
	}
}

func TestIsEligible_EdgeCaseExactThreshold(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-8",
		EndDate:         time.Now().Add(48 * time.Hour), // Exactly 48h
		Liquidity:       100.0,                          // Exactly $100
		Active:          true,
		OutcomeYesPrice: 0.80, // Exactly 80%
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if !result.Eligible {
		t.Errorf("Expected market to be eligible at exact thresholds, got ineligible. Reasons: %v", result.Reasons)
	}
}

func TestIsEligible_HighProbabilityNoOutcome(t *testing.T) {
	// Test with OutcomeNoPrice being the high probability (betting NO)
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-9",
		EndDate:         time.Now().Add(24 * time.Hour),
		Liquidity:       500.0,
		Active:          true,
		OutcomeYesPrice: 0.10, // YES is low
		OutcomeNoPrice:  0.90, // NO is high probability
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if !result.Eligible {
		t.Errorf("Expected market to be eligible (high NO probability), got ineligible. Reasons: %v", result.Reasons)
	}

	if result.BetSide != "NO" {
		t.Errorf("Expected bet side to be NO, got: %s", result.BetSide)
	}
}

func TestIsEligible_ReturnsCorrectProbability(t *testing.T) {
	params := config.Parameters{
		ProbabilityThreshold: 0.80,
	}

	market := types.Market{
		ID:              "test-market-10",
		EndDate:         time.Now().Add(24 * time.Hour),
		Liquidity:       500.0,
		Active:          true,
		OutcomeYesPrice: 0.92,
	}

	filter := NewEligibilityFilter(params)
	result := filter.IsEligible(market)

	if result.Probability != 0.92 {
		t.Errorf("Expected probability 0.92, got: %f", result.Probability)
	}

	if result.BetSide != "YES" {
		t.Errorf("Expected bet side to be YES, got: %s", result.BetSide)
	}
}

// Helper function to check if any reason contains a substring
func containsReason(reasons []string, substr string) bool {
	for _, r := range reasons {
		if containsIgnoreCase(r, substr) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[0] == substr[0] || s[0]+32 == substr[0] || s[0] == substr[0]+32) &&
			containsIgnoreCase(s[1:], substr[1:])) ||
		(len(s) > 0 && containsIgnoreCase(s[1:], substr)))
}
