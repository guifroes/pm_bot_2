package kalshi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"prediction-bot/pkg/types"
)

// KalshiMarket represents a market from the Kalshi API.
type KalshiMarket struct {
	Ticker           string  `json:"ticker"`
	EventTicker      string  `json:"event_ticker"`
	MarketType       string  `json:"market_type"`
	Title            string  `json:"title"`
	Subtitle         string  `json:"subtitle"`
	YesBid           int     `json:"yes_bid"`
	YesAsk           int     `json:"yes_ask"`
	NoBid            int     `json:"no_bid"`
	NoAsk            int     `json:"no_ask"`
	LastPrice        int     `json:"last_price"`
	PreviousYesBid   int     `json:"previous_yes_bid"`
	PreviousYesAsk   int     `json:"previous_yes_ask"`
	PreviousPrice    int     `json:"previous_price"`
	Volume           int     `json:"volume"`
	Volume24H        int     `json:"volume_24h"`
	Liquidity        int     `json:"liquidity"`
	OpenInterest     int     `json:"open_interest"`
	Result           string  `json:"result"`
	Status           string  `json:"status"`
	CloseTime        string  `json:"close_time"`
	ExpirationTime   string  `json:"expiration_time"`
	LatestExpiryTime int64   `json:"latest_expiry_time"`
	Category         string  `json:"category"`
	RiskLimitCents   int     `json:"risk_limit_cents"`
	StrikeType       string  `json:"strike_type"`
	FloorStrike      float64 `json:"floor_strike"`
	CapStrike        float64 `json:"cap_strike"`
}

// MarketsResponse represents the API response for listing markets.
type MarketsResponse struct {
	Markets []KalshiMarket `json:"markets"`
	Cursor  string         `json:"cursor"`
}

// ListMarkets returns a list of markets matching the filter criteria.
func (c *Client) ListMarkets(filter types.MarketFilter) ([]types.Market, error) {
	// Build query parameters
	params := make(map[string]string)

	if filter.IsActive != nil && *filter.IsActive {
		// API accepts "open" as filter value but returns "active" in response
		params["status"] = "open"
	}

	if filter.Limit > 0 {
		params["limit"] = strconv.Itoa(filter.Limit)
	}

	if filter.Offset > 0 {
		// Kalshi uses cursor-based pagination, but we can use offset for simplicity
		// Note: This may not work exactly like offset in Kalshi API
	}

	path := BuildURL("/markets", params)
	body, err := c.doPublicRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("list markets: %w", err)
	}

	var response MarketsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse markets response: %w", err)
	}

	// Convert Kalshi markets to common Market type
	markets := make([]types.Market, 0, len(response.Markets))
	for _, km := range response.Markets {
		market := convertKalshiMarket(km)
		markets = append(markets, market)
	}

	return markets, nil
}

// convertKalshiMarket converts a Kalshi-specific market to the common Market type.
func convertKalshiMarket(km KalshiMarket) types.Market {
	// Parse close time
	var endDate time.Time
	if km.CloseTime != "" {
		endDate, _ = time.Parse(time.RFC3339, km.CloseTime)
	}

	// Determine active status from Kalshi status field
	// Kalshi statuses: "active", "closed", "settled"
	isActive := km.Status == "active"
	isClosed := km.Status == "closed" || km.Status == "settled"

	// Convert prices from cents (0-100) to decimal (0.0-1.0)
	// YesBid is the highest bid price, YesAsk is the lowest ask price
	// For OutcomeYesPrice, use the mid-price or last price
	var yesPrice float64
	if km.YesBid > 0 && km.YesAsk > 0 {
		yesPrice = float64(km.YesBid+km.YesAsk) / 2.0 / 100.0
	} else if km.LastPrice > 0 {
		yesPrice = float64(km.LastPrice) / 100.0
	}

	noPrice := 1.0 - yesPrice

	return types.Market{
		ID:              km.Ticker,
		Platform:        "kalshi",
		ConditionID:     km.EventTicker,
		Title:           km.Title,
		Description:     km.Subtitle,
		EndDate:         endDate,
		Volume:          float64(km.Volume24H) / 100.0, // Convert cents to dollars
		Liquidity:       float64(km.Liquidity) / 100.0, // Convert cents to dollars
		Active:          isActive,
		Closed:          isClosed,
		OutcomeYesPrice: yesPrice,
		OutcomeNoPrice:  noPrice,
		Tokens:          nil, // Kalshi doesn't use tokens like Polymarket
	}
}
