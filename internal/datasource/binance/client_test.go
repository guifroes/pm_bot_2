package binance

import (
	"testing"
)

func TestGetPrice_BTCUSDT_ReturnsPositivePrice(t *testing.T) {
	client := NewClient()

	price, err := client.GetPrice("BTCUSDT")
	if err != nil {
		t.Fatalf("GetPrice: %v", err)
	}

	if price.Price <= 0 {
		t.Errorf("expected positive price, got %f", price.Price)
	}

	if price.Symbol != "BTCUSDT" {
		t.Errorf("expected symbol BTCUSDT, got %s", price.Symbol)
	}

	if price.Source != "binance" {
		t.Errorf("expected source binance, got %s", price.Source)
	}

	t.Logf("BTC price: $%.2f", price.Price)
}

func TestGetPrice_ETHUSDT_ReturnsPositivePrice(t *testing.T) {
	client := NewClient()

	price, err := client.GetPrice("ETHUSDT")
	if err != nil {
		t.Fatalf("GetPrice: %v", err)
	}

	if price.Price <= 0 {
		t.Errorf("expected positive price, got %f", price.Price)
	}

	t.Logf("ETH price: $%.2f", price.Price)
}

func TestGetPrice_InvalidSymbol_ReturnsError(t *testing.T) {
	client := NewClient()

	_, err := client.GetPrice("INVALIDXYZ")
	if err == nil {
		t.Error("expected error for invalid symbol, got nil")
	}
}

func TestGetHistory_BTCUSDT_Returns336Points(t *testing.T) {
	client := NewClient()

	prices, err := client.GetHistory("BTCUSDT", 336)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}

	if len(prices) != 336 {
		t.Errorf("expected 336 prices, got %d", len(prices))
	}

	// Verify all prices are positive
	for i, p := range prices {
		if p.Price <= 0 {
			t.Errorf("price at index %d is not positive: %f", i, p.Price)
			break
		}
	}

	// Verify timestamps are in ascending order
	for i := 1; i < len(prices); i++ {
		if prices[i].Timestamp.Before(prices[i-1].Timestamp) {
			t.Errorf("timestamps not in ascending order at index %d", i)
			break
		}
	}

	t.Logf("Got %d price points, oldest: %s, newest: %s",
		len(prices),
		prices[0].Timestamp.Format("2006-01-02 15:04"),
		prices[len(prices)-1].Timestamp.Format("2006-01-02 15:04"))
}
