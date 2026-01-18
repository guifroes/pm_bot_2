package scanner

import (
	"testing"
)

func TestParseMarketTitle_Bitcoin_Above(t *testing.T) {
	title := "Will Bitcoin be above $100,000 on Jan 18?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Asset != "BTC" {
		t.Errorf("expected Asset='BTC', got '%s'", result.Asset)
	}

	if result.Strike != 100000 {
		t.Errorf("expected Strike=100000, got %f", result.Strike)
	}

	if result.Direction != "above" {
		t.Errorf("expected Direction='above', got '%s'", result.Direction)
	}
}

func TestParseMarketTitle_Bitcoin_Below(t *testing.T) {
	title := "Will Bitcoin fall below $95,000 by Friday?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Asset != "BTC" {
		t.Errorf("expected Asset='BTC', got '%s'", result.Asset)
	}

	if result.Strike != 95000 {
		t.Errorf("expected Strike=95000, got %f", result.Strike)
	}

	if result.Direction != "below" {
		t.Errorf("expected Direction='below', got '%s'", result.Direction)
	}
}

func TestParseMarketTitle_Ethereum_Over(t *testing.T) {
	title := "Ethereum over $4,500 by end of week?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Asset != "ETH" {
		t.Errorf("expected Asset='ETH', got '%s'", result.Asset)
	}

	if result.Strike != 4500 {
		t.Errorf("expected Strike=4500, got %f", result.Strike)
	}

	if result.Direction != "above" {
		t.Errorf("expected Direction='above', got '%s'", result.Direction)
	}
}

func TestParseMarketTitle_SP500_Above(t *testing.T) {
	title := "S&P 500 above 5000 on January 20th?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Asset != "SPY" {
		t.Errorf("expected Asset='SPY', got '%s'", result.Asset)
	}

	if result.Strike != 5000 {
		t.Errorf("expected Strike=5000, got %f", result.Strike)
	}

	if result.Direction != "above" {
		t.Errorf("expected Direction='above', got '%s'", result.Direction)
	}
}

func TestParseMarketTitle_BTC_Symbol(t *testing.T) {
	title := "BTC above $100k?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Asset != "BTC" {
		t.Errorf("expected Asset='BTC', got '%s'", result.Asset)
	}

	if result.Strike != 100000 {
		t.Errorf("expected Strike=100000, got %f", result.Strike)
	}

	if result.Direction != "above" {
		t.Errorf("expected Direction='above', got '%s'", result.Direction)
	}
}

func TestParseMarketTitle_Under(t *testing.T) {
	title := "Will ETH be under $3,000 tomorrow?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Asset != "ETH" {
		t.Errorf("expected Asset='ETH', got '%s'", result.Asset)
	}

	if result.Strike != 3000 {
		t.Errorf("expected Strike=3000, got %f", result.Strike)
	}

	if result.Direction != "below" {
		t.Errorf("expected Direction='below', got '%s'", result.Direction)
	}
}

func TestParseMarketTitle_InvalidNoAsset(t *testing.T) {
	title := "Will price be above $100,000?"

	_, err := ParseMarketTitle(title)

	if err == nil {
		t.Error("expected error for missing asset, got nil")
	}
}

func TestParseMarketTitle_InvalidNoStrike(t *testing.T) {
	title := "Will Bitcoin be above target?"

	_, err := ParseMarketTitle(title)

	if err == nil {
		t.Error("expected error for missing strike, got nil")
	}
}

func TestParseMarketTitle_InvalidNoDirection(t *testing.T) {
	title := "Bitcoin $100,000 on Jan 18?"

	_, err := ParseMarketTitle(title)

	if err == nil {
		t.Error("expected error for missing direction, got nil")
	}
}

func TestParseMarketTitle_Solana(t *testing.T) {
	title := "Solana above $200 by weekend?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Asset != "SOL" {
		t.Errorf("expected Asset='SOL', got '%s'", result.Asset)
	}

	if result.Strike != 200 {
		t.Errorf("expected Strike=200, got %f", result.Strike)
	}
}

func TestParseMarketTitle_AtOrAbove(t *testing.T) {
	title := "Will Bitcoin be at or above $100,000?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Direction != "above" {
		t.Errorf("expected Direction='above', got '%s'", result.Direction)
	}
}

func TestParseMarketTitle_AtOrBelow(t *testing.T) {
	title := "Will Bitcoin be at or below $90,000?"

	result, err := ParseMarketTitle(title)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Direction != "below" {
		t.Errorf("expected Direction='below', got '%s'", result.Direction)
	}
}
