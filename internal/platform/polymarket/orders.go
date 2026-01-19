package polymarket

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"prediction-bot/pkg/types"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// OrderResponse represents the API response from order placement.
type OrderResponse struct {
	Success     bool     `json:"success"`
	ErrorMsg    string   `json:"errorMsg,omitempty"`
	OrderID     string   `json:"orderId,omitempty"`
	OrderHashes []string `json:"orderHashes,omitempty"`
}

// PlaceOrder places an order on Polymarket.
// When dryRun is true, it returns a simulated result without actually placing the order.
// When dryRun is false, it submits the order to the CLOB API for real execution.
func (c *Client) PlaceOrder(order types.Order, dryRun bool) (types.OrderResult, error) {
	// Validate order fields
	if err := validateOrder(order); err != nil {
		return types.OrderResult{}, err
	}

	if dryRun {
		return simulateOrder(order), nil
	}

	// LIVE TRADING: Submit order to Polymarket CLOB API
	log.Warn().
		Str("market_id", order.MarketID).
		Str("token_id", order.TokenID).
		Str("side", string(order.Side)).
		Float64("price", order.Price).
		Float64("size", order.Size).
		Msg("⚠️ PLACING LIVE ORDER ON POLYMARKET")

	// Build the order payload
	payload, err := c.buildOrderPayload(order)
	if err != nil {
		return types.OrderResult{}, fmt.Errorf("build order payload: %w", err)
	}

	// Log the payload for audit trail
	payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
	log.Debug().
		RawJSON("payload", payloadJSON).
		Msg("Order payload")

	// Serialize payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return types.OrderResult{}, fmt.Errorf("marshal order payload: %w", err)
	}

	// Submit to CLOB API
	respBody, err := c.doRequest("POST", "/order", body)
	if err != nil {
		log.Error().
			Err(err).
			Str("market_id", order.MarketID).
			Msg("Failed to place order")
		return types.OrderResult{}, fmt.Errorf("place order: %w", err)
	}

	// Parse response
	var resp OrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return types.OrderResult{}, fmt.Errorf("parse order response: %w", err)
	}

	// Check for API-level error
	if !resp.Success {
		log.Error().
			Str("error_msg", resp.ErrorMsg).
			Str("market_id", order.MarketID).
			Msg("Order placement failed")
		return types.OrderResult{}, fmt.Errorf("order rejected: %s", resp.ErrorMsg)
	}

	log.Info().
		Str("order_id", resp.OrderID).
		Str("market_id", order.MarketID).
		Str("side", string(order.Side)).
		Float64("price", order.Price).
		Float64("size", order.Size).
		Msg("✅ Order placed successfully")

	return types.OrderResult{
		OrderID:   resp.OrderID,
		MarketID:  order.MarketID,
		TokenID:   order.TokenID,
		Side:      order.Side,
		Price:     order.Price,
		Size:      order.Size,
		Status:    types.OrderStatusPending,
		IsDryRun:  false,
		CreatedAt: time.Now(),
	}, nil
}

// buildOrderPayload constructs the order payload for the CLOB API.
// Based on Polymarket CLOB API documentation:
// https://docs.polymarket.com/developers/CLOB/orders/create-order
func (c *Client) buildOrderPayload(order types.Order) (map[string]interface{}, error) {
	// Calculate maker and taker amounts based on side and price
	// For a BUY order: makerAmount = size * price (USDC), takerAmount = size (shares)
	// For a SELL order: makerAmount = size (shares), takerAmount = size * price (USDC)
	//
	// Amounts in the Polymarket API are in the smallest unit:
	// - USDC: 6 decimals (1 USDC = 1,000,000 units)
	// - Conditional tokens: 6 decimals
	const decimals = 1e6

	var makerAmount, takerAmount string
	if order.Side == types.OrderSideBuy {
		// Buying shares: pay USDC, receive shares
		makerAmount = strconv.FormatUint(uint64(math.Round(order.Size*order.Price*decimals)), 10)
		takerAmount = strconv.FormatUint(uint64(math.Round(order.Size*decimals)), 10)
	} else {
		// Selling shares: pay shares, receive USDC
		makerAmount = strconv.FormatUint(uint64(math.Round(order.Size*decimals)), 10)
		takerAmount = strconv.FormatUint(uint64(math.Round(order.Size*order.Price*decimals)), 10)
	}

	// Map order type and time-in-force to API format
	orderType := mapOrderTypeToAPI(order.Type, order.TimeInForce)

	// Build payload
	payload := map[string]interface{}{
		"tokenID":   order.TokenID,
		"price":     formatPrice(order.Price),
		"size":      order.Size,
		"side":      mapSideToAPI(order.Side),
		"orderType": orderType,
		// The following fields would be needed for full implementation
		// but require EIP-712 signing which is handled by the API credentials
		"makerAmount": makerAmount,
		"takerAmount": takerAmount,
	}

	return payload, nil
}

// formatPrice formats a price (0.0-1.0) to the API format (2 decimal string).
func formatPrice(price float64) string {
	return strconv.FormatFloat(price, 'f', 2, 64)
}

// mapOrderTypeToAPI maps order type and time-in-force to Polymarket API format.
func mapOrderTypeToAPI(orderType types.OrderType, tif types.TimeInForce) string {
	// Polymarket uses order types: FOK, FAK, GTC, GTD
	switch tif {
	case types.TimeInForceFOK:
		return "FOK"
	case types.TimeInForceIOC:
		return "IOC" // Polymarket calls this FAK (Fill-And-Kill) but accepts IOC
	case types.TimeInForceGTC:
		return "GTC"
	default:
		return "GTC" // Default to Good-Till-Cancelled
	}
}

// mapSideToAPI maps order side to Polymarket API format.
func mapSideToAPI(side types.OrderSide) string {
	switch side {
	case types.OrderSideBuy:
		return "BUY"
	case types.OrderSideSell:
		return "SELL"
	default:
		return "BUY"
	}
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
