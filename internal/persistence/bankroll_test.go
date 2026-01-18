package persistence

import (
	"os"
	"testing"
)

func TestBankrollRepository_Get(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_bankroll_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewBankrollRepository(db)

	// Test: Get bankroll for polymarket (should be initialized from migrations)
	bankroll, err := repo.Get("polymarket")
	if err != nil {
		t.Fatalf("failed to get bankroll: %v", err)
	}

	if bankroll == nil {
		t.Fatal("expected bankroll, got nil")
	}

	if bankroll.InitialAmount != 50.0 {
		t.Errorf("expected initial amount 50.0, got %f", bankroll.InitialAmount)
	}

	if bankroll.CurrentAmount != 50.0 {
		t.Errorf("expected current amount 50.0, got %f", bankroll.CurrentAmount)
	}
}

func TestBankrollRepository_Update(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_bankroll_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewBankrollRepository(db)

	// Test: Update bankroll after a winning trade
	err = repo.Update("polymarket", 55.0)
	if err != nil {
		t.Fatalf("failed to update bankroll: %v", err)
	}

	// Verify update
	bankroll, _ := repo.Get("polymarket")
	if bankroll.CurrentAmount != 55.0 {
		t.Errorf("expected current amount 55.0, got %f", bankroll.CurrentAmount)
	}

	// Initial amount should remain unchanged
	if bankroll.InitialAmount != 50.0 {
		t.Errorf("expected initial amount 50.0, got %f", bankroll.InitialAmount)
	}
}

func TestBankrollRepository_Initialize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_bankroll_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewBankrollRepository(db)

	// Test: Initialize a new platform bankroll
	err = repo.Initialize("newplatform", 100.0)
	if err != nil {
		t.Fatalf("failed to initialize bankroll: %v", err)
	}

	// Verify initialization
	bankroll, _ := repo.Get("newplatform")
	if bankroll == nil {
		t.Fatal("expected bankroll for newplatform")
	}

	if bankroll.InitialAmount != 100.0 {
		t.Errorf("expected initial amount 100.0, got %f", bankroll.InitialAmount)
	}

	if bankroll.CurrentAmount != 100.0 {
		t.Errorf("expected current amount 100.0, got %f", bankroll.CurrentAmount)
	}
}

func TestBankrollRepository_GetAll(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_bankroll_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewBankrollRepository(db)

	// Test: Get all bankrolls (should have polymarket and kalshi from migrations)
	bankrolls, err := repo.GetAll()
	if err != nil {
		t.Fatalf("failed to get all bankrolls: %v", err)
	}

	if len(bankrolls) < 2 {
		t.Errorf("expected at least 2 bankrolls, got %d", len(bankrolls))
	}

	// Check both platforms exist
	platforms := make(map[string]bool)
	for _, b := range bankrolls {
		platforms[b.Platform] = true
	}

	if !platforms["polymarket"] {
		t.Error("expected polymarket in bankrolls")
	}
	if !platforms["kalshi"] {
		t.Error("expected kalshi in bankrolls")
	}
}

func TestBankrollRepository_AddToBalance(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_bankroll_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := OpenDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewBankrollRepository(db)

	// Test: Add profit to bankroll
	err = repo.AddToBalance("polymarket", 5.0) // Win $5
	if err != nil {
		t.Fatalf("failed to add to balance: %v", err)
	}

	bankroll, _ := repo.Get("polymarket")
	if bankroll.CurrentAmount != 55.0 {
		t.Errorf("expected current amount 55.0, got %f", bankroll.CurrentAmount)
	}

	// Test: Subtract loss from bankroll
	err = repo.AddToBalance("polymarket", -10.0) // Lose $10
	if err != nil {
		t.Fatalf("failed to subtract from balance: %v", err)
	}

	bankroll, _ = repo.Get("polymarket")
	if bankroll.CurrentAmount != 45.0 {
		t.Errorf("expected current amount 45.0, got %f", bankroll.CurrentAmount)
	}
}

