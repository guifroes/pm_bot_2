# Learning System

Automatically adjusts trading parameters based on realized outcomes.

## Purpose

The bot starts with conservative default parameters. As it accumulates trade history, it learns which settings produce better results and adjusts in real-time.

## Learnable Parameters

| Parameter | Initial Value | Range | Description |
|-----------|---------------|-------|-------------|
| `probability_threshold` | 0.80 | 0.70 - 0.95 | Minimum probability to consider market |
| `volatility_safety_margin` | 1.5 | 1.0 - 2.5 | Required distance from strike |
| `stop_loss_percent` | 0.15 | 0.05 - 0.30 | When to cut losses |
| `kelly_fraction` | 0.25 | 0.10 - 0.50 | Kelly multiplier for position size |

## Learning Trigger

After each trade closes:

```go
type TradeOutcome struct {
    TradeID      string
    WonOrLost    string   // "won", "lost", "early_exit"
    PnL          float64
    PnLPercent   float64
    
    // Parameters at entry
    EntryProbability   float64
    SafetyMargin       float64
    StopLossUsed       float64
    KellyFractionUsed  float64
    
    // Context
    TimeHeld     time.Duration
    ExitReason   string
}
```

## Adjustment Algorithm

Simple windowed analysis (no complex ML for MVP):

```
1. Collect last N trades (N = 20 minimum for statistical significance)
2. Segment by parameter value ranges
3. Calculate win rate and average PnL for each segment
4. Shift parameter toward better-performing range
```

### Example: Adjusting Probability Threshold

```
Trades with entry_probability 0.80-0.85: 10 trades, 60% win rate, avg PnL +2%
Trades with entry_probability 0.85-0.90: 8 trades, 75% win rate, avg PnL +4%
Trades with entry_probability 0.90-0.95: 5 trades, 90% win rate, avg PnL +1%

→ Best risk-adjusted returns at 0.85-0.90
→ Adjust probability_threshold from 0.80 to 0.83 (gradual shift)
```

## Adjustment Rules

1. **Minimum sample size**: Need 20+ trades before any adjustment
2. **Gradual changes**: Max 10% adjustment per learning cycle
3. **Cooldown**: Wait 5 trades between adjustments to same parameter
4. **Bounds**: Never exceed defined range limits
5. **Revert on drawdown**: If bankroll drops 20%, revert to initial values

## Parameter Storage

```go
type Parameters struct {
    ProbabilityThreshold  float64   `json:"probability_threshold"`
    SafetyMargin          float64   `json:"volatility_safety_margin"`
    StopLossPercent       float64   `json:"stop_loss_percent"`
    KellyFraction         float64   `json:"kelly_fraction"`
    
    // Metadata
    UpdatedAt             time.Time `json:"updated_at"`
    TradeCountAtUpdate    int       `json:"trade_count_at_update"`
    Version               int       `json:"version"`
}
```

Store in SQLite. Keep history of all parameter versions for analysis.

## Reporting

Dashboard shows:
- Current parameter values
- Historical parameter evolution graph
- Win rate by parameter segment
- Last adjustment reason

## Safety Guardrails

- **Hard limits**: Parameters cannot exceed defined ranges
- **Manual override**: User can lock any parameter to prevent learning
- **Emergency stop**: If 5 consecutive losses, pause learning and alert
- **Audit trail**: All adjustments logged with reasoning

## Future Enhancements

1. Multi-armed bandit for exploration vs exploitation
2. Bayesian optimization for parameter tuning
3. Per-platform parameter learning
4. Per-asset-type parameter learning (crypto vs stocks may need different settings)

## Integration

- Receives trade outcomes from: Position Manager
- Stores parameters in: Persistence
- Parameters read by: Market Scanner, Volatility Analyzer, Position Manager, Sizing
- Reports to: Dashboard
