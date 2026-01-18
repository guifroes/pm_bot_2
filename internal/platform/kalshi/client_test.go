package kalshi

import (
	"os"
	"testing"
)

func TestNewClient_RequiresCredentials(t *testing.T) {
	// Clear environment variables to test error case
	originalKey := os.Getenv("KALSHI_API_KEY")
	originalPrivate := os.Getenv("KALSHI_PRIVATE_KEY")
	originalPrivatePath := os.Getenv("KALSHI_PRIVATE_KEY_PATH")
	defer func() {
		os.Setenv("KALSHI_API_KEY", originalKey)
		os.Setenv("KALSHI_PRIVATE_KEY", originalPrivate)
		os.Setenv("KALSHI_PRIVATE_KEY_PATH", originalPrivatePath)
	}()

	os.Unsetenv("KALSHI_API_KEY")
	os.Unsetenv("KALSHI_PRIVATE_KEY")
	os.Unsetenv("KALSHI_PRIVATE_KEY_PATH")

	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when credentials are missing")
	}
}

func hasKalshiCredentials() bool {
	apiKey := os.Getenv("KALSHI_API_KEY")
	privateKey := os.Getenv("KALSHI_PRIVATE_KEY")
	privateKeyPath := os.Getenv("KALSHI_PRIVATE_KEY_PATH")
	return apiKey != "" && (privateKey != "" || privateKeyPath != "")
}

func TestClient_Ping(t *testing.T) {
	// Skip if no credentials
	if !hasKalshiCredentials() {
		t.Skip("KALSHI_API_KEY and KALSHI_PRIVATE_KEY (or KALSHI_PRIVATE_KEY_PATH) required for this test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test connectivity with a public endpoint
	err = client.Ping()
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestClient_AuthenticatedRequest(t *testing.T) {
	// This test verifies that authenticated requests don't return auth errors
	if !hasKalshiCredentials() {
		t.Skip("KALSHI_API_KEY and KALSHI_PRIVATE_KEY (or KALSHI_PRIVATE_KEY_PATH) required for this test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test authenticated endpoint - GetBalance uses auth
	balance, err := client.GetBalance()
	if err != nil {
		t.Fatalf("authenticated request failed: %v", err)
	}

	// Balance should be >= 0 (can be zero for new accounts)
	if balance.Available < 0 {
		t.Errorf("balance should be >= 0, got %f", balance.Available)
	}
}

func TestGenerateSignature(t *testing.T) {
	// Test signature generation with a valid test RSA private key (PKCS#8 format)
	testPrivateKeyPEM := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC2T+VSQfomVxPr
h0MppFAxA80XXkuTgJFNpKnlxH9tH38IDoGeZRz7vGJP/ycA5oRIGX1rLURZBQlx
eUhDITmVj01Hv468Q/ZKPdX1ng8xxdhD8vq5zbkMw3VTfGwXlEi9laIboznTDn7x
7rhZlZXifVT6mzCh+7EKqwtX9R6QM2FsrfrHdbdbJq9ZCQaJMmpjPHxNMUbUtj5K
uHWd5Pc74B+H+6gJKP1ocHsORFwNUta4zh85b9g2n4yEN4zdBYhm0uRrxFsLwBnD
aKaIQIP2o8V+VM1XstD+C6LdprIj2bwUQy9xncrOvR1lteu+G/9bgL0I1DAShA3W
ESWyh1K/AgMBAAECggEAAM1JsomYSfNqxOFhB9lTkxsVtTCbH7IrefsnK6AXPUNf
AL0AqcDlcAHCyb45kbHYRtfX/i7TdM2VZuqkDWVnv/hk3DoQiqlR3eEy8zHkYvNk
oXbQ8NhIJBeQiRU043FlZdidlJdoRj0eFY+s2PopVDcKg5UDKxVdZf3Qemz3MGIc
5+5S/szpe7ehanGTPUbvMcKdWh8YH/SrSpwYmgER2ma/q0ivCUr8meXWIweiuOxQ
psx/BGAiKld49AUYLGftdBkKMbOi+PgL4ARhucBfE3D17VGmsA+vVH4hgC8vamia
talF62t3x0B8MnIzdayq3sf79ISKtIPA1mXfE2ZpwQKBgQDbedqZqUX7UyG6mNTn
AAwQ3WNaxFmeXAziJTw6j7A9uTqZGXlfsSdkyJRbStWbK5Axz9LRvMAUVfWSg5qj
iM6Key2pSJSLsrp4+9CQ6urEvYe/AdKuOSCEy7H18OM4PR9acQlOficsNPDTC1s3
wAiYJMnjkABf2ygRvrFoNkeXFwKBgQDUpsfIFQhoWXFcDk1jkw9u7hli8/Ybdg/c
Md68n8kZaudqk5zghRl9jVrKczwtFcSPrEjJhBJknDCkQGSdKX926W5inpnHd/gX
IdSVQm3gx+DKNGqrsHm8SmR039cDf3UvUJeOoi/QDGuvPTo5bnOKg/ymHEOSLdWx
PbNy8VXqmQKBgDFn9+a5bVCLQT+BIgQyRYUSYUhQhSAZ9qh921YPfIwYg3Ftg54g
Ag80+/ilGvrITrh34SxnwhGR3Cs0Rv5jUKNp4TiHZzEfdczAWw4UY+8P/1vnLCce
Iwzh0djcdjn1wHYalg6+ZVEVRdUsbEdbilO9jFkW1I6/hgCgnc0o0urXAoGAKlm8
2AA4WG/Xv7mpd/dFz5XjwG1NylJM/lGARpib+E/uHq+fQqe/V93bAw7IIUKAjwyE
wn1nHFpu5Yddgl9NX2VF8qYbgjpGUnUOXVuJfobQIfUmeWMAG5vFPfGGZM/xiqbG
SEXMt+aBW7kZ624v3JpEquBeJLK0KERdhLrDnaECgYEAjtG4XJKRb1+K9ZNPa5ZX
4cbVi9Kcd6ZlflfWbT1xWPLOnGuf5rlaEEAgK85xkan6Ty0zfZ7XfGdZFpycUCzD
UvgwYD0JrnNMK0z3iZhLJkyIWOpS3FwL1WMPhMq/YKuHWQSpaNwiq0dI1amIVZ+4
0ZtF9mcGvPgtdWfOKiZ3kTg=
-----END PRIVATE KEY-----`

	timestamp := "1705600000000"
	method := "GET"
	path := "/trade-api/v2/portfolio/balance"

	sig, err := generateSignature(testPrivateKeyPEM, timestamp, method, path)
	if err != nil {
		t.Fatalf("failed to generate signature: %v", err)
	}

	// Signature should be non-empty base64 string
	if sig == "" {
		t.Fatal("signature should not be empty")
	}

	// Signature should be valid base64
	if len(sig) < 10 {
		t.Error("signature seems too short")
	}
}
