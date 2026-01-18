package types

import "time"

// Price represents a price point for an asset.
type Price struct {
	Symbol    string
	Price     float64
	Timestamp time.Time
	Source    string
}
