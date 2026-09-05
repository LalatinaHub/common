// Package database provides a thread-safe connection pool manager for Turso LibSQL.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var (
	dbInstance *sql.DB
	once       sync.Once
	initErr    error
	mu         sync.Mutex
)

// GetDB returns a singleton database connection pool.
// Connection is lazily initialized on first call with robust pool configuration.
func GetDB() (*sql.DB, error) {
	mu.Lock()
	defer mu.Unlock()

	if dbInstance != nil {
		return dbInstance, nil
	}

	once.Do(func() {
		dbURL := os.Getenv("TURSO_DATABASE_URL")
		if dbURL == "" {
			initErr = fmt.Errorf("TURSO_DATABASE_URL environment variable not set")
			return
		}

		// Support separate TURSO_AUTH_TOKEN environment variable
		if authToken := os.Getenv("TURSO_AUTH_TOKEN"); authToken != "" && !strings.Contains(dbURL, "authToken=") {
			sep := "?"
			if strings.Contains(dbURL, "?") {
				sep = "&"
			}
			dbURL = fmt.Sprintf("%s%sauthToken=%s", dbURL, sep, authToken)
		}

		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			initErr = fmt.Errorf("failed to open database: %w", err)
			return
		}

		// Configure connection pool for high concurrency and resilience
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetConnMaxIdleTime(1 * time.Minute)

		// Verify connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			initErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		dbInstance = db
	})

	return dbInstance, initErr
}

// SetDB explicitly overrides the singleton DB instance (e.g. for testing or custom pool configuration).
func SetDB(db *sql.DB) {
	mu.Lock()
	defer mu.Unlock()
	dbInstance = db
	initErr = nil
}

// ResetDBInstance resets the singleton state so GetDB can be initialized again.
func ResetDBInstance() {
	mu.Lock()
	defer mu.Unlock()
	if dbInstance != nil {
		_ = dbInstance.Close()
	}
	dbInstance = nil
	initErr = nil
	once = sync.Once{}
}

// Ping verifies that the database connection is alive.
func Ping(ctx context.Context) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

// InitIndexes ensures critical performance indexes exist on tables.
func InitIndexes(ctx context.Context) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_expired ON users(expired);`,
		`CREATE INDEX IF NOT EXISTS idx_users_quota ON users(quota);`,
		`CREATE INDEX IF NOT EXISTS idx_users_server_code ON users(server_code);`,
		`CREATE INDEX IF NOT EXISTS idx_users_vpn ON users(vpn);`,
	}

	for _, idx := range indexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("failed to create index (%s): %w", idx, err)
		}
	}

	return nil
}

// Close closes the database connection pool.
// Should only be called during graceful shutdown.
func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if dbInstance != nil {
		err := dbInstance.Close()
		dbInstance = nil
		return err
	}
	return nil
}
