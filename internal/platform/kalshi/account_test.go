package kalshi

import (
	"os"
	"testing"
)

func TestClient_GetBalanceDetails(t *testing.T) {
	if !hasKalshiCredentials() {
		t.Skip("KALSHI_API_KEY and KALSHI_PRIVATE_KEY (or KALSHI_PRIVATE_KEY_PATH) required for this test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	balance, err := client.GetBalanceDetails()
	if err != nil {
		t.Fatalf("GetBalanceDetails failed: %v", err)
	}

	// Balance should be >= 0 (can be zero for new accounts)
	if balance.Available < 0 {
		t.Errorf("available balance should be >= 0, got %f", balance.Available)
	}

	if balance.Reserved < 0 {
		t.Errorf("reserved balance should be >= 0, got %f", balance.Reserved)
	}

	t.Logf("Balance: Available=$%.2f, Reserved=$%.2f, BonusCash=$%.2f",
		balance.Available, balance.Reserved, balance.BonusCashBalance)
}

func TestClient_GetBalance_ImplementsPlatformInterface(t *testing.T) {
	if !hasKalshiCredentials() {
		t.Skip("KALSHI_API_KEY and KALSHI_PRIVATE_KEY (or KALSHI_PRIVATE_KEY_PATH) required for this test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	balance, err := client.GetBalance()
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}

	// Balance should be >= 0 (can be zero for new accounts)
	if balance < 0 {
		t.Errorf("balance should be >= 0, got %f", balance)
	}

	t.Logf("Balance via Platform interface: $%.2f", balance)
}

func TestClient_GetPositions(t *testing.T) {
	if !hasKalshiCredentials() {
		t.Skip("KALSHI_API_KEY and KALSHI_PRIVATE_KEY (or KALSHI_PRIVATE_KEY_PATH) required for this test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	positions, err := client.GetPositions()
	if err != nil {
		t.Fatalf("GetPositions failed: %v", err)
	}

	// Positions can be empty for accounts without open positions
	t.Logf("Found %d positions", len(positions))

	for _, pos := range positions {
		t.Logf("Position: Market=%s, Qty=%d, Exposure=$%.2f, UnrealizedPnL=$%.2f",
			pos.MarketTicker, pos.Quantity, pos.MarketExposure, pos.UnrealizedPnL)
	}
}

func TestClient_GetPositions_NoCredentials(t *testing.T) {
	// Clear environment variables to test error case
	originalKey := os.Getenv("KALSHI_API_KEY")
	originalPrivate := os.Getenv("KALSHI_PRIVATE_KEY")
	originalPrivatePath := os.Getenv("KALSHI_PRIVATE_KEY_PATH")
	defer func() {
		os.Setenv("KALSHI_API_KEY", originalKey)
		if originalPrivate != "" {
			os.Setenv("KALSHI_PRIVATE_KEY", originalPrivate)
		}
		if originalPrivatePath != "" {
			os.Setenv("KALSHI_PRIVATE_KEY_PATH", originalPrivatePath)
		}
	}()

	os.Unsetenv("KALSHI_API_KEY")
	os.Unsetenv("KALSHI_PRIVATE_KEY")
	os.Unsetenv("KALSHI_PRIVATE_KEY_PATH")

	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when credentials are missing")
	}
}
