-- Initial schema: schema_version and bankroll tables

CREATE TABLE IF NOT EXISTS bankroll (
    id INTEGER PRIMARY KEY,
    platform TEXT NOT NULL UNIQUE,
    initial_amount REAL NOT NULL,
    current_amount REAL NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default bankroll values
INSERT INTO bankroll (platform, initial_amount, current_amount) VALUES ('polymarket', 50.0, 50.0);
INSERT INTO bankroll (platform, initial_amount, current_amount) VALUES ('kalshi', 50.0, 50.0);
