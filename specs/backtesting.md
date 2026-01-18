# Backtesting

Test the trading strategy against historical data before risking real money.

## Purpose

1. Validate strategy logic before deployment
2. Tune initial parameters
3. Understand expected performance characteristics
4. Identify edge cases and failure modes

## Data Requirements

### Historical Market Data
Need historical prediction market data:
- Market open/close times
- Price history (Yes/No prices over time)
- Resolution outcomes
- Order book snapshots (if available)

**Sources**:
- Polymarket: Historical data via their API or third-party archives
- Kalshi: Historical API endpoints
- Manual collection: Start collecting now for future use

### Historical Asset Prices
Already available via Data Sources:
- Binance historical klines (crypto)
- Alpha Vantage historical prices (stocks)

## Backtest Engine

```go
type BacktestConfig struct {
    StartDate   time.Time
    EndDate     time.Time
    InitialBankroll map[string]float64  // per platform
    Parameters  Parameters  // strategy parameters to test
    
    // Simulation settings
    SlippageBps int     // basis points of slippage to simulate
    FeesPercent float64 // trading fees
}

type BacktestResult struct {
    // Performance
    FinalBankroll   map[string]float64
    TotalReturn     float64
    TotalReturnPct  float64
    
    // Risk metrics
    MaxDrawdown     float64
    MaxDrawdownPct  float64
    SharpeRatio     float64
    
    // Trade stats
    TotalTrades     int
    WinRate         float64
    AvgWinPnl       float64
    AvgLossPnl      float64
    AvgHoldTime     time.Duration
    
    // By exit reason
    ExitsByResolution    int
    ExitsByStopLoss      int
    ExitsByVolatility    int
    
    // Timeline
    EquityCurve     []EquityPoint
    TradeLog        []BacktestTrade
}

type EquityPoint struct {
    Timestamp time.Time
    Equity    float64
}

type BacktestTrade struct {
    MarketID    string
    EntryTime   time.Time
    ExitTime    time.Time
    EntryPrice  float64
    ExitPrice   float64
    PnL         float64
    ExitReason  string
}
```

## Simulation Logic

```
for each timestamp in simulation_period:
    1. Update current prices from historical data
    
    2. Check open positions:
       - Would stop loss trigger?
       - Would volatility exit trigger?
       - Did market resolve?
       
    3. Scan for new opportunities:
       - Apply same eligibility criteria as live
       - Check volatility using historical asset prices
       - Calculate position size
       
    4. Execute simulated trades:
       - Apply slippage to entry/exit prices
       - Deduct fees
       - Update bankroll
       
    5. Record equity curve point
```

## Parameter Sweep

Test multiple parameter combinations:

```go
type ParameterSweep struct {
    ProbabilityThresholds []float64  // e.g., [0.75, 0.80, 0.85, 0.90]
    SafetyMargins         []float64  // e.g., [1.0, 1.25, 1.5, 2.0]
    StopLossPercents      []float64  // e.g., [0.10, 0.15, 0.20]
    KellyFractions        []float64  // e.g., [0.15, 0.25, 0.35]
}

// Run all combinations, output ranked by risk-adjusted return
```

## Output Reports

### Summary Report (terminal)
```
╔══════════════════════════════════════════════════════════════════════════════╗
║  BACKTEST RESULTS: 2024-01-01 to 2024-06-30                                 ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  Initial: $100.00  →  Final: $142.50  (+42.5%)                              ║
║                                                                              ║
║  Trades: 156 | Won: 124 (79%) | Lost: 32                                    ║
║  Avg Win: +$0.85 | Avg Loss: -$1.20 | Expectancy: +$0.41                    ║
║                                                                              ║
║  Max Drawdown: $12.30 (8.6%)                                                ║
║  Sharpe Ratio: 1.82                                                          ║
║                                                                              ║
║  Exit Reasons:                                                               ║
║  ├─ Resolution: 130 (83%)                                                    ║
║  ├─ Stop Loss: 18 (12%)                                                      ║
║  └─ Volatility: 8 (5%)                                                       ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

### Detailed CSV Export
- `backtest_trades.csv`: Every trade with full details
- `backtest_equity.csv`: Equity curve data points
- `backtest_parameters.csv`: Parameter sweep results

### Equity Curve Chart
ASCII chart in terminal, or export data for external plotting.

## CLI Commands

```bash
# Run single backtest with current parameters
bot backtest --start 2024-01-01 --end 2024-06-30

# Run with specific parameters
bot backtest --prob-threshold 0.85 --safety-margin 1.5

# Run parameter sweep
bot backtest --sweep

# Export results
bot backtest --output results/
```

## Limitations & Caveats

Display prominently in output:

1. **Survivorship bias**: We only see markets that existed; failed markets may be missing
2. **Liquidity simulation**: Backtest assumes infinite liquidity; real execution may differ
3. **Look-ahead bias**: Ensure no future data leaks into decisions
4. **Market impact**: Large positions would move prices; not simulated
5. **Data quality**: Historical data may have gaps or errors
6. **Regime changes**: Past performance doesn't guarantee future results

## Integration

- Uses: Data Sources (historical prices)
- Uses: Volatility Analyzer (same logic as live)
- Uses: Sizing (same logic as live)
- Output to: Dashboard (results display), Files (CSV export)
- Independent of: Platform Integration (no real API calls)
