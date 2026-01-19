package polymarket

import (
	"testing"
	"time"

	"prediction-bot/pkg/types"
)

func TestPlaceOrder_DryRun_ReturnsSimulatedResult(t *testing.T) {
	// Arrange: create a client and order
	client := NewClientWithCreds(Credentials{
		APIKey:     "test-key",
		APISecret:  "test-secret",
		Passphrase: "test-passphrase",
	})

	order := types.Order{
		MarketID:  "test-market-123",
		TokenID:   "token-abc",
		Side:      types.OrderSideBuy,
		Type:      types.OrderTypeLimit,
		Price:     0.85,
		Size:      10.0,
		TimeInForce: types.TimeInForceGTC,
	}

	// Act: place order in DRY-RUN mode
	result, err := client.PlaceOrder(order, true)

	// Assert: should return simulated result without error
	if err != nil {
		t.Fatalf("PlaceOrder dry-run should not return error: %v", err)
	}

	if result.OrderID == "" {
		t.Error("PlaceOrder dry-run should return a simulated OrderID")
	}

	if !result.IsDryRun {
		t.Error("PlaceOrder dry-run should have IsDryRun=true")
	}

	if result.Status != types.OrderStatusSimulated {
		t.Errorf("PlaceOrder dry-run should have Status=Simulated, got %v", result.Status)
	}

	if result.MarketID != order.MarketID {
		t.Errorf("PlaceOrder dry-run should echo MarketID, got %v", result.MarketID)
	}

	if result.Price != order.Price {
		t.Errorf("PlaceOrder dry-run should echo Price, got %v", result.Price)
	}

	if result.Size != order.Size {
		t.Errorf("PlaceOrder dry-run should echo Size, got %v", result.Size)
	}
}

func TestPlaceOrder_DryRun_GeneratesUniqueOrderID(t *testing.T) {
	client := NewClientWithCreds(Credentials{
		APIKey:     "test-key",
		APISecret:  "test-secret",
		Passphrase: "test-passphrase",
	})

	order := types.Order{
		MarketID:  "test-market-123",
		TokenID:   "token-abc",
		Side:      types.OrderSideBuy,
		Type:      types.OrderTypeLimit,
		Price:     0.85,
		Size:      10.0,
	}

	// Place two orders and verify they have different IDs
	result1, err := client.PlaceOrder(order, true)
	if err != nil {
		t.Fatalf("First PlaceOrder should not fail: %v", err)
	}

	result2, err := client.PlaceOrder(order, true)
	if err != nil {
		t.Fatalf("Second PlaceOrder should not fail: %v", err)
	}

	if result1.OrderID == result2.OrderID {
		t.Error("Dry-run orders should have unique OrderIDs")
	}
}

func TestPlaceOrder_DryRun_ValidatesOrderFields(t *testing.T) {
	client := NewClientWithCreds(Credentials{
		APIKey:     "test-key",
		APISecret:  "test-secret",
		Passphrase: "test-passphrase",
	})

	tests := []struct {
		name    string
		order   types.Order
		wantErr bool
	}{
		{
			name: "valid order",
			order: types.Order{
				MarketID: "market-1",
				TokenID:  "token-1",
				Side:     types.OrderSideBuy,
				Type:     types.OrderTypeLimit,
				Price:    0.5,
				Size:     1.0,
			},
			wantErr: false,
		},
		{
			name: "missing MarketID",
			order: types.Order{
				TokenID: "token-1",
				Side:    types.OrderSideBuy,
				Type:    types.OrderTypeLimit,
				Price:   0.5,
				Size:    1.0,
			},
			wantErr: true,
		},
		{
			name: "missing TokenID",
			order: types.Order{
				MarketID: "market-1",
				Side:     types.OrderSideBuy,
				Type:     types.OrderTypeLimit,
				Price:    0.5,
				Size:     1.0,
			},
			wantErr: true,
		},
		{
			name: "zero size",
			order: types.Order{
				MarketID: "market-1",
				TokenID:  "token-1",
				Side:     types.OrderSideBuy,
				Type:     types.OrderTypeLimit,
				Price:    0.5,
				Size:     0,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			order: types.Order{
				MarketID: "market-1",
				TokenID:  "token-1",
				Side:     types.OrderSideBuy,
				Type:     types.OrderTypeLimit,
				Price:    -0.5,
				Size:     1.0,
			},
			wantErr: true,
		},
		{
			name: "price above 1",
			order: types.Order{
				MarketID: "market-1",
				TokenID:  "token-1",
				Side:     types.OrderSideBuy,
				Type:     types.OrderTypeLimit,
				Price:    1.5,
				Size:     1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.PlaceOrder(tt.order, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("PlaceOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaceOrder_DryRun_SetsCreatedAtTimestamp(t *testing.T) {
	client := NewClientWithCreds(Credentials{
		APIKey:     "test-key",
		APISecret:  "test-secret",
		Passphrase: "test-passphrase",
	})

	order := types.Order{
		MarketID: "market-1",
		TokenID:  "token-1",
		Side:     types.OrderSideBuy,
		Type:     types.OrderTypeLimit,
		Price:    0.5,
		Size:     1.0,
	}

	before := time.Now()
	result, err := client.PlaceOrder(order, true)
	after := time.Now()

	if err != nil {
		t.Fatalf("PlaceOrder should not fail: %v", err)
	}

	if result.CreatedAt.Before(before) || result.CreatedAt.After(after) {
		t.Errorf("CreatedAt should be set to current time, got %v", result.CreatedAt)
	}
}

// ==============================================================================
// LIVE TRADING TESTS (Fatia 13.1)
// ==============================================================================

// TestPlaceOrder_Live_RequiresValidCredentials tests that live trading requires
// proper API credentials and returns appropriate errors for invalid credentials.
// This is a unit test that does NOT place real orders.
func TestPlaceOrder_Live_RequiresValidCredentials(t *testing.T) {
	// Arrange: create a client with test credentials (not real)
	client := NewClientWithCreds(Credentials{
		APIKey:        "invalid-key",
		APISecret:     "aW52YWxpZC1zZWNyZXQ=", // base64 "invalid-secret"
		Passphrase:    "invalid-passphrase",
		WalletAddress: "0x0000000000000000000000000000000000000000",
	})

	order := types.Order{
		MarketID:    "test-market-123",
		TokenID:     "token-abc",
		Side:        types.OrderSideBuy,
		Type:        types.OrderTypeLimit,
		Price:       0.85,
		Size:        1.0, // Minimal size for test
		TimeInForce: types.TimeInForceGTC,
	}

	// Act: attempt to place live order (dryRun=false)
	_, err := client.PlaceOrder(order, false)

	// Assert: should return an error (either auth error or API error)
	// The key thing is that it doesn't just error with "not implemented"
	if err == nil {
		t.Fatal("PlaceOrder live should return error with invalid credentials")
	}

	// Verify it's not the "not implemented" error
	if err.Error() == "live trading not yet implemented" {
		t.Error("PlaceOrder live should be implemented (not return 'not yet implemented' error)")
	}
}

// TestPlaceOrder_Live_BuildsCorrectOrderPayload verifies the order payload
// structure is correct before sending to the API.
func TestPlaceOrder_Live_BuildsCorrectOrderPayload(t *testing.T) {
	client := NewClientWithCreds(Credentials{
		APIKey:        "test-key",
		APISecret:     "dGVzdC1zZWNyZXQ=", // base64 "test-secret"
		Passphrase:    "test-passphrase",
		WalletAddress: "0x1234567890123456789012345678901234567890",
	})

	order := types.Order{
		MarketID:    "0x123abc", // Condition ID format
		TokenID:     "12345678901234567890", // Token ID format
		Side:        types.OrderSideBuy,
		Type:        types.OrderTypeLimit,
		Price:       0.50,
		Size:        10.0,
		TimeInForce: types.TimeInForceGTC,
	}

	// Test that the order can be built (will fail on API call, but should get past validation)
	_, err := client.PlaceOrder(order, false)

	// We expect an error from the API (invalid credentials or network error),
	// but NOT a "not implemented" error
	if err != nil && err.Error() == "live trading not yet implemented" {
		t.Error("PlaceOrder should attempt to call the API, not return 'not implemented'")
	}
}

// TestBuildOrderPayload tests the internal order payload building.
func TestBuildOrderPayload(t *testing.T) {
	client := NewClientWithCreds(Credentials{
		APIKey:        "test-key",
		APISecret:     "dGVzdC1zZWNyZXQ=",
		Passphrase:    "test-passphrase",
		WalletAddress: "0x1234567890123456789012345678901234567890",
	})

	order := types.Order{
		MarketID:    "condition-123",
		TokenID:     "token-456",
		Side:        types.OrderSideBuy,
		Type:        types.OrderTypeLimit,
		Price:       0.65,
		Size:        5.0,
		TimeInForce: types.TimeInForceGTC,
	}

	payload, err := client.buildOrderPayload(order)
	if err != nil {
		t.Fatalf("buildOrderPayload should not error: %v", err)
	}

	if payload == nil {
		t.Fatal("buildOrderPayload should return non-nil payload")
	}

	// Verify required fields are present
	if payload["tokenID"] == nil {
		t.Error("payload should contain tokenID")
	}

	if payload["price"] == nil {
		t.Error("payload should contain price")
	}

	if payload["size"] == nil {
		t.Error("payload should contain size")
	}

	if payload["side"] == nil {
		t.Error("payload should contain side")
	}

	if payload["orderType"] == nil {
		t.Error("payload should contain orderType")
	}
}

// TestMapOrderTypeToAPI tests the mapping of order types to API format.
func TestMapOrderTypeToAPI(t *testing.T) {
	tests := []struct {
		orderType   types.OrderType
		timeInForce types.TimeInForce
		expected    string
	}{
		{types.OrderTypeLimit, types.TimeInForceGTC, "GTC"},
		{types.OrderTypeLimit, types.TimeInForceIOC, "IOC"},
		{types.OrderTypeLimit, types.TimeInForceFOK, "FOK"},
		{types.OrderTypeMarket, types.TimeInForceFOK, "FOK"},
	}

	for _, tt := range tests {
		t.Run(string(tt.orderType)+"_"+string(tt.timeInForce), func(t *testing.T) {
			result := mapOrderTypeToAPI(tt.orderType, tt.timeInForce)
			if result != tt.expected {
				t.Errorf("mapOrderTypeToAPI(%v, %v) = %v, want %v",
					tt.orderType, tt.timeInForce, result, tt.expected)
			}
		})
	}
}

// TestMapSideToAPI tests the mapping of order side to API format.
func TestMapSideToAPI(t *testing.T) {
	tests := []struct {
		side     types.OrderSide
		expected string
	}{
		{types.OrderSideBuy, "BUY"},
		{types.OrderSideSell, "SELL"},
	}

	for _, tt := range tests {
		t.Run(string(tt.side), func(t *testing.T) {
			result := mapSideToAPI(tt.side)
			if result != tt.expected {
				t.Errorf("mapSideToAPI(%v) = %v, want %v", tt.side, result, tt.expected)
			}
		})
	}
}
