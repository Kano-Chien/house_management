package database

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRunMigrations_Fresh(t *testing.T) {
	db := newTestDB(t)

	if err := RunMigrations(db); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 migrations recorded, got %d", count)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db := newTestDB(t)

	if err := RunMigrations(db); err != nil {
		t.Fatalf("first run: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("second run: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 migrations recorded after two runs, got %d", count)
	}
}

func TestRunMigrations_Partial(t *testing.T) {
	db := newTestDB(t)

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}
	_, err = db.Exec("INSERT INTO schema_migrations (version, name) VALUES (1, '001_initial_schema.up.sql')")
	if err != nil {
		t.Fatalf("insert fake migration: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 total migrations, got %d", count)
	}
}
