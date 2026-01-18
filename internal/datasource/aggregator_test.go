package datasource

import (
	"testing"
)

func TestAggregator_GetPrice_Bitcoin_RoutesBinance(t *testing.T) {
	agg := NewAggregator("")

	price, err := agg.GetPrice("Bitcoin")
	if err != nil {
		t.Fatalf("GetPrice: %v", err)
	}

	if price.Price <= 0 {
		t.Errorf("expected positive price, got %f", price.Price)
	}

	if price.Source != "binance" {
		t.Errorf("expected source binance, got %s", price.Source)
	}

	if price.Symbol != "BTCUSDT" {
		t.Errorf("expected symbol BTCUSDT, got %s", price.Symbol)
	}

	t.Logf("Bitcoin price: $%.2f (via %s)", price.Price, price.Source)
}

func TestAggregator_GetPrice_BTC_RoutesBinance(t *testing.T) {
	agg := NewAggregator("")

	price, err := agg.GetPrice("BTC")
	if err != nil {
		t.Fatalf("GetPrice: %v", err)
	}

	if price.Source != "binance" {
		t.Errorf("expected source binance, got %s", price.Source)
	}

	t.Logf("BTC price: $%.2f", price.Price)
}

func TestAggregator_GetPrice_CaseInsensitive(t *testing.T) {
	agg := NewAggregator("")

	testCases := []string{"bitcoin", "BITCOIN", "Bitcoin", "BiTcOiN"}
	for _, tc := range testCases {
		price, err := agg.GetPrice(tc)
		if err != nil {
			t.Errorf("GetPrice(%s): %v", tc, err)
			continue
		}
		if price.Price <= 0 {
			t.Errorf("GetPrice(%s): expected positive price", tc)
		}
	}
}

func TestAggregator_GetPrice_UnknownAsset_ReturnsError(t *testing.T) {
	agg := NewAggregator("")

	_, err := agg.GetPrice("UnknownAsset123")
	if err == nil {
		t.Error("expected error for unknown asset, got nil")
	}
}

func TestAggregator_GetHistory_Bitcoin(t *testing.T) {
	agg := NewAggregator("")

	prices, err := agg.GetHistory("Bitcoin", 24)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}

	if len(prices) != 24 {
		t.Errorf("expected 24 prices, got %d", len(prices))
	}

	t.Logf("Got %d Bitcoin price points", len(prices))
}

func TestAggregator_IsCrypto(t *testing.T) {
	agg := NewAggregator("")

	if !agg.IsCrypto("Bitcoin") {
		t.Error("Bitcoin should be crypto")
	}
	if !agg.IsCrypto("ETH") {
		t.Error("ETH should be crypto")
	}
	if agg.IsCrypto("SPY") {
		t.Error("SPY should not be crypto")
	}
	if agg.IsCrypto("S&P 500") {
		t.Error("S&P 500 should not be crypto")
	}
}

func TestSymbolMapper_Lookup(t *testing.T) {
	m := NewSymbolMapper()

	testCases := []struct {
		input    string
		expected string
		isCrypto bool
	}{
		{"Bitcoin", "BTCUSDT", true},
		{"bitcoin", "BTCUSDT", true},
		{"BTC", "BTCUSDT", true},
		{"Ethereum", "ETHUSDT", true},
		{"S&P 500", "SPY", false},
		{"SPY", "SPY", false},
	}

	for _, tc := range testCases {
		mapping, ok := m.Lookup(tc.input)
		if !ok {
			t.Errorf("Lookup(%s): not found", tc.input)
			continue
		}
		if tc.isCrypto {
			if mapping.BinanceSymbol != tc.expected {
				t.Errorf("Lookup(%s): expected BinanceSymbol %s, got %s",
					tc.input, tc.expected, mapping.BinanceSymbol)
			}
		} else {
			if mapping.AlphaSymbol != tc.expected {
				t.Errorf("Lookup(%s): expected AlphaSymbol %s, got %s",
					tc.input, tc.expected, mapping.AlphaSymbol)
			}
		}
	}
}
