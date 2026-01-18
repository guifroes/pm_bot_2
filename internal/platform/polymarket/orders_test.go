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
