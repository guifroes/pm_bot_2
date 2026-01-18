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
