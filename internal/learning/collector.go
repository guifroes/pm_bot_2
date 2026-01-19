package learning

import (
	"database/sql"
	"fmt"
	"time"
)

// TradeOutcome represents a completed trade with all its parameters and results.
type TradeOutcome struct {
	PositionID  int64
	Platform    string
	Asset       string
	Strike      float64
	Direction   string
	Side        string
	EntryPrice  float64
	ExitPrice   float64
	Quantity    float64
	RealizedPnL float64
	EntryTime   time.Time
	ExitTime    time.Time
	ExitReason  string

	// Parameters used at entry time
	SafetyMargin float64
	Volatility   float64
}

// IsWin returns true if the trade had a positive PnL.
func (t TradeOutcome) IsWin() bool {
	return t.RealizedPnL > 0
}

// ReturnPercent calculates the percentage return on the trade.
// Formula: (exit - entry) / entry * 100
func (t TradeOutcome) ReturnPercent() float64 {
	if t.EntryPrice == 0 {
		return 0
	}
	return (t.ExitPrice - t.EntryPrice) / t.EntryPrice * 100
}

// Collector collects trade outcomes from the database.
type Collector struct {
	db *sql.DB
}

// NewCollector creates a new Collector.
func NewCollector(db *sql.DB) *Collector {
	return &Collector{db: db}
}

// CollectOutcomes retrieves closed trades from the database.
// Returns empty slice if there are fewer than minTrades closed positions.
// Results are ordered by exit time descending (most recent first).
func (c *Collector) CollectOutcomes(minTrades int) ([]TradeOutcome, error) {
	// First, count how many closed trades we have
	var count int
	err := c.db.QueryRow(`
		SELECT COUNT(*) FROM positions WHERE status = 'closed'
	`).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("count closed positions: %w", err)
	}

	// If not enough trades, return empty slice
	if count < minTrades {
		return []TradeOutcome{}, nil
	}

	// Query closed positions ordered by exit time desc, limited to minTrades
	rows, err := c.db.Query(`
		SELECT
			id, platform, COALESCE(asset, ''), COALESCE(strike, 0),
			COALESCE(direction, ''), side, entry_price, COALESCE(exit_price, 0),
			quantity, COALESCE(realized_pnl, 0), entry_time, COALESCE(exit_time, entry_time),
			COALESCE(exit_reason, ''),
			COALESCE(safety_margin_at_entry, 0), COALESCE(volatility_at_entry, 0)
		FROM positions
		WHERE status = 'closed'
		ORDER BY exit_time DESC
		LIMIT ?
	`, minTrades)
	if err != nil {
		return nil, fmt.Errorf("query closed positions: %w", err)
	}
	defer rows.Close()

	var outcomes []TradeOutcome
	for rows.Next() {
		var o TradeOutcome
		var entryTimeStr, exitTimeStr string
		err := rows.Scan(
			&o.PositionID, &o.Platform, &o.Asset, &o.Strike,
			&o.Direction, &o.Side, &o.EntryPrice, &o.ExitPrice,
			&o.Quantity, &o.RealizedPnL, &entryTimeStr, &exitTimeStr,
			&o.ExitReason,
			&o.SafetyMargin, &o.Volatility,
		)
		if err != nil {
			return nil, fmt.Errorf("scan trade outcome: %w", err)
		}

		// Parse timestamps from SQLite format
		o.EntryTime = parseTime(entryTimeStr)
		o.ExitTime = parseTime(exitTimeStr)

		outcomes = append(outcomes, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trade outcomes: %w", err)
	}

	return outcomes, nil
}

// parseTime attempts to parse a timestamp string from SQLite.
// Returns zero time on failure.
func parseTime(s string) time.Time {
	// Try common SQLite datetime formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	return time.Time{}
}
