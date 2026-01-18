-- Core tables for the prediction market bot

-- Positions: tracks all trades (open and closed)
CREATE TABLE positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    platform TEXT NOT NULL,
    market_id TEXT NOT NULL,
    market_title TEXT,
    asset TEXT,
    strike REAL,
    direction TEXT,
    entry_price REAL NOT NULL,
    exit_price REAL,
    quantity REAL NOT NULL,
    side TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    entry_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    exit_time DATETIME,
    exit_reason TEXT,
    realized_pnl REAL,
    safety_margin_at_entry REAL,
    volatility_at_entry REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_positions_status ON positions(status);
CREATE INDEX idx_positions_platform ON positions(platform);
CREATE INDEX idx_positions_market_id ON positions(market_id);

-- Parameters: trading parameters (can be adjusted by learning system)
CREATE TABLE parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    value REAL NOT NULL,
    min_value REAL,
    max_value REAL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default parameters
INSERT INTO parameters (name, value, min_value, max_value) VALUES
    ('probability_threshold', 0.80, 0.75, 0.95),
    ('volatility_safety_margin', 1.5, 1.0, 3.0),
    ('stop_loss_percent', 0.15, 0.05, 0.30),
    ('kelly_fraction', 0.25, 0.10, 0.50);

-- Events: significant bot events for logging/debugging
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    platform TEXT,
    market_id TEXT,
    position_id INTEGER,
    details TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (position_id) REFERENCES positions(id)
);

CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_created_at ON events(created_at);

-- Price cache: temporary cache for current prices
CREATE TABLE price_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL UNIQUE,
    price REAL NOT NULL,
    source TEXT NOT NULL,
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Price history: historical price data for volatility calculation
CREATE TABLE price_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    price REAL NOT NULL,
    timestamp DATETIME NOT NULL,
    source TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_price_history_symbol ON price_history(symbol);
CREATE INDEX idx_price_history_timestamp ON price_history(timestamp);
CREATE UNIQUE INDEX idx_price_history_symbol_timestamp ON price_history(symbol, timestamp);

-- API log: track all API calls for rate limiting and debugging
CREATE TABLE api_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL DEFAULT 'GET',
    status_code INTEGER,
    response_time_ms INTEGER,
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_log_api ON api_log(api);
CREATE INDEX idx_api_log_created_at ON api_log(created_at);
