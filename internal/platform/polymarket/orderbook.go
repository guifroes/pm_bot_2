package polymarket

import (
	"encoding/json"
	"fmt"
	"strconv"

	"prediction-bot/pkg/types"
)

// polymarketOrderBook represents the Polymarket API order book response.
type polymarketOrderBook struct {
	Market   string                `json:"market"`
	AssetID  string                `json:"asset_id"`
	Hash     string                `json:"hash"`
	Bids     []polymarketBookLevel `json:"bids"`
	Asks     []polymarketBookLevel `json:"asks"`
}

type polymarketBookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// GetOrderBook fetches the order book for a specific token.
func (c *Client) GetOrderBook(tokenID string) (*types.OrderBook, error) {
	path := fmt.Sprintf("/book?token_id=%s", tokenID)

	body, err := c.doPublicRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("get order book: %w", err)
	}

	var ob polymarketOrderBook
	if err := json.Unmarshal(body, &ob); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	result := &types.OrderBook{
		MarketID: ob.Market,
		TokenID:  ob.AssetID,
		Bids:     make([]types.Level, 0, len(ob.Bids)),
		Asks:     make([]types.Level, 0, len(ob.Asks)),
	}

	// Convert bids (already sorted highest to lowest)
	for _, b := range ob.Bids {
		price, err := strconv.ParseFloat(b.Price, 64)
		if err != nil {
			continue
		}
		size, err := strconv.ParseFloat(b.Size, 64)
		if err != nil {
			continue
		}
		result.Bids = append(result.Bids, types.Level{Price: price, Size: size})
	}

	// Convert asks (already sorted lowest to highest)
	for _, a := range ob.Asks {
		price, err := strconv.ParseFloat(a.Price, 64)
		if err != nil {
			continue
		}
		size, err := strconv.ParseFloat(a.Size, 64)
		if err != nil {
			continue
		}
		result.Asks = append(result.Asks, types.Level{Price: price, Size: size})
	}

	return result, nil
}

// GetMarketOrderBooks fetches order books for all tokens in a market.
func (c *Client) GetMarketOrderBooks(conditionID string) (map[string]*types.OrderBook, error) {
	// First get the market to find token IDs
	market, err := c.GetMarket(conditionID)
	if err != nil {
		return nil, fmt.Errorf("get market: %w", err)
	}

	result := make(map[string]*types.OrderBook)
	for _, token := range market.Tokens {
		ob, err := c.GetOrderBook(token.TokenID)
		if err != nil {
			// Continue with other tokens if one fails
			continue
		}
		result[token.Outcome] = ob
	}

	return result, nil
}
