package types

import "time"

// Position represents an open position in a market.
type Position struct {
	Platform         string
	MarketTicker     string
	Quantity         int       // Number of contracts held (positive = Yes, negative = No)
	AveragePrice     float64   // Average entry price (0.0 to 1.0)
	MarketExposure   float64   // Total market exposure in dollars
	RealizedPnL      float64   // Realized profit/loss
	UnrealizedPnL    float64   // Unrealized profit/loss
	TotalTraded      int       // Total contracts traded
	FeesPaid         float64   // Total fees paid
	RestingOrdersQty int       // Quantity in resting orders
	Timestamp        time.Time // When the position was last updated
}
