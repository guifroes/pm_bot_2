# Volatility Analyzer

Calculates and validates whether the underlying asset's volatility supports the trade thesis.

## Purpose

Before entering a high-probability trade, verify that the underlying asset is unlikely to move enough to invalidate the prediction. A market showing 90% probability that "BTC stays above $95k" is only a good trade if BTC's current volatility suggests it won't drop below $95k.

## Core Calculation

For MVP: **Historical Standard Deviation** (7-14 day window)

```
volatility = std_dev(daily_returns, window=14)
annualized_volatility = volatility * sqrt(365)  # for crypto
                      = volatility * sqrt(252)  # for stocks
```

Future enhancement: GARCH(1,1) for adaptive volatility that responds faster to regime changes.

## Validation Logic

Given:
- `current_price`: Current price of underlying asset
- `strike_price`: The threshold in the prediction market
- `time_to_close`: Hours until market resolution
- `volatility`: Calculated volatility

Calculate probability of crossing strike:

```
distance_to_strike = abs(current_price - strike_price) / current_price
time_factor = sqrt(time_to_close / 24)  # normalize to days
expected_move = volatility * time_factor

# Simplified: if expected 2-sigma move doesn't reach strike, trade is valid
safety_margin = distance_to_strike / (2 * expected_move)
```

Trade is valid if `safety_margin >= 1.0` (configurable threshold).

## Input

```
{
  asset_symbol: string,
  asset_type: "crypto" | "stock" | "index",
  current_price: number,
  strike_price: number,
  direction: "above" | "below",
  hours_to_close: number
}
```

## Output

```
{
  volatility: number,              // annualized
  expected_move_percent: number,   // expected move in time window
  safety_margin: number,           // how many sigmas away is strike
  recommendation: "valid" | "risky" | "reject",
  reasoning: string
}
```

## Thresholds

Initial values (learning system will adjust):

| Safety Margin | Recommendation |
|--------------|----------------|
| >= 1.5       | valid          |
| 1.0 - 1.5    | risky          |
| < 1.0        | reject         |

## Data Requirements

- Crypto: 14 days of hourly prices (Binance)
- Stocks/indices: 14 days of daily prices (Alpha Vantage)

## Edge Cases

- Insufficient price history: reject trade
- Asset not found: reject trade
- Extreme volatility spike detected: flag as risky regardless of calculation
- Weekend/holiday for stocks: adjust time calculations

## Integration

- Receives requests from: Position Manager
- Fetches data from: Data Sources
- Results used by: Position Manager for trade decisions
