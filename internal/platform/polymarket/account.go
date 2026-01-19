package polymarket

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"prediction-bot/pkg/types"
)

const (
	// Polygon RPC endpoint (public)
	polygonRPC = "https://polygon-rpc.com"

	// USDC contract address on Polygon
	usdcContractAddress = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"

	// ERC20 balanceOf function selector: keccak256("balanceOf(address)")[:4]
	balanceOfSelector = "70a08231"

	// USDC has 6 decimals on Polygon
	usdcDecimals = 6
)

// jsonRPCRequest represents a JSON-RPC request.
type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// jsonRPCResponse represents a JSON-RPC response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// jsonRPCError represents a JSON-RPC error.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetBalanceForWallet retrieves the USDC balance for a wallet address on Polygon.
func (c *Client) GetBalanceForWallet(walletAddress string) (types.Balance, error) {
	// Normalize address
	address := strings.ToLower(strings.TrimPrefix(walletAddress, "0x"))
	if len(address) != 40 {
		return types.Balance{}, fmt.Errorf("invalid wallet address: %s", walletAddress)
	}

	// Construct the balanceOf call data
	// Function selector (4 bytes) + address padded to 32 bytes
	callData := balanceOfSelector + strings.Repeat("0", 24) + address

	// Create eth_call request
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_call",
		Params: []interface{}{
			map[string]string{
				"to":   usdcContractAddress,
				"data": "0x" + callData,
			},
			"latest",
		},
		ID: 1,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return types.Balance{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", polygonRPC, bytes.NewReader(reqBody))
	if err != nil {
		return types.Balance{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return types.Balance{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.Balance{}, fmt.Errorf("read response: %w", err)
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return types.Balance{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return types.Balance{}, fmt.Errorf("rpc error: %s", rpcResp.Error.Message)
	}

	// Parse the result (hex string representing uint256)
	var resultHex string
	if err := json.Unmarshal(rpcResp.Result, &resultHex); err != nil {
		return types.Balance{}, fmt.Errorf("unmarshal result: %w", err)
	}

	amount, err := parseUSDCBalance(resultHex)
	if err != nil {
		return types.Balance{}, fmt.Errorf("parse balance: %w", err)
	}

	return types.Balance{
		Platform:  "polymarket",
		Currency:  "USDC",
		Amount:    amount,
		Timestamp: time.Now(),
	}, nil
}

// GetBalance implements platform.Platform interface.
// Returns the USDC balance for the configured wallet address.
func (c *Client) GetBalance() (float64, error) {
	if c.creds.WalletAddress == "" {
		return 0, fmt.Errorf("wallet address not configured (set POLYMARKET_WALLET_ADDRESS)")
	}

	balance, err := c.GetBalanceForWallet(c.creds.WalletAddress)
	if err != nil {
		return 0, err
	}

	return balance.Amount, nil
}

// GetPositions implements platform.Platform interface.
// Returns current positions (placeholder - Polymarket positions require on-chain queries).
func (c *Client) GetPositions() ([]types.Position, error) {
	// Polymarket positions are stored on-chain and require subgraph queries.
	// For now, return empty list. Full implementation requires GraphQL to Polymarket subgraph.
	return []types.Position{}, nil
}

// parseUSDCBalance converts a hex string to a USDC amount (6 decimals).
func parseUSDCBalance(hexStr string) (float64, error) {
	// Remove 0x prefix
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Handle empty or zero result
	if hexStr == "" || hexStr == "0" {
		return 0, nil
	}

	// Decode hex to bytes
	hexBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("decode hex: %w", err)
	}

	// Convert to big.Int
	balance := new(big.Int).SetBytes(hexBytes)

	// Convert to float64 with 6 decimals (USDC)
	divisor := new(big.Float).SetInt(big.NewInt(1e6))
	balanceFloat := new(big.Float).SetInt(balance)
	result, _ := new(big.Float).Quo(balanceFloat, divisor).Float64()

	return result, nil
}
