package datasource

import "strings"

// SymbolMapping contains the mapping from a common name to exchange symbols.
type SymbolMapping struct {
	CommonName    string
	BinanceSymbol string
	AlphaSymbol   string
	IsCrypto      bool
}

// SymbolMapper maps common asset names to exchange-specific symbols.
type SymbolMapper struct {
	mappings map[string]SymbolMapping
}

// NewSymbolMapper creates a new symbol mapper with default mappings.
func NewSymbolMapper() *SymbolMapper {
	m := &SymbolMapper{
		mappings: make(map[string]SymbolMapping),
	}

	// Cryptocurrencies (Binance)
	m.addMapping(SymbolMapping{
		CommonName:    "Bitcoin",
		BinanceSymbol: "BTCUSDT",
		IsCrypto:      true,
	})
	m.addMapping(SymbolMapping{
		CommonName:    "BTC",
		BinanceSymbol: "BTCUSDT",
		IsCrypto:      true,
	})
	m.addMapping(SymbolMapping{
		CommonName:    "Ethereum",
		BinanceSymbol: "ETHUSDT",
		IsCrypto:      true,
	})
	m.addMapping(SymbolMapping{
		CommonName:    "ETH",
		BinanceSymbol: "ETHUSDT",
		IsCrypto:      true,
	})
	m.addMapping(SymbolMapping{
		CommonName:    "Solana",
		BinanceSymbol: "SOLUSDT",
		IsCrypto:      true,
	})
	m.addMapping(SymbolMapping{
		CommonName:    "SOL",
		BinanceSymbol: "SOLUSDT",
		IsCrypto:      true,
	})

	// Stocks/ETFs (Alpha Vantage)
	m.addMapping(SymbolMapping{
		CommonName:  "S&P 500",
		AlphaSymbol: "SPY",
		IsCrypto:    false,
	})
	m.addMapping(SymbolMapping{
		CommonName:  "SPY",
		AlphaSymbol: "SPY",
		IsCrypto:    false,
	})
	m.addMapping(SymbolMapping{
		CommonName:  "Nasdaq",
		AlphaSymbol: "QQQ",
		IsCrypto:    false,
	})
	m.addMapping(SymbolMapping{
		CommonName:  "QQQ",
		AlphaSymbol: "QQQ",
		IsCrypto:    false,
	})

	return m
}

func (m *SymbolMapper) addMapping(mapping SymbolMapping) {
	// Store by lowercase common name for case-insensitive lookup
	key := strings.ToLower(mapping.CommonName)
	m.mappings[key] = mapping
}

// Lookup finds the mapping for a common asset name.
// Returns the mapping and true if found, or empty mapping and false if not.
func (m *SymbolMapper) Lookup(commonName string) (SymbolMapping, bool) {
	key := strings.ToLower(commonName)
	mapping, ok := m.mappings[key]
	return mapping, ok
}

// IsCrypto returns true if the asset is a cryptocurrency.
func (m *SymbolMapper) IsCrypto(commonName string) bool {
	mapping, ok := m.Lookup(commonName)
	if !ok {
		return false
	}
	return mapping.IsCrypto
}
