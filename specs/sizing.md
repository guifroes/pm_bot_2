# Sizing

Calculates optimal position size using Kelly Criterion with conservative fractional adjustment.

## Kelly Criterion

Formula:
```
f = (p * b - q) / b

where:
  f = fraction of bankroll to bet
  p = probability of winning
  q = probability of losing (1 - p)
  b = odds received (net profit per $1 bet)
```

For prediction markets:
```
# Buying Yes at price $0.85 (85 cents)
# If win: receive $1.00, profit = $0.15
# b = profit / cost = 0.15 / 0.85 = 0.176

b = (1 - entry_price) / entry_price
```

## Fractional Kelly

Full Kelly is too aggressive. Use fractional Kelly:

```
position_size = bankroll * kelly_fraction * fractional_multiplier

where:
  fractional_multiplier = 0.25 to 0.50 (configurable, starts at 0.25)
```

## Input

```
{
  platform: "polymarket" | "kalshi",
  entry_price: number,           // e.g., 0.85 for 85%
  estimated_win_probability: number,  // our assessment, not market price
  volatility_safety_margin: number    // from volatility analyzer
}
```

## Win Probability Estimation

Since we're doing tail-end trading (high probability markets), our estimated win probability is:

```
# Base: use market price as starting point
base_prob = entry_price

# Adjust based on volatility safety margin
# Higher safety margin = higher confidence
if safety_margin >= 1.5:
    confidence_boost = 0.02  # +2%
elif safety_margin >= 1.0:
    confidence_boost = 0.00
else:
    confidence_boost = -0.05  # penalize risky trades

estimated_prob = min(0.99, base_prob + confidence_boost)
```

## Output

```
{
  kelly_fraction: number,        // raw Kelly result
  adjusted_fraction: number,     // after fractional multiplier
  position_size_usd: number,     // actual amount to bet
  position_size_shares: number,  // shares to buy
  reasoning: string
}
```

## Bankroll Management

Each platform has independent bankroll:

```
{
  polymarket: { initial: 50, current: 52.30 },
  kalshi: { initial: 50, current: 48.15 }
}
```

Position size is calculated against the specific platform's current bankroll.

## Constraints

- Minimum bet: $1 (skip if calculated size < $1)
- Maximum single bet: 20% of bankroll (hard cap regardless of Kelly)
- Never bet more than available balance
- Round down to platform's minimum increment

## Examples

### Example 1: Strong Trade
```
entry_price = 0.90 ($0.90 for Yes)
estimated_prob = 0.92 (92% win chance)
safety_margin = 1.8
bankroll = $50

b = (1 - 0.90) / 0.90 = 0.111
kelly = (0.92 * 0.111 - 0.08) / 0.111 = 0.20 (20%)
adjusted = 0.20 * 0.25 = 0.05 (5%)
position_size = $50 * 0.05 = $2.50
```

### Example 2: Marginal Trade
```
entry_price = 0.82
estimated_prob = 0.82 (no boost, safety_margin = 1.1)
bankroll = $50

b = 0.18 / 0.82 = 0.22
kelly = (0.82 * 0.22 - 0.18) / 0.22 = 0.0 (break-even)
position_size = $0 → skip trade
```

## Integration

- Called by: Position Manager
- Reads bankroll from: Persistence
- Fractional multiplier adjusted by: Learning System
