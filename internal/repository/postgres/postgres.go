package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// postgresRepository implements both topology and classification repository interfaces
type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(config Config) (*postgresRepository, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", config.BuildDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return &postgresRepository{db: db}, nil
}

// Close closes the database connection
func (r *postgresRepository) Close() error {
	return r.db.Close()
}

// Health checks database connectivity
func (r *postgresRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Migrate runs database migrations
func (r *postgresRepository) Migrate() error {
	// PostgreSQL migrations would go here
	return fmt.Errorf("PostgreSQL migrations not implemented yet")
}

// Clear clears the database
func (r *postgresRepository) Clear() error {
	_, err := r.db.Exec("DELETE FROM links")
	if err != nil {
		return fmt.Errorf("failed to clear links: %w", err)
	}
	_, err = r.db.Exec("DELETE FROM devices")
	return err
}