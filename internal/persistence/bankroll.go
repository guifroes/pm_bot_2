package persistence

import (
	"database/sql"
	"fmt"
)

// Bankroll represents a bankroll record in the database.
type Bankroll struct {
	ID            int64
	Platform      string
	InitialAmount float64
	CurrentAmount float64
	UpdatedAt     string
}

// BankrollRepository handles database operations for bankroll.
type BankrollRepository struct {
	db *sql.DB
}

// NewBankrollRepository creates a new BankrollRepository.
func NewBankrollRepository(db *sql.DB) *BankrollRepository {
	return &BankrollRepository{db: db}
}

// Get retrieves the bankroll for a specific platform.
func (r *BankrollRepository) Get(platform string) (*Bankroll, error) {
	b := &Bankroll{}
	err := r.db.QueryRow(`
		SELECT id, platform, initial_amount, current_amount, updated_at
		FROM bankroll WHERE platform = ?
	`, platform).Scan(&b.ID, &b.Platform, &b.InitialAmount, &b.CurrentAmount, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get bankroll: %w", err)
	}
	return b, nil
}

// GetAll retrieves all bankroll records.
func (r *BankrollRepository) GetAll() ([]*Bankroll, error) {
	rows, err := r.db.Query(`
		SELECT id, platform, initial_amount, current_amount, updated_at
		FROM bankroll ORDER BY platform
	`)
	if err != nil {
		return nil, fmt.Errorf("get all bankrolls: %w", err)
	}
	defer rows.Close()

	var bankrolls []*Bankroll
	for rows.Next() {
		b := &Bankroll{}
		if err := rows.Scan(&b.ID, &b.Platform, &b.InitialAmount, &b.CurrentAmount, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan bankroll: %w", err)
		}
		bankrolls = append(bankrolls, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bankrolls: %w", err)
	}
	return bankrolls, nil
}

// Update sets the current amount for a platform.
func (r *BankrollRepository) Update(platform string, amount float64) error {
	result, err := r.db.Exec(`
		UPDATE bankroll SET current_amount = ?, updated_at = CURRENT_TIMESTAMP
		WHERE platform = ?
	`, amount, platform)
	if err != nil {
		return fmt.Errorf("update bankroll: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bankroll not found for platform: %s", platform)
	}

	return nil
}

// Initialize creates a new bankroll record for a platform.
func (r *BankrollRepository) Initialize(platform string, amount float64) error {
	_, err := r.db.Exec(`
		INSERT INTO bankroll (platform, initial_amount, current_amount)
		VALUES (?, ?, ?)
		ON CONFLICT(platform) DO UPDATE SET
			initial_amount = excluded.initial_amount,
			current_amount = excluded.current_amount,
			updated_at = CURRENT_TIMESTAMP
	`, platform, amount, amount)
	if err != nil {
		return fmt.Errorf("initialize bankroll: %w", err)
	}
	return nil
}

// AddToBalance adds (or subtracts if negative) an amount to the current balance.
func (r *BankrollRepository) AddToBalance(platform string, amount float64) error {
	result, err := r.db.Exec(`
		UPDATE bankroll SET
			current_amount = current_amount + ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE platform = ?
	`, amount, platform)
	if err != nil {
		return fmt.Errorf("add to balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bankroll not found for platform: %s", platform)
	}

	return nil
}
