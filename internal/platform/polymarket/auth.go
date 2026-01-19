package polymarket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

// Credentials holds Polymarket API credentials.
type Credentials struct {
	APIKey        string
	APISecret     string
	Passphrase    string
	WalletAddress string
}

// generateL2Signature generates the HMAC signature for L2 API requests.
// Based on Polymarket CLOB API documentation.
func generateL2Signature(creds Credentials, timestamp, method, path string, body []byte) (string, error) {
	// Decode the base64 secret
	secretBytes, err := base64.StdEncoding.DecodeString(creds.APISecret)
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	// Create the message to sign: timestamp + method + path + body
	var bodyStr string
	if body != nil && len(body) > 0 {
		bodyStr = string(body)
	}
	message := timestamp + method + path + bodyStr

	// Create HMAC-SHA256 signature
	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature, nil
}

// getTimestamp returns the current timestamp in the format required by Polymarket.
func getTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
