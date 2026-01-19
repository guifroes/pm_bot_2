package persistence

import (
	"database/sql"
	"fmt"
	"time"
)

// Position represents a trading position in the database.
type Position struct {
	ID                  int64
	Platform            string
	MarketID            string
	MarketTitle         string
	Asset               string
	Strike              float64
	Direction           string
	EntryPrice          float64
	ExitPrice           *float64
	Quantity            float64
	Side                string
	Status              string
	EntryTime           time.Time
	ExitTime            *time.Time
	ExitReason          *string
	RealizedPnL         *float64
	SafetyMarginAtEntry float64
	VolatilityAtEntry   float64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// PositionRepository handles database operations for positions.
type PositionRepository struct {
	db *sql.DB
}

// NewPositionRepository creates a new PositionRepository.
func NewPositionRepository(db *sql.DB) *PositionRepository {
	return &PositionRepository{db: db}
}

// Create inserts a new position and returns its ID.
func (r *PositionRepository) Create(pos *Position) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO positions (
			platform, market_id, market_title, asset, strike, direction,
			entry_price, quantity, side, status,
			safety_margin_at_entry, volatility_at_entry
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		pos.Platform, pos.MarketID, pos.MarketTitle, pos.Asset, pos.Strike, pos.Direction,
		pos.EntryPrice, pos.Quantity, pos.Side, pos.Status,
		pos.SafetyMarginAtEntry, pos.VolatilityAtEntry,
	)
	if err != nil {
		return 0, fmt.Errorf("create position: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

// GetByID retrieves a position by its ID.
func (r *PositionRepository) GetByID(id int64) (*Position, error) {
	pos := &Position{}
	err := r.db.QueryRow(`
		SELECT id, platform, market_id, COALESCE(market_title, ''), COALESCE(asset, ''),
			COALESCE(strike, 0), COALESCE(direction, ''), entry_price, exit_price,
			quantity, side, status, entry_time, exit_time, exit_reason, realized_pnl,
			COALESCE(safety_margin_at_entry, 0), COALESCE(volatility_at_entry, 0),
			created_at, updated_at
		FROM positions WHERE id = ?
	`, id).Scan(
		&pos.ID, &pos.Platform, &pos.MarketID, &pos.MarketTitle, &pos.Asset,
		&pos.Strike, &pos.Direction, &pos.EntryPrice, &pos.ExitPrice,
		&pos.Quantity, &pos.Side, &pos.Status, &pos.EntryTime, &pos.ExitTime,
		&pos.ExitReason, &pos.RealizedPnL,
		&pos.SafetyMarginAtEntry, &pos.VolatilityAtEntry,
		&pos.CreatedAt, &pos.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get position by id: %w", err)
	}
	return pos, nil
}

// GetOpen retrieves all open positions.
func (r *PositionRepository) GetOpen() ([]*Position, error) {
	rows, err := r.db.Query(`
		SELECT id, platform, market_id, COALESCE(market_title, ''), COALESCE(asset, ''),
			COALESCE(strike, 0), COALESCE(direction, ''), entry_price, exit_price,
			quantity, side, status, entry_time, exit_time, exit_reason, realized_pnl,
			COALESCE(safety_margin_at_entry, 0), COALESCE(volatility_at_entry, 0),
			created_at, updated_at
		FROM positions WHERE status = 'open'
		ORDER BY entry_time DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("get open positions: %w", err)
	}
	defer rows.Close()

	return r.scanPositions(rows)
}

// GetClosed retrieves all closed positions.
func (r *PositionRepository) GetClosed() ([]*Position, error) {
	rows, err := r.db.Query(`
		SELECT id, platform, market_id, COALESCE(market_title, ''), COALESCE(asset, ''),
			COALESCE(strike, 0), COALESCE(direction, ''), entry_price, exit_price,
			quantity, side, status, entry_time, exit_time, exit_reason, realized_pnl,
			COALESCE(safety_margin_at_entry, 0), COALESCE(volatility_at_entry, 0),
			created_at, updated_at
		FROM positions WHERE status = 'closed'
		ORDER BY exit_time DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("get closed positions: %w", err)
	}
	defer rows.Close()

	return r.scanPositions(rows)
}

// GetOpenByPlatform retrieves all open positions for a specific platform.
func (r *PositionRepository) GetOpenByPlatform(platform string) ([]*Position, error) {
	rows, err := r.db.Query(`
		SELECT id, platform, market_id, COALESCE(market_title, ''), COALESCE(asset, ''),
			COALESCE(strike, 0), COALESCE(direction, ''), entry_price, exit_price,
			quantity, side, status, entry_time, exit_time, exit_reason, realized_pnl,
			COALESCE(safety_margin_at_entry, 0), COALESCE(volatility_at_entry, 0),
			created_at, updated_at
		FROM positions WHERE status = 'open' AND platform = ?
		ORDER BY entry_time DESC
	`, platform)
	if err != nil {
		return nil, fmt.Errorf("get open positions by platform: %w", err)
	}
	defer rows.Close()

	return r.scanPositions(rows)
}

// GetByMarket retrieves an open position by platform and market ID.
func (r *PositionRepository) GetByMarket(platform, marketID string) (*Position, error) {
	pos := &Position{}
	err := r.db.QueryRow(`
		SELECT id, platform, market_id, COALESCE(market_title, ''), COALESCE(asset, ''),
			COALESCE(strike, 0), COALESCE(direction, ''), entry_price, exit_price,
			quantity, side, status, entry_time, exit_time, exit_reason, realized_pnl,
			COALESCE(safety_margin_at_entry, 0), COALESCE(volatility_at_entry, 0),
			created_at, updated_at
		FROM positions WHERE platform = ? AND market_id = ? AND status = 'open'
	`, platform, marketID).Scan(
		&pos.ID, &pos.Platform, &pos.MarketID, &pos.MarketTitle, &pos.Asset,
		&pos.Strike, &pos.Direction, &pos.EntryPrice, &pos.ExitPrice,
		&pos.Quantity, &pos.Side, &pos.Status, &pos.EntryTime, &pos.ExitTime,
		&pos.ExitReason, &pos.RealizedPnL,
		&pos.SafetyMarginAtEntry, &pos.VolatilityAtEntry,
		&pos.CreatedAt, &pos.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get position by market: %w", err)
	}
	return pos, nil
}

// Update updates an existing position.
func (r *PositionRepository) Update(pos *Position) error {
	_, err := r.db.Exec(`
		UPDATE positions SET
			market_title = ?,
			asset = ?,
			strike = ?,
			direction = ?,
			entry_price = ?,
			exit_price = ?,
			quantity = ?,
			side = ?,
			status = ?,
			exit_time = ?,
			exit_reason = ?,
			realized_pnl = ?,
			safety_margin_at_entry = ?,
			volatility_at_entry = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		pos.MarketTitle, pos.Asset, pos.Strike, pos.Direction,
		pos.EntryPrice, pos.ExitPrice, pos.Quantity, pos.Side, pos.Status,
		pos.ExitTime, pos.ExitReason, pos.RealizedPnL,
		pos.SafetyMarginAtEntry, pos.VolatilityAtEntry,
		pos.ID,
	)
	if err != nil {
		return fmt.Errorf("update position: %w", err)
	}
	return nil
}

// Close marks a position as closed with exit details.
func (r *PositionRepository) Close(id int64, exitPrice float64, reason string, pnl float64) error {
	_, err := r.db.Exec(`
		UPDATE positions SET
			status = 'closed',
			exit_price = ?,
			exit_time = CURRENT_TIMESTAMP,
			exit_reason = ?,
			realized_pnl = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, exitPrice, reason, pnl, id)
	if err != nil {
		return fmt.Errorf("close position: %w", err)
	}
	return nil
}

// scanPositions scans multiple positions from rows.
func (r *PositionRepository) scanPositions(rows *sql.Rows) ([]*Position, error) {
	var positions []*Position
	for rows.Next() {
		pos := &Position{}
		err := rows.Scan(
			&pos.ID, &pos.Platform, &pos.MarketID, &pos.MarketTitle, &pos.Asset,
			&pos.Strike, &pos.Direction, &pos.EntryPrice, &pos.ExitPrice,
			&pos.Quantity, &pos.Side, &pos.Status, &pos.EntryTime, &pos.ExitTime,
			&pos.ExitReason, &pos.RealizedPnL,
			&pos.SafetyMarginAtEntry, &pos.VolatilityAtEntry,
			&pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		positions = append(positions, pos)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions: %w", err)
	}
	return positions, nil
}
