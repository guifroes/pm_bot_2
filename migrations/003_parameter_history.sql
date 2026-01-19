-- Parameter history for tracking adjustments over time
CREATE TABLE parameter_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    old_value REAL NOT NULL,
    new_value REAL NOT NULL,
    reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_parameter_history_name ON parameter_history(name);
CREATE INDEX idx_parameter_history_created_at ON parameter_history(created_at);
