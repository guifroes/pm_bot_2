package polymarket

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestClient_GetBalanceForWallet_ReturnsUSDCBalance(t *testing.T) {
	// Skip if no wallet address is provided
	walletAddress := os.Getenv("POLYMARKET_WALLET_ADDRESS")
	if walletAddress == "" {
		t.Skip("POLYMARKET_WALLET_ADDRESS not set - skipping real balance test")
	}

	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    clobBaseURL,
	}

	balance, err := client.GetBalanceForWallet(walletAddress)
	if err != nil {
		t.Fatalf("GetBalanceForWallet: %v", err)
	}

	// Balance should be non-negative (can be 0)
	if balance.Amount < 0 {
		t.Errorf("balance amount should not be negative, got %f", balance.Amount)
	}

	// Currency should be USDC
	if balance.Currency != "USDC" {
		t.Errorf("expected currency 'USDC', got '%s'", balance.Currency)
	}

	// Platform should be polymarket
	if balance.Platform != "polymarket" {
		t.Errorf("expected platform 'polymarket', got '%s'", balance.Platform)
	}

	t.Logf("Balance: %.6f %s on %s", balance.Amount, balance.Currency, balance.Platform)
}

func TestClient_GetBalanceForWallet_WithKnownAddress(t *testing.T) {
	// Test with a known Polygon address to verify the API works
	// This is a public address with known USDC activity
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    clobBaseURL,
	}

	// Use a well-known Polymarket address that likely has some USDC
	// We use a public address for testing - no private keys involved
	testAddress := "0x0000000000000000000000000000000000000000"

	balance, err := client.GetBalanceForWallet(testAddress)
	if err != nil {
		t.Fatalf("GetBalanceForWallet: %v", err)
	}

	// Balance should be >= 0 (zero address likely has 0)
	if balance.Amount < 0 {
		t.Errorf("balance amount should not be negative, got %f", balance.Amount)
	}

	if balance.Currency != "USDC" {
		t.Errorf("expected currency 'USDC', got '%s'", balance.Currency)
	}

	t.Logf("Zero address balance: %.6f %s", balance.Amount, balance.Currency)
}

func TestClient_GetBalance_ImplementsPlatformInterface(t *testing.T) {
	// Skip if wallet address is not set
	walletAddress := os.Getenv("POLYMARKET_WALLET_ADDRESS")
	if walletAddress == "" {
		t.Skip("POLYMARKET_WALLET_ADDRESS not set - skipping real balance test")
	}

	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    clobBaseURL,
		creds: Credentials{
			WalletAddress: walletAddress,
		},
	}

	balance, err := client.GetBalance()
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}

	// Balance should be non-negative (can be 0)
	if balance < 0 {
		t.Errorf("balance should not be negative, got %f", balance)
	}

	t.Logf("Balance via Platform interface: %.6f", balance)
}
