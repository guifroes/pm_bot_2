# AGENTS.md

## Project Overview

Autonomous trading bot for prediction markets (Polymarket, Kalshi) using a tail-end trading strategy. The bot identifies high-probability markets (>80%) with short time to resolution (<48h) and validates trades using underlying asset volatility analysis.

## Tech Stack

- **Language**: Go (primary), Python (auxiliary for GARCH if needed)
- **Database**: SQLite
- **Data Sources**: Binance WebSocket (crypto), Alpha Vantage (stocks)
- **Platforms**: Polymarket, Kalshi
- **UI**: Terminal/CLI using bubbletea or tcell

## Project Structure

```
prediction-market-bot/
├── cmd/
│   └── bot/
│       └── main.go           # Entry point
├── internal/
│   ├── scanner/              # Market scanning
│   ├── volatility/           # Volatility analysis
│   ├── position/             # Position management
│   ├── sizing/               # Kelly criterion
│   ├── platform/             # Platform integrations
│   │   ├── polymarket/
│   │   └── kalshi/
│   ├── datasource/           # Price data sources
│   │   ├── binance/
│   │   └── alphavantage/
│   ├── learning/             # Parameter learning
│   ├── persistence/          # SQLite storage
│   └── dashboard/            # Terminal UI
├── pkg/
│   └── types/                # Shared types
├── migrations/               # SQL migrations
├── config/
│   └── config.yaml           # Configuration
├── specs/                    # Specification documents
├── scripts/                  # Helper scripts
├── go.mod
├── go.sum
└── README.md
```

## Development Principles

1. **Test-Driven Development (TDD)**: Write tests first
2. **Small, focused functions**: Each function does one thing
3. **Explicit error handling**: No silent failures
4. **Dependency injection**: For testability
5. **No global state**: Pass dependencies explicitly

## Code Style

- Follow standard Go conventions (gofmt, golint)
- Use meaningful variable names
- Keep functions under 50 lines
- Document exported functions and types
- Use interfaces for external dependencies

## Backpressure Gates

Before marking any task complete:

1. `go build ./...` must succeed
2. `go test ./...` must pass
3. `golint ./...` must have no warnings
4. `go vet ./...` must pass

## Configuration

Environment variables:
- `POLYMARKET_PRIVATE_KEY`: Wallet private key for Polymarket
- `KALSHI_API_KEY`: Kalshi API key
- `KALSHI_API_SECRET`: Kalshi API secret
- `ALPHAVANTAGE_API_KEY`: Alpha Vantage API key

Config file (`config/config.yaml`):
```yaml
bankroll:
  polymarket: 50.0
  kalshi: 50.0

scan:
  interval_seconds: 10
  
parameters:
  probability_threshold: 0.80
  volatility_safety_margin: 1.5
  stop_loss_percent: 0.15
  kelly_fraction: 0.25

database:
  path: "~/.prediction-bot/bot.db"
```

## Key Algorithms

### Kelly Criterion
```
f = (p * b - q) / b
where:
  p = win probability
  q = lose probability (1-p)
  b = odds (profit per dollar risked)
```

### Volatility Safety Margin
```
safety_margin = distance_to_strike / (2 * expected_move)
where:
  distance_to_strike = |current_price - strike| / current_price
  expected_move = volatility * sqrt(time_to_close / 365)
```

## External APIs

### Binance
- WebSocket: `wss://stream.binance.com:9443/ws`
- REST: `https://api.binance.com/api/v3`
- No auth for public data

### Alpha Vantage
- REST: `https://www.alphavantage.co/query`
- Auth: API key in query param
- Rate limit: 25/day (free tier)

### Polymarket
- CLOB API: `https://clob.polymarket.com`
- Auth: Wallet signature
- Rate limit: 100/min

### Kalshi
- Trade API: `https://api.elections.kalshi.com/trade-api/v2`
- Auth: HMAC signature
- Rate limit: 30/min

## Testing Strategy

- Unit tests for all business logic
- Integration tests for database operations
- E2E tests for funcionality and behavior
- Mock external APIs in tests
- Table-driven tests for edge cases

## Error Handling Pattern

```go
if err != nil {
    return fmt.Errorf("operation_name: %w", err)
}
```

Always wrap errors with context.

## Logging

Use structured logging (zerolog or zap):
```go
log.Info().
    Str("market_id", market.ID).
    Float64("probability", prob).
    Msg("found eligible market")
```

Levels:
- Debug: Detailed flow information
- Info: Normal operations
- Warn: Recoverable issues
- Error: Failures requiring attention
