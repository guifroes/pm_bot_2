package kalshi

import (
	"encoding/json"
	"fmt"
	"time"

	"prediction-bot/pkg/types"
)

// kalshiPosition represents a position as returned by the Kalshi API.
type kalshiPosition struct {
	MarketTicker         string `json:"market_ticker"`
	Position             int    `json:"position"`
	TotalTraded          int    `json:"total_traded"`
	RealizedPnL          int    `json:"realized_pnl"`
	RestingOrdersCount   int    `json:"resting_orders_count"`
	FeesPaid             int    `json:"fees_paid"`
	MarketExposure       int    `json:"market_exposure"`
}

// positionsResponse represents the API response for positions.
type positionsResponse struct {
	Cursor          string           `json:"cursor"`
	MarketPositions []kalshiPosition `json:"market_positions"`
}

// GetPositions returns all open positions for the account.
func (c *Client) GetPositions() ([]types.Position, error) {
	path := BuildURL("/portfolio/positions", map[string]string{
		"limit": "100",
	})

	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	var response positionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse positions response: %w", err)
	}

	positions := make([]types.Position, 0, len(response.MarketPositions))
	now := time.Now()

	for _, kp := range response.MarketPositions {
		pos := types.Position{
			Platform:         "kalshi",
			MarketTicker:     kp.MarketTicker,
			Quantity:         kp.Position,
			MarketExposure:   float64(kp.MarketExposure) / 100.0, // Convert cents to dollars
			RealizedPnL:      float64(kp.RealizedPnL) / 100.0,    // Convert cents to dollars
			TotalTraded:      kp.TotalTraded,
			FeesPaid:         float64(kp.FeesPaid) / 100.0, // Convert cents to dollars
			RestingOrdersQty: kp.RestingOrdersCount,
			Timestamp:        now,
		}
		positions = append(positions, pos)
	}

	return positions, nil
}
