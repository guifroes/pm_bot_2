package types

import "time"

// Market represents a prediction market.
type Market struct {
	ID              string
	Platform        string
	ConditionID     string
	Title           string
	Description     string
	EndDate         time.Time
	Volume          float64
	Liquidity       float64
	Active          bool
	Closed          bool
	OutcomeYesPrice float64
	OutcomeNoPrice  float64
	Tokens          []Token
}

// Token represents a market outcome token.
type Token struct {
	TokenID  string
	Outcome  string
	Price    float64
	Winner   bool
}

// MarketFilter contains filter options for listing markets.
type MarketFilter struct {
	IsActive     *bool
	Closed       *bool
	EndDateAfter *time.Time
	MinLiquidity float64
	Limit        int
	Offset       int
}
