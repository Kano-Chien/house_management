package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	applied := map[int]bool{}
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		rows.Scan(&v)
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, filename := range files {
		var version int
		fmt.Sscanf(filename, "%d_", &version)

		if applied[version] {
			continue
		}

		content, err := migrationsFS.ReadFile(path.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		for _, stmt := range splitSQL(string(content)) {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("execute migration %s: %w\nSQL: %s", filename, err, stmt)
			}
		}

		if _, err := db.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", version, filename); err != nil {
			return fmt.Errorf("record migration %s: %w", filename, err)
		}
		log.Printf("Applied migration: %s", filename)
	}

	return nil
}

// splitSQL splits a SQL file into individual statements, ignoring empty lines and comments.
func splitSQL(content string) []string {
	var stmts []string
	for _, s := range strings.Split(content, ";") {
		var lines []string
		for _, line := range strings.Split(s, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				lines = append(lines, line)
			}
		}
		stmt := strings.TrimSpace(strings.Join(lines, "\n"))
		if stmt != "" {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}
