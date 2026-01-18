package types

import "time"

// OrderSide represents the side of an order (buy or sell).
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents the type of an order.
type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

// TimeInForce specifies how long an order remains active.
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancelled
	TimeInForceIOC TimeInForce = "IOC" // Immediate Or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill Or Kill
)

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusOpen      OrderStatus = "open"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusSimulated OrderStatus = "simulated"
)

// Order represents an order to place on a prediction market.
type Order struct {
	MarketID    string
	TokenID     string
	Side        OrderSide
	Type        OrderType
	Price       float64
	Size        float64
	TimeInForce TimeInForce
}

// OrderResult represents the result of placing an order.
type OrderResult struct {
	OrderID   string
	MarketID  string
	TokenID   string
	Side      OrderSide
	Price     float64
	Size      float64
	Status    OrderStatus
	IsDryRun  bool
	CreatedAt time.Time
}
