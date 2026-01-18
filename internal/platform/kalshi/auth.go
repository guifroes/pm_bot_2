package kalshi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strconv"
	"time"
)

// Credentials holds Kalshi API credentials.
type Credentials struct {
	APIKey     string
	PrivateKey string // PEM-encoded RSA private key
}

// generateSignature generates the RSA-PSS signature for Kalshi API requests.
// Message format: timestamp + method + path (without query parameters)
func generateSignature(privateKeyPEM, timestamp, method, path string) (string, error) {
	// Parse the PEM-encoded private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing private key")
	}

	var privateKey *rsa.PrivateKey
	var err error

	// Try parsing as PKCS#1 first, then PKCS#8
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS#8
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return "", fmt.Errorf("failed to parse private key: PKCS#1 error: %v, PKCS#8 error: %v", err, err2)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("private key is not RSA")
		}
	}

	// Create the message to sign: timestamp + method + path
	message := timestamp + method + path

	// Hash the message with SHA256
	hash := sha256.Sum256([]byte(message))

	// Sign with RSA-PSS
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash[:], &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	})
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Encode signature as base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

// getTimestampMS returns the current timestamp in milliseconds.
func getTimestampMS() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}
