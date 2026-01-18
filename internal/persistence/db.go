package persistence

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDB opens a SQLite database with WAL mode enabled.
func OpenDB(path string) (*sql.DB, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return db, nil
}

// RunMigrations executes all SQL migration files in order.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Create schema_version table if not exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_version table: %w", err)
	}

	// Get current version
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Filter and sort SQL files
	var migrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrations = append(migrations, entry.Name())
		}
	}
	sort.Strings(migrations)

	// Apply pending migrations
	for _, filename := range migrations {
		// Extract version number from filename (e.g., "001_initial.sql" -> 1)
		var version int
		if _, err := fmt.Sscanf(filename, "%d_", &version); err != nil {
			continue // Skip files without version prefix
		}

		if version <= currentVersion {
			continue // Already applied
		}

		// Read and execute migration
		path := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", filename, err)
		}

		// Record migration
		if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("record migration %s: %w", filename, err)
		}
	}

	return nil
}
