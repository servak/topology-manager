package sqlite

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// sqliteRepository implements both topology and classification repository interfaces
type sqliteRepository struct {
	db *sqlx.DB
}

// NewSQliteRepository creates a new SQLite repository
func NewSQliteRepository(config Config) (*sqliteRepository, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("sqlite3", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency (except for :memory:)
	if config.Path != ":memory:" {
		if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return &sqliteRepository{db: db}, nil
}

// Close closes the database connection
func (r *sqliteRepository) Close() error {
	return r.db.Close()
}

// Health checks database connectivity
func (r *sqliteRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Migrate runs database migrations
func (r *sqliteRepository) Migrate() error {
	return RunMigrations(r.db)
}

// Clear clears the database
func (r *sqliteRepository) Clear() error {
	_, err := r.db.Exec("DELETE FROM links")
	if err != nil {
		return fmt.Errorf("failed to clear links: %w", err)
	}
	_, err = r.db.Exec("DELETE FROM devices")
	if err != nil {
		return fmt.Errorf("failed to clear devices: %w", err)
	}
	_, err = r.db.Exec("DELETE FROM device_classifications")
	if err != nil {
		return fmt.Errorf("failed to clear device classifications: %w", err)
	}
	_, err = r.db.Exec("DELETE FROM classification_rules")
	if err != nil {
		return fmt.Errorf("failed to clear classification rules: %w", err)
	}
	_, err = r.db.Exec("DELETE FROM hierarchy_layers")
	if err != nil {
		return fmt.Errorf("failed to clear hierarchy layers: %w", err)
	}
	return nil
}