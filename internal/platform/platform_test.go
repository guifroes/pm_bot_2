package platform

import (
	"testing"

	"prediction-bot/pkg/types"
)

// MockPlatform is a test implementation of the Platform interface
type MockPlatform struct {
	name      string
	markets   []types.Market
	positions []types.Position
	balance   float64
}

func (m *MockPlatform) Name() string {
	return m.name
}

func (m *MockPlatform) ListMarkets(filter types.MarketFilter) ([]types.Market, error) {
	return m.markets, nil
}

func (m *MockPlatform) GetOrderBook(tokenID string) (*types.OrderBook, error) {
	return &types.OrderBook{TokenID: tokenID}, nil
}

func (m *MockPlatform) GetBalance() (float64, error) {
	return m.balance, nil
}

func (m *MockPlatform) GetPositions() ([]types.Position, error) {
	return m.positions, nil
}

func TestPlatformInterface(t *testing.T) {
	// This test verifies the Platform interface is properly defined
	// and can be implemented
	var p Platform = &MockPlatform{
		name:    "test",
		balance: 100.0,
		markets: []types.Market{
			{ID: "market-1", Title: "Test Market"},
		},
		positions: []types.Position{
			{MarketTicker: "TICKER-1"},
		},
	}

	// Test Name
	if p.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", p.Name())
	}

	// Test ListMarkets
	markets, err := p.ListMarkets(types.MarketFilter{})
	if err != nil {
		t.Errorf("ListMarkets returned error: %v", err)
	}
	if len(markets) != 1 {
		t.Errorf("expected 1 market, got %d", len(markets))
	}

	// Test GetOrderBook
	ob, err := p.GetOrderBook("token-1")
	if err != nil {
		t.Errorf("GetOrderBook returned error: %v", err)
	}
	if ob.TokenID != "token-1" {
		t.Errorf("expected token ID 'token-1', got '%s'", ob.TokenID)
	}

	// Test GetBalance
	balance, err := p.GetBalance()
	if err != nil {
		t.Errorf("GetBalance returned error: %v", err)
	}
	if balance != 100.0 {
		t.Errorf("expected balance 100.0, got %f", balance)
	}

	// Test GetPositions
	positions, err := p.GetPositions()
	if err != nil {
		t.Errorf("GetPositions returned error: %v", err)
	}
	if len(positions) != 1 {
		t.Errorf("expected 1 position, got %d", len(positions))
	}
}
