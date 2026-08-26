package db

import (
	"database/sql"
	_ "embed"
	"strings"
)

//go:embed schema.sql
var schema string

func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// migrate applies schema changes CREATE TABLE IF NOT EXISTS can't retrofit
// onto databases that were created before the change.
func migrate(conn *sql.DB) error {
	if _, err := conn.Exec(`ALTER TABLE books ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0`); err != nil &&
		!strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

// SeedAdmin creates the initial author account if no users exist yet.
func SeedAdmin(conn *sql.DB, username, passwordHash string) error {
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := conn.Exec(`INSERT INTO users (username, password_hash, is_admin) VALUES (?, ?, 1)`, username, passwordHash)
	return err
}
