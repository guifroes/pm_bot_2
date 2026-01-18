package datasource

import (
	"fmt"

	"prediction-bot/internal/datasource/alphavantage"
	"prediction-bot/internal/datasource/binance"
	"prediction-bot/pkg/types"
)

// Aggregator routes price requests to the appropriate data source.
type Aggregator struct {
	mapper     *SymbolMapper
	binance    *binance.Client
	alphaVantage *alphavantage.Client
}

// NewAggregator creates a new data source aggregator.
// alphaVantageKey can be empty if Alpha Vantage is not needed.
func NewAggregator(alphaVantageKey string) *Aggregator {
	var avClient *alphavantage.Client
	if alphaVantageKey != "" {
		avClient = alphavantage.NewClientWithKey(alphaVantageKey)
	}

	return &Aggregator{
		mapper:       NewSymbolMapper(),
		binance:      binance.NewClient(),
		alphaVantage: avClient,
	}
}

// GetPrice fetches the current price for an asset, routing to the appropriate source.
func (a *Aggregator) GetPrice(asset string) (types.Price, error) {
	mapping, ok := a.mapper.Lookup(asset)
	if !ok {
		return types.Price{}, fmt.Errorf("unknown asset: %s", asset)
	}

	if mapping.IsCrypto {
		return a.binance.GetPrice(mapping.BinanceSymbol)
	}

	if a.alphaVantage == nil {
		return types.Price{}, fmt.Errorf("Alpha Vantage not configured for asset: %s", asset)
	}

	return a.alphaVantage.GetPrice(mapping.AlphaSymbol)
}

// GetHistory fetches historical prices for an asset.
// Currently only supported for crypto assets via Binance.
func (a *Aggregator) GetHistory(asset string, hours int) ([]types.Price, error) {
	mapping, ok := a.mapper.Lookup(asset)
	if !ok {
		return nil, fmt.Errorf("unknown asset: %s", asset)
	}

	if !mapping.IsCrypto {
		return nil, fmt.Errorf("historical data not supported for stocks yet: %s", asset)
	}

	return a.binance.GetHistory(mapping.BinanceSymbol, hours)
}

// IsCrypto returns true if the asset is a cryptocurrency.
func (a *Aggregator) IsCrypto(asset string) bool {
	return a.mapper.IsCrypto(asset)
}
