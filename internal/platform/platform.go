package platform

import "prediction-bot/pkg/types"

// Platform defines the common interface for prediction market platforms.
// Both Polymarket and Kalshi clients should implement this interface.
type Platform interface {
	// Name returns the platform identifier (e.g., "polymarket", "kalshi")
	Name() string

	// ListMarkets returns markets matching the given filter
	ListMarkets(filter types.MarketFilter) ([]types.Market, error)

	// GetOrderBook returns the order book for a given token/market
	GetOrderBook(tokenID string) (*types.OrderBook, error)

	// GetBalance returns the available balance in dollars
	GetBalance() (float64, error)

	// GetPositions returns all current positions
	GetPositions() ([]types.Position, error)
}
