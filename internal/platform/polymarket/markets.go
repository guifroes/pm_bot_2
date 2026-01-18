package polymarket

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"prediction-bot/pkg/types"
)

// polymarketMarket represents the Polymarket API market response.
type polymarketMarket struct {
	ConditionID    string  `json:"condition_id"`
	QuestionID     string  `json:"question_id"`
	Question       string  `json:"question"`
	Description    string  `json:"description"`
	EndDateISO     string  `json:"end_date_iso"`
	GameStartTime  string  `json:"game_start_time"`
	Active         bool    `json:"active"`
	Closed         bool    `json:"closed"`
	MarketSlug     string  `json:"market_slug"`
	MinIncentiveSizeQual float64 `json:"minimum_order_size"`
	MinTickSize    float64 `json:"minimum_tick_size"`
	Tokens         []polymarketToken `json:"tokens"`
}

type polymarketToken struct {
	TokenID  string  `json:"token_id"`
	Outcome  string  `json:"outcome"`
	Price    float64 `json:"price"`
	Winner   bool    `json:"winner"`
}

// marketsResponse is the response from the markets endpoint.
type marketsResponse struct {
	Data       []polymarketMarket `json:"data"`
	NextCursor string             `json:"next_cursor"`
}

// ListMarkets fetches markets from Polymarket API.
func (c *Client) ListMarkets(filter types.MarketFilter) ([]types.Market, error) {
	// Build query parameters
	params := url.Values{}

	if filter.IsActive != nil {
		params.Set("active", strconv.FormatBool(*filter.IsActive))
	}

	if filter.Closed != nil {
		params.Set("closed", strconv.FormatBool(*filter.Closed))
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	params.Set("limit", strconv.Itoa(limit))

	if filter.Offset > 0 {
		params.Set("offset", strconv.Itoa(filter.Offset))
	}

	path := "/markets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := c.doPublicRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("list markets: %w", err)
	}

	// Try parsing as array first (some endpoints return array directly)
	var markets []polymarketMarket
	if err := json.Unmarshal(body, &markets); err != nil {
		// Try parsing as object with data field
		var resp marketsResponse
		if err2 := json.Unmarshal(body, &resp); err2 != nil {
			return nil, fmt.Errorf("parse response: %w (original: %v)", err2, err)
		}
		markets = resp.Data
	}

	// Convert to our types
	result := make([]types.Market, 0, len(markets))
	for _, m := range markets {
		market := convertMarket(m)

		// Apply post-filter for liquidity and end date
		if filter.MinLiquidity > 0 && market.Liquidity < filter.MinLiquidity {
			continue
		}
		if filter.EndDateAfter != nil && market.EndDate.Before(*filter.EndDateAfter) {
			continue
		}

		result = append(result, market)
	}

	return result, nil
}

// GetMarket fetches a single market by condition ID.
func (c *Client) GetMarket(conditionID string) (*types.Market, error) {
	path := fmt.Sprintf("/markets/%s", conditionID)

	body, err := c.doPublicRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("get market: %w", err)
	}

	var m polymarketMarket
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	market := convertMarket(m)
	return &market, nil
}

func convertMarket(m polymarketMarket) types.Market {
	market := types.Market{
		ID:          m.ConditionID,
		Platform:    "polymarket",
		ConditionID: m.ConditionID,
		Title:       m.Question,
		Description: m.Description,
		Active:      m.Active,
		Closed:      m.Closed,
	}

	// Parse end date
	if m.EndDateISO != "" {
		if t, err := time.Parse(time.RFC3339, m.EndDateISO); err == nil {
			market.EndDate = t
		}
	}

	// Convert tokens
	market.Tokens = make([]types.Token, 0, len(m.Tokens))
	for _, t := range m.Tokens {
		token := types.Token{
			TokenID: t.TokenID,
			Outcome: t.Outcome,
			Price:   t.Price,
			Winner:  t.Winner,
		}
		market.Tokens = append(market.Tokens, token)

		// Set yes/no prices
		if t.Outcome == "Yes" {
			market.OutcomeYesPrice = t.Price
		} else if t.Outcome == "No" {
			market.OutcomeNoPrice = t.Price
		}
	}

	return market
}
