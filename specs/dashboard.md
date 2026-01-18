# Dashboard

Terminal-based real-time interface for monitoring bot activity.

## Display Layout

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  PREDICTION MARKET BOT                                   2024-01-15 14:32:05 ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  BANKROLL                                                                    ║
║  ├─ Polymarket: $52.30 (+$2.30 / +4.6%)                                     ║
║  └─ Kalshi:     $48.15 (-$1.85 / -3.7%)                                     ║
║  Total: $100.45 (+$0.45 / +0.4%)                                            ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  OPEN POSITIONS (2)                                                          ║
║  ┌────────────────────────────────────────────────────────────────────────┐  ║
║  │ POLY | BTC > $100k Jan 16 | YES @ 0.91 | $5.00 | +$0.35 (+7%) | 12h   │  ║
║  │ KLSH | SPY < 6000 Jan 16  | NO  @ 0.88 | $3.00 | -$0.15 (-5%) | 8h    │  ║
║  └────────────────────────────────────────────────────────────────────────┘  ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  SCANNER (last 10s)                                                          ║
║  ├─ Markets scanned: 847                                                     ║
║  ├─ Eligible: 12                                                             ║
║  └─ Passed volatility: 3                                                     ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  RECENT TRADES                                                               ║
║  ├─ 14:20 | WON  | ETH > $3500 | +$1.20 (+8%)                               ║
║  ├─ 12:05 | LOST | BTC > $98k  | -$2.00 (-15%) [stop loss]                  ║
║  └─ 09:30 | WON  | SPY > 5900  | +$0.80 (+5%)                               ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  PARAMETERS (v3)                                                             ║
║  ├─ Prob threshold: 0.83 (↑ from 0.80)                                      ║
║  ├─ Safety margin:  1.4  (↓ from 1.5)                                       ║
║  ├─ Stop loss:      15%                                                      ║
║  └─ Kelly fraction: 0.25                                                     ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  STATS (all time)                                                            ║
║  ├─ Trades: 23 | Won: 18 (78%) | Lost: 5                                    ║
║  ├─ Total PnL: +$4.50 | Avg: +$0.20/trade                                   ║
║  └─ Max drawdown: -$3.20 (6.4%)                                             ║
╚══════════════════════════════════════════════════════════════════════════════╝
  [Q]uit  [P]ause  [R]efresh  [D]etail  [L]ogs
```

## Sections

### Header
- Bot name and current timestamp
- Updates every second

### Bankroll
- Per-platform balance with absolute and percentage change from initial
- Total across all platforms

### Open Positions
- List of all current positions
- Shows: platform, market title (truncated), direction, entry price, size, unrealized PnL, time remaining
- Color coding: green for profit, red for loss, yellow for near stop loss

### Scanner Status
- Metrics from last scan cycle
- Markets scanned, eligible count, passed volatility check count

### Recent Trades
- Last 5 closed trades
- Shows: time, outcome, market, PnL, exit reason if not natural resolution

### Parameters
- Current learning system parameters
- Shows version and direction of last change

### Stats
- All-time statistics
- Win rate, total PnL, average PnL per trade, max drawdown

### Commands
- Keyboard shortcuts for interaction

## Interactivity

| Key | Action |
|-----|--------|
| Q | Quit bot gracefully (close positions first? prompt) |
| P | Pause/resume scanning (keeps monitoring open positions) |
| R | Force refresh all data |
| D | Show detail view for selected position |
| L | Toggle log panel (shows last 20 log lines) |
| ↑↓ | Navigate position list |
| Enter | Select position for detail view |

## Detail View (Position)

When pressing D on a position:

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  POSITION DETAIL                                                             ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  Market: Will Bitcoin be above $100,000 on January 16, 2024?                 ║
║  Platform: Polymarket                                                        ║
║  Direction: YES                                                              ║
║                                                                              ║
║  Entry:  0.91 @ 2024-01-15 08:30:00                                         ║
║  Current: 0.93                                                               ║
║  Size: $5.00 (5.5 shares)                                                    ║
║  PnL: +$0.35 (+7.0%)                                                        ║
║                                                                              ║
║  Underlying: BTC @ $101,234                                                  ║
║  Strike: $100,000 (above)                                                    ║
║  Distance: +1.2%                                                             ║
║                                                                              ║
║  Volatility at entry: 45% (annualized)                                       ║
║  Safety margin at entry: 1.8                                                 ║
║  Current safety margin: 1.6                                                  ║
║                                                                              ║
║  Stop loss: 0.77 (-15%)                                                      ║
║  Closes: 2024-01-16 00:00:00 (12h remaining)                                ║
╚══════════════════════════════════════════════════════════════════════════════╝
  [B]ack  [S]ell now
```

## Technical Implementation

- Use library like `tcell` or `bubbletea` for terminal UI in Go
- Refresh rate: 1 second for prices, 10 seconds for scanner stats
- Graceful terminal restore on exit

## Integration

- Receives data from: Position Manager, Market Scanner, Learning System
- Reads from: Persistence
- Sends commands to: Position Manager (pause, manual sell)
