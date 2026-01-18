# Data Sources

Provides real-time and historical price data for underlying assets.

## Supported Sources

### Binance (Crypto)
- **Use for**: BTC, ETH, and other cryptocurrencies
- **Endpoint**: WebSocket streams (real-time), REST API (historical)
- **Auth**: None required for public market data
- **Rate limits**: 1200 requests/minute (REST), no limit on WebSocket

```
# WebSocket for real-time
wss://stream.binance.com:9443/ws/btcusdt@ticker

# REST for historical
GET https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1h&limit=336
```

### Alpha Vantage (Stocks/Indices)
- **Use for**: SPY, QQQ, individual stocks
- **Endpoint**: REST API
- **Auth**: API key (free tier)
- **Rate limits**: 25 requests/day (free), 500/day (premium)

```
GET https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=SPY&apikey=XXX
```

## Data Interface

All sources implement common interface:

```go
type PriceSource interface {
    // Get current price
    GetPrice(symbol string) (Price, error)
    
    // Get historical prices for volatility calculation
    GetHistory(symbol string, hours int) ([]Price, error)
    
    // Subscribe to real-time updates
    Subscribe(symbol string, callback func(Price)) error
}

type Price struct {
    Symbol    string
    Price     float64
    Timestamp time.Time
    Volume    float64
}
```

## Symbol Mapping

Map prediction market references to data source symbols:

```
# Crypto (Binance uses pairs)
"BTC", "Bitcoin" → "BTCUSDT"
"ETH", "Ethereum" → "ETHUSDT"

# Stocks/Indices (Alpha Vantage uses tickers)
"S&P 500", "S&P500" → "SPY"
"NASDAQ", "Nasdaq 100" → "QQQ"
"Tesla" → "TSLA"
```

## Caching Strategy

- **Real-time prices**: Cache for 5 seconds
- **Historical data**: Cache for 1 hour
- **Symbol mappings**: Cache indefinitely

Reduces API calls, especially important for Alpha Vantage's strict limits.

## Failover

If primary source fails:

1. Crypto: Binance → CoinGecko (backup)
2. Stocks: Alpha Vantage → Yahoo Finance unofficial API (backup, less reliable)

Log warning when using fallback.

## Error Handling

```go
type DataError struct {
    Source    string  // "binance", "alphavantage"
    Symbol    string
    ErrorType string  // "not_found", "rate_limit", "network", "parse"
    Message   string
}
```

- **not_found**: Symbol doesn't exist → reject trade
- **rate_limit**: Wait and retry with backoff
- **network**: Retry 3x, then use cached data if available
- **parse**: Log error, skip this data point

## Configuration

```yaml
data_sources:
  binance:
    ws_url: "wss://stream.binance.com:9443/ws"
    rest_url: "https://api.binance.com/api/v3"
    
  alphavantage:
    url: "https://www.alphavantage.co/query"
    api_key: "${ALPHAVANTAGE_API_KEY}"
    
  cache:
    realtime_ttl_seconds: 5
    history_ttl_seconds: 3600
```

## Integration

- Used by: Volatility Analyzer (historical data)
- Used by: Dashboard (current prices)
- Symbols requested by: Market Scanner (via parsed market data)
