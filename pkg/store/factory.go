package store

import (
	"os"
)

// New creates a Store based on environment configuration.
// Supported backends are "memory" (default), "sqlite" and "postgres".
func New() (Store, error) {
	switch os.Getenv("STORE_BACKEND") {
	case "sqlite":
		dsn := os.Getenv("STORE_SQLITE_DSN")
		if dsn == "" {
			dsn = "file:authorization.db?_foreign_keys=on"
		}
		return NewSQLite(dsn)
	case "postgres":
		dsn := os.Getenv("STORE_PG_DSN")
		if dsn == "" {
			dsn = "postgres://postgres:postgres@localhost:5432/authz?sslmode=disable"
		}
		return NewPostgres(dsn)
	default:
		return NewMemory(), nil
	}
}
