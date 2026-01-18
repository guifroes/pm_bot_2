package types

import "time"

// Balance represents an account balance on a platform.
type Balance struct {
	Platform  string
	Currency  string
	Amount    float64
	Timestamp time.Time
}
