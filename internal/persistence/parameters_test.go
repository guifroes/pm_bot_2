package persistence

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParametersRepository_GetCurrent(t *testing.T) {
	// Setup temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Get migrations dir
	wd, _ := os.Getwd()
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	repo := NewParametersRepository(db)

	// Get all parameters
	params, err := repo.GetCurrent()
	if err != nil {
		t.Fatalf("get current: %v", err)
	}

	// Should have default parameters from migration
	if len(params) < 4 {
		t.Errorf("expected at least 4 parameters, got %d", len(params))
	}

	// Check probability_threshold exists with correct value
	if p, ok := params["probability_threshold"]; !ok {
		t.Error("missing probability_threshold")
	} else if p.Value != 0.80 {
		t.Errorf("probability_threshold: expected 0.80, got %v", p.Value)
	}

	// Check bounds are loaded
	if p, ok := params["probability_threshold"]; ok {
		if p.MinValue != 0.75 || p.MaxValue != 0.95 {
			t.Errorf("probability_threshold bounds: expected [0.75, 0.95], got [%v, %v]", p.MinValue, p.MaxValue)
		}
	}
}

func TestParametersRepository_GetByName(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	wd, _ := os.Getwd()
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	repo := NewParametersRepository(db)

	// Get specific parameter
	param, err := repo.GetByName("kelly_fraction")
	if err != nil {
		t.Fatalf("get by name: %v", err)
	}

	if param.Value != 0.25 {
		t.Errorf("kelly_fraction: expected 0.25, got %v", param.Value)
	}
	if param.MinValue != 0.10 || param.MaxValue != 0.50 {
		t.Errorf("bounds: expected [0.10, 0.50], got [%v, %v]", param.MinValue, param.MaxValue)
	}
}

func TestParametersRepository_GetByName_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	wd, _ := os.Getwd()
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	repo := NewParametersRepository(db)

	_, err = repo.GetByName("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent parameter")
	}
}

func TestParametersRepository_Save(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	wd, _ := os.Getwd()
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	repo := NewParametersRepository(db)

	// Update parameter
	err = repo.Save("probability_threshold", 0.85)
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify update
	param, err := repo.GetByName("probability_threshold")
	if err != nil {
		t.Fatalf("get after save: %v", err)
	}

	if param.Value != 0.85 {
		t.Errorf("expected 0.85, got %v", param.Value)
	}
}

func TestParametersRepository_SaveHistory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	wd, _ := os.Getwd()
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	repo := NewParametersRepository(db)

	// Save with history
	err = repo.SaveWithReason("probability_threshold", 0.85, "learning_adjustment")
	if err != nil {
		t.Fatalf("save with reason: %v", err)
	}

	// Get history
	history, err := repo.GetHistory("probability_threshold", 10)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}

	if len(history) < 1 {
		t.Error("expected at least 1 history entry")
	}

	// Most recent entry should have new value
	if history[0].NewValue != 0.85 {
		t.Errorf("expected new value 0.85, got %v", history[0].NewValue)
	}
	if history[0].Reason != "learning_adjustment" {
		t.Errorf("expected reason 'learning_adjustment', got %v", history[0].Reason)
	}
}

func TestParametersRepository_GetLastAdjustmentTime(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	wd, _ := os.Getwd()
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	repo := NewParametersRepository(db)

	// No adjustments yet - should return zero time
	lastTime, err := repo.GetLastAdjustmentTime("probability_threshold")
	if err != nil {
		t.Fatalf("get last adjustment time: %v", err)
	}
	if !lastTime.IsZero() {
		t.Errorf("expected zero time for no adjustments, got %v", lastTime)
	}

	// Make an adjustment
	err = repo.SaveWithReason("probability_threshold", 0.85, "learning_adjustment")
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	// Now should have a time
	lastTime, err = repo.GetLastAdjustmentTime("probability_threshold")
	if err != nil {
		t.Fatalf("get last adjustment time after save: %v", err)
	}

	// Should be recent (within last minute)
	if time.Since(lastTime) > time.Minute {
		t.Errorf("expected recent adjustment time, got %v", lastTime)
	}
}
