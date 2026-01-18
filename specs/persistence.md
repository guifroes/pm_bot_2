# Persistence

SQLite database for all bot state and history.

## Database File

```
~/.prediction-bot/bot.db
```

Single file, easy to backup and inspect.

## Schema

### positions
```sql
CREATE TABLE positions (
    id TEXT PRIMARY KEY,
    platform TEXT NOT NULL,  -- "polymarket" | "kalshi"
    market_id TEXT NOT NULL,
    market_title TEXT NOT NULL,
    
    -- Asset info
    underlying_symbol TEXT,
    underlying_type TEXT,  -- "crypto" | "stock" | "index"
    strike_price REAL,
    direction TEXT,  -- "above" | "below"
    
    -- Position
    side TEXT NOT NULL,  -- "yes" | "no"
    entry_price REAL NOT NULL,
    entry_time TEXT NOT NULL,  -- ISO8601
    quantity REAL NOT NULL,
    cost_basis REAL NOT NULL,
    
    -- Risk params at entry
    stop_loss_price REAL,
    volatility_at_entry REAL,
    safety_margin_at_entry REAL,
    
    -- Exit
    status TEXT NOT NULL DEFAULT 'open',  -- "open" | "closed"
    exit_price REAL,
    exit_time TEXT,
    exit_reason TEXT,  -- "resolution" | "stop_loss" | "volatility_exit" | "manual"
    realized_pnl REAL,
    
    -- Metadata
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_positions_status ON positions(status);
CREATE INDEX idx_positions_platform ON positions(platform);
CREATE INDEX idx_positions_market ON positions(platform, market_id);
```

### bankroll
```sql
CREATE TABLE bankroll (
    id INTEGER PRIMARY KEY,
    platform TEXT NOT NULL UNIQUE,
    initial_amount REAL NOT NULL,
    current_amount REAL NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### parameters
```sql
CREATE TABLE parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    probability_threshold REAL NOT NULL,
    volatility_safety_margin REAL NOT NULL,
    stop_loss_percent REAL NOT NULL,
    kelly_fraction REAL NOT NULL,
    
    version INTEGER NOT NULL,
    reason TEXT,  -- why this version was created
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Current parameters = highest version
CREATE INDEX idx_parameters_version ON parameters(version DESC);
```

### price_cache
```sql
CREATE TABLE price_cache (
    symbol TEXT NOT NULL,
    source TEXT NOT NULL,  -- "binance" | "alphavantage"
    price REAL NOT NULL,
    timestamp TEXT NOT NULL,
    PRIMARY KEY (symbol, source)
);

CREATE TABLE price_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    source TEXT NOT NULL,
    price REAL NOT NULL,
    timestamp TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_price_history_symbol ON price_history(symbol, timestamp DESC);
```

### api_log
```sql
CREATE TABLE api_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    platform TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    request_body TEXT,
    response_code INTEGER,
    response_body TEXT,
    duration_ms INTEGER,
    error TEXT
);

CREATE INDEX idx_api_log_timestamp ON api_log(timestamp DESC);
CREATE INDEX idx_api_log_platform ON api_log(platform, timestamp DESC);

-- Auto-cleanup: keep only last 7 days
-- Run periodically: DELETE FROM api_log WHERE timestamp < datetime('now', '-7 days');
```

### events
```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type TEXT NOT NULL,  -- "trade_opened" | "trade_closed" | "parameter_updated" | "error" | "warning"
    severity TEXT NOT NULL DEFAULT 'info',  -- "info" | "warning" | "error"
    message TEXT NOT NULL,
    data TEXT  -- JSON for additional context
);

CREATE INDEX idx_events_timestamp ON events(timestamp DESC);
CREATE INDEX idx_events_type ON events(type, timestamp DESC);
```

## Queries

### Get current parameters
```sql
SELECT * FROM parameters ORDER BY version DESC LIMIT 1;
```

### Get open positions
```sql
SELECT * FROM positions WHERE status = 'open';
```

### Get trade history with stats
```sql
SELECT 
    COUNT(*) as total_trades,
    SUM(CASE WHEN realized_pnl > 0 THEN 1 ELSE 0 END) as wins,
    SUM(CASE WHEN realized_pnl <= 0 THEN 1 ELSE 0 END) as losses,
    SUM(realized_pnl) as total_pnl,
    AVG(realized_pnl) as avg_pnl
FROM positions 
WHERE status = 'closed';
```

### Get recent trades for learning
```sql
SELECT * FROM positions 
WHERE status = 'closed' 
ORDER BY exit_time DESC 
LIMIT 20;
```

## Migrations

Store schema version and run migrations on startup:

```sql
CREATE TABLE schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Migration files in `migrations/` directory:
- `001_initial.sql`
- `002_add_price_history.sql`
- etc.

## Backup

- Automatic backup before each migration
- Manual backup command: `bot backup`
- Backup location: `~/.prediction-bot/backups/bot_YYYYMMDD_HHMMSS.db`

## Concurrency

SQLite with WAL mode for concurrent reads:

```sql
PRAGMA journal_mode=WAL;
PRAGMA busy_timeout=5000;
```

Single writer (bot), multiple readers (dashboard, CLI tools).

## Integration

- Used by: All components
- Location: `~/.prediction-bot/bot.db`
- Backup: Before migrations and on-demand
