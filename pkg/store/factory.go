package store

import (
	"os"
)

// New creates a Store based on environment configuration.
// Supported backends are "memory" (default) and "sqlite".
func New() (Store, error) {
	switch os.Getenv("STORE_BACKEND") {
	case "sqlite":
		dsn := os.Getenv("STORE_SQLITE_DSN")
		if dsn == "" {
			dsn = "file:authorization.db?_foreign_keys=on"
		}
		return NewSQLite(dsn)
	default:
		return NewMemory(), nil
	}
}
