# Position Manager

Orchestrates the full lifecycle of trades: entry decisions, monitoring, and exits.

## Responsibilities

1. Receive eligible markets from Scanner
2. Request volatility validation
3. Calculate position size via Sizing
4. Execute entry orders
5. Monitor open positions continuously
6. Execute exits (stop loss, take profit, or volatility-based)

## Entry Flow

```
1. Scanner finds eligible market
2. Position Manager checks: already have position in this market?
3. If no existing position:
   a. Request volatility analysis
   b. If volatility says "valid" or "risky":
      - Calculate position size (Sizing)
      - Execute buy order
      - Record position in Persistence
   c. If volatility says "reject": skip market
```

## Position State

```
{
  id: uuid,
  platform: "polymarket" | "kalshi",
  market_id: string,
  direction: "yes" | "no",
  entry_price: number,
  entry_time: timestamp,
  quantity: number,
  cost_basis: number,
  
  // Monitoring
  current_price: number,
  unrealized_pnl: number,
  unrealized_pnl_percent: number,
  
  // Risk params (from entry)
  stop_loss_price: number,
  volatility_at_entry: number,
  safety_margin_at_entry: number,
  
  // Exit
  status: "open" | "closed",
  exit_price: number | null,
  exit_time: timestamp | null,
  exit_reason: "resolution" | "stop_loss" | "volatility_exit" | "manual" | null,
  realized_pnl: number | null
}
```

## Exit Conditions

Check every 10 seconds for open positions:

### 1. Stop Loss (15% initial, adjustable by learning)
```
if current_price <= entry_price * (1 - stop_loss_percent):
    execute_sell("stop_loss")
```

### 2. Market Resolution
```
if market.is_resolved:
    record_outcome()  # win or lose
```

### 3. Volatility-Based Exit (Key Innovation)
Even if in profit, exit if underlying asset volatility increases enough that the trade becomes risky:

```
1. Re-run volatility analysis with current data
2. If new safety_margin < 0.8:  # getting too close to strike
   execute_sell("volatility_exit")
```

This protects profits when market conditions change.

## Order Execution

- Order type: Market orders for simplicity (MVP)
- Slippage tolerance: 2% (configurable)
- If order fails: retry 3x with backoff, then abandon and log

## Constraints

- Maximum one position per market
- No position if bankroll too low (< $5 remaining)
- Respect platform-specific limits

## Integration

- Receives eligible markets from: Market Scanner
- Requests analysis from: Volatility Analyzer
- Requests sizing from: Sizing
- Executes via: Platform Integration
- Stores state in: Persistence
- Reports to: Dashboard
- Feeds outcomes to: Learning System
