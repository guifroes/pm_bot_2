# Platform Integration

Handles all communication with prediction market platforms (Polymarket, Kalshi).

## Supported Platforms

### Polymarket
- **Type**: Decentralized (Polygon blockchain)
- **API**: REST + on-chain transactions
- **Auth**: Wallet signature (private key)
- **Settlement**: USDC on Polygon

### Kalshi
- **Type**: Centralized (CFTC-regulated)
- **API**: REST
- **Auth**: API key + secret
- **Settlement**: USD (bank transfer or card)

## Common Interface

```go
type Platform interface {
    // Discovery
    ListMarkets(filter MarketFilter) ([]Market, error)
    GetMarket(marketID string) (Market, error)
    GetOrderBook(marketID string) (OrderBook, error)
    
    // Trading
    PlaceOrder(order Order) (OrderResult, error)
    CancelOrder(orderID string) error
    GetPosition(marketID string) (Position, error)
    GetPositions() ([]Position, error)
    
    // Account
    GetBalance() (Balance, error)
}
```

## Data Types

```go
type Market struct {
    ID            string
    Title         string
    Description   string
    Category      string
    CloseTime     time.Time
    IsResolved    bool
    Resolution    string    // "yes", "no", "invalid", ""
    YesPrice      float64   // 0.0 to 1.0
    NoPrice       float64
    Volume24h     float64
    Liquidity     float64
}

type OrderBook struct {
    MarketID  string
    YesBids   []Level  // people wanting to buy Yes
    YesAsks   []Level  // people wanting to sell Yes
    NoBids    []Level
    NoAsks    []Level
}

type Level struct {
    Price    float64
    Quantity float64
}

type Order struct {
    MarketID  string
    Side      string  // "yes" or "no"
    Type      string  // "market" or "limit"
    Quantity  float64 // number of shares
    Price     float64 // only for limit orders
}

type OrderResult struct {
    OrderID     string
    Status      string  // "filled", "partial", "pending", "rejected"
    FilledQty   float64
    AvgPrice    float64
    Fees        float64
}
```

## Market Filter

```go
type MarketFilter struct {
    Categories     []string  // e.g., ["crypto", "finance"]
    MinProbability float64   // minimum yes or no price
    MaxCloseTime   time.Time // must close before this
    MinLiquidity   float64   // minimum available liquidity
    IsActive       bool      // only non-resolved markets
}
```

## Polymarket Specifics

```yaml
polymarket:
  api_url: "https://clob.polymarket.com"
  chain_id: 137  # Polygon mainnet
  
  # Wallet config
  private_key: "${POLYMARKET_PRIVATE_KEY}"
  
  # Contract addresses
  ctf_exchange: "0x..."
  usdc_token: "0x..."
```

**Order Flow**:
1. Build order parameters
2. Sign with wallet
3. Submit to CLOB API
4. Wait for on-chain confirmation

**Fees**: ~1% on winnings (no fee if lose)

## Kalshi Specifics

```yaml
kalshi:
  api_url: "https://api.elections.kalshi.com/trade-api/v2"
  
  # Auth
  api_key: "${KALSHI_API_KEY}"
  api_secret: "${KALSHI_API_SECRET}"
```

**Order Flow**:
1. Authenticate request with HMAC signature
2. Submit order to REST API
3. Receive immediate confirmation

**Fees**: Variable, typically 5-10% of profit

## Rate Limiting

| Platform   | Limit              | Strategy                    |
|------------|--------------------|-----------------------------|
| Polymarket | 100 req/min        | Token bucket, 1.5/sec       |
| Kalshi     | 30 req/min         | Token bucket, 0.5/sec       |

Implement per-platform rate limiters.

## Error Handling

```go
type PlatformError struct {
    Platform  string
    Code      string  // "auth", "rate_limit", "insufficient_funds", "market_closed", "network"
    Message   string
    Retryable bool
}
```

Retry logic:
- `rate_limit`: Wait for reset, then retry
- `network`: Retry 3x with exponential backoff
- `insufficient_funds`: Do not retry, alert user
- `market_closed`: Do not retry, skip market
- `auth`: Do not retry, halt bot

## Integration

- Called by: Market Scanner (list markets)
- Called by: Position Manager (place orders, get positions)
- Config from: Environment variables
- Logs to: Persistence (all API calls for debugging)
