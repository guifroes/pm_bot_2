# Market Scanner

## Overview

The market scanner continuously monitors Polymarket and Kalshi for eligible trading opportunities. It identifies binary markets with high probability (>80%) that close within 2 days and have a verifiable underlying asset (crypto or stock/index).

## Eligibility Criteria

A market is eligible when ALL of the following conditions are met:

1. **Market Type**: Binary (Yes/No outcome only)
2. **Probability Threshold**: Current price >= $0.80 (80% implied probability) for either Yes or No
3. **Time to Resolution**: Closes within 48 hours (2 days)
4. **Underlying Asset**: Market references a trackable financial asset:
   - Cryptocurrencies (BTC, ETH, etc.)
   - Stock indices (S&P 500, NASDAQ, etc.)
   - Individual stocks
5. **Liquidity**: Sufficient order book depth to execute trades (minimum $100 available at target price)
6. **Market Category**: Price/value predictions (e.g., "Will BTC be above $100k tomorrow?")

## Scan Behavior

- **Frequency**: Every 10 seconds
- **Platforms**: Polymarket, Kalshi (extensible to others)
- **Parallel Scanning**: Both platforms scanned simultaneously
- **Deduplication**: Same market on different platforms identified for potential arbitrage

## Output

For each eligible market, the scanner emits:

```
{
  platform: "polymarket" | "kalshi",
  market_id: string,
  title: string,
  underlying_asset: {
    symbol: string,        // e.g., "BTC", "SPY"
    type: "crypto" | "stock" | "index"
  },
  strike_price: number,    // The price threshold in the market
  direction: "above" | "below",
  current_probability: {
    yes: number,           // 0.0 to 1.0
    no: number
  },
  closes_at: timestamp,
  order_book: {
    best_bid: number,
    best_ask: number,
    depth_at_bid: number   // USD available
  }
}
```

## Market Parsing

The scanner must parse market titles to extract:

1. **Underlying Asset**: Identify which asset the market references
2. **Strike Price**: The threshold value (e.g., "$100,000" in "Will BTC close above $100,000?")
3. **Direction**: Whether the market is "above" or "below" the strike
4. **Resolution Time**: When the market closes

Examples:
- "Will Bitcoin be above $100,000 on Jan 18?" → BTC, $100000, above, Jan 18
- "S&P 500 to close below 6000 tomorrow?" → SPY, 6000, below, tomorrow

## Error Handling

- API rate limits: Implement exponential backoff
- Network failures: Retry with backoff, log failures
- Malformed markets: Skip and log warning
- Missing data: Skip market, do not guess

## Integration Points

- **Output to**: Position Manager (for trade decisions)
- **Input from**: Platform Integration (API responses)
- **Notifies**: Dashboard (new eligible markets found)
