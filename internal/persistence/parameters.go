package persistence

import (
	"database/sql"
	"fmt"
	"time"
)

// Parameter represents a trading parameter with its bounds.
type Parameter struct {
	Name      string
	Value     float64
	MinValue  float64
	MaxValue  float64
	UpdatedAt time.Time
}

// ParameterChange represents a historical parameter adjustment.
type ParameterChange struct {
	ID        int64
	Name      string
	OldValue  float64
	NewValue  float64
	Reason    string
	CreatedAt time.Time
}

// ParametersRepository manages trading parameters in the database.
type ParametersRepository struct {
	db *sql.DB
}

// NewParametersRepository creates a new ParametersRepository.
func NewParametersRepository(db *sql.DB) *ParametersRepository {
	return &ParametersRepository{db: db}
}

// GetCurrent returns all current parameter values as a map.
func (r *ParametersRepository) GetCurrent() (map[string]Parameter, error) {
	rows, err := r.db.Query(`
		SELECT name, value, COALESCE(min_value, 0), COALESCE(max_value, 1),
		       COALESCE(updated_at, CURRENT_TIMESTAMP)
		FROM parameters
	`)
	if err != nil {
		return nil, fmt.Errorf("query parameters: %w", err)
	}
	defer rows.Close()

	params := make(map[string]Parameter)
	for rows.Next() {
		var p Parameter
		var updatedAtStr string
		if err := rows.Scan(&p.Name, &p.Value, &p.MinValue, &p.MaxValue, &updatedAtStr); err != nil {
			return nil, fmt.Errorf("scan parameter: %w", err)
		}
		p.UpdatedAt = parseTimestamp(updatedAtStr)
		params[p.Name] = p
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate parameters: %w", err)
	}

	return params, nil
}

// GetByName returns a specific parameter by name.
func (r *ParametersRepository) GetByName(name string) (Parameter, error) {
	var p Parameter
	var updatedAtStr string

	err := r.db.QueryRow(`
		SELECT name, value, COALESCE(min_value, 0), COALESCE(max_value, 1),
		       COALESCE(updated_at, CURRENT_TIMESTAMP)
		FROM parameters
		WHERE name = ?
	`, name).Scan(&p.Name, &p.Value, &p.MinValue, &p.MaxValue, &updatedAtStr)

	if err == sql.ErrNoRows {
		return Parameter{}, fmt.Errorf("parameter not found: %s", name)
	}
	if err != nil {
		return Parameter{}, fmt.Errorf("query parameter %s: %w", name, err)
	}

	p.UpdatedAt = parseTimestamp(updatedAtStr)
	return p, nil
}

// Save updates a parameter value without recording history.
func (r *ParametersRepository) Save(name string, value float64) error {
	result, err := r.db.Exec(`
		UPDATE parameters
		SET value = ?, updated_at = CURRENT_TIMESTAMP
		WHERE name = ?
	`, value, name)
	if err != nil {
		return fmt.Errorf("update parameter %s: %w", name, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("parameter not found: %s", name)
	}

	return nil
}

// SaveWithReason updates a parameter value and records the change in history.
func (r *ParametersRepository) SaveWithReason(name string, value float64, reason string) error {
	// Get current value for history
	current, err := r.GetByName(name)
	if err != nil {
		return fmt.Errorf("get current value: %w", err)
	}

	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update parameter
	_, err = tx.Exec(`
		UPDATE parameters
		SET value = ?, updated_at = CURRENT_TIMESTAMP
		WHERE name = ?
	`, value, name)
	if err != nil {
		return fmt.Errorf("update parameter: %w", err)
	}

	// Record history
	_, err = tx.Exec(`
		INSERT INTO parameter_history (name, old_value, new_value, reason)
		VALUES (?, ?, ?, ?)
	`, name, current.Value, value, reason)
	if err != nil {
		return fmt.Errorf("insert history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetHistory returns the most recent parameter changes.
func (r *ParametersRepository) GetHistory(name string, limit int) ([]ParameterChange, error) {
	rows, err := r.db.Query(`
		SELECT id, name, old_value, new_value, COALESCE(reason, ''),
		       COALESCE(created_at, CURRENT_TIMESTAMP)
		FROM parameter_history
		WHERE name = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, name, limit)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var changes []ParameterChange
	for rows.Next() {
		var c ParameterChange
		var createdAtStr string
		if err := rows.Scan(&c.ID, &c.Name, &c.OldValue, &c.NewValue, &c.Reason, &createdAtStr); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		c.CreatedAt = parseTimestamp(createdAtStr)
		changes = append(changes, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history: %w", err)
	}

	return changes, nil
}

// GetLastAdjustmentTime returns the time of the most recent adjustment for a parameter.
// Returns zero time if no adjustments have been made.
func (r *ParametersRepository) GetLastAdjustmentTime(name string) (time.Time, error) {
	var createdAtStr sql.NullString

	err := r.db.QueryRow(`
		SELECT created_at
		FROM parameter_history
		WHERE name = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, name).Scan(&createdAtStr)

	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("query last adjustment: %w", err)
	}

	if !createdAtStr.Valid {
		return time.Time{}, nil
	}

	return parseTimestamp(createdAtStr.String), nil
}

// parseTimestamp attempts to parse a timestamp string from SQLite.
func parseTimestamp(s string) time.Time {
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
