package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenDB_CreatesDatabase(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "persistence_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}

	// Verify WAL mode is enabled
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("check journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected WAL mode, got %s", journalMode)
	}
}

func TestRunMigrations_CreatesSchemaVersion(t *testing.T) {
	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "persistence_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	migrationsDir := filepath.Join(tmpDir, "migrations")

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("create migrations dir: %v", err)
	}

	// Create a test migration
	migration := `CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);`
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_test.sql"), []byte(migration), 0644); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	// Verify schema_version table exists
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='schema_version'`).Scan(&tableName)
	if err != nil {
		t.Errorf("schema_version table not found: %v", err)
	}

	// Verify test_table was created
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'`).Scan(&tableName)
	if err != nil {
		t.Errorf("test_table not created: %v", err)
	}

	// Verify version was recorded
	var version int
	if err := db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version); err != nil {
		t.Fatalf("query version: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
}

func TestRunMigrations_SkipsAppliedMigrations(t *testing.T) {
	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "persistence_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	migrationsDir := filepath.Join(tmpDir, "migrations")

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("create migrations dir: %v", err)
	}

	// Create first migration
	migration1 := `CREATE TABLE table1 (id INTEGER PRIMARY KEY);`
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_first.sql"), []byte(migration1), 0644); err != nil {
		t.Fatalf("write migration 1: %v", err)
	}

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	// Run first migration
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("RunMigrations (1): %v", err)
	}

	// Add second migration
	migration2 := `CREATE TABLE table2 (id INTEGER PRIMARY KEY);`
	if err := os.WriteFile(filepath.Join(migrationsDir, "002_second.sql"), []byte(migration2), 0644); err != nil {
		t.Fatalf("write migration 2: %v", err)
	}

	// Run migrations again (should only apply migration 2)
	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("RunMigrations (2): %v", err)
	}

	// Verify both tables exist
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('table1', 'table2')`).Scan(&count)
	if err != nil {
		t.Fatalf("count tables: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tables, got %d", count)
	}

	// Verify version is 2
	var version int
	if err := db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version); err != nil {
		t.Fatalf("query version: %v", err)
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
}
