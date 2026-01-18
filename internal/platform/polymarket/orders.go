package polymarket

import (
	"fmt"
	"time"

	"prediction-bot/pkg/types"

	"github.com/google/uuid"
)

// PlaceOrder places an order on Polymarket.
// When dryRun is true, it returns a simulated result without actually placing the order.
func (c *Client) PlaceOrder(order types.Order, dryRun bool) (types.OrderResult, error) {
	// Validate order fields
	if err := validateOrder(order); err != nil {
		return types.OrderResult{}, err
	}

	if dryRun {
		return simulateOrder(order), nil
	}

	// TODO: Implement real order placement in Fatia 13.1
	return types.OrderResult{}, fmt.Errorf("live trading not yet implemented")
}

// validateOrder checks that all required fields are present and valid.
func validateOrder(order types.Order) error {
	if order.MarketID == "" {
		return fmt.Errorf("order validation: MarketID is required")
	}

	if order.TokenID == "" {
		return fmt.Errorf("order validation: TokenID is required")
	}

	if order.Size <= 0 {
		return fmt.Errorf("order validation: Size must be positive")
	}

	if order.Price < 0 {
		return fmt.Errorf("order validation: Price cannot be negative")
	}

	if order.Price > 1 {
		return fmt.Errorf("order validation: Price cannot exceed 1.0")
	}

	return nil
}

// simulateOrder creates a simulated order result for dry-run mode.
func simulateOrder(order types.Order) types.OrderResult {
	return types.OrderResult{
		OrderID:   fmt.Sprintf("dryrun-%s", uuid.New().String()),
		MarketID:  order.MarketID,
		TokenID:   order.TokenID,
		Side:      order.Side,
		Price:     order.Price,
		Size:      order.Size,
		Status:    types.OrderStatusSimulated,
		IsDryRun:  true,
		CreatedAt: time.Now(),
	}
}
