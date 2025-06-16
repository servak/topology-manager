package repository

import (
	"fmt"

	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository/postgres"
	"github.com/servak/topology-manager/internal/repository/sqlite"
)

// Config represents database configuration
type Config struct {
	Type     string          `yaml:"type"` // "postgres" or "sqlite"
	Postgres postgres.Config `yaml:"postgres"`
	SQLite   sqlite.Config   `yaml:"sqlite"`
}

// Repository represents a combined repository interface
type Repository interface {
	topology.Repository
	classification.Repository
	Migrate() error
	Clear() error
}

// NewRepository creates a new repository based on configuration
func NewRepository(config Config) (Repository, error) {
	switch config.Type {
	case "postgres":
		if err := config.Postgres.Validate(); err != nil {
			return nil, fmt.Errorf("invalid postgres config: %w", err)
		}
		return postgres.NewPostgresRepository(config.Postgres)
	case "sqlite":
		if err := config.SQLite.Validate(); err != nil {
			return nil, fmt.Errorf("invalid sqlite config: %w", err)
		}
		return sqlite.NewSQliteRepository(config.SQLite)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// NewTestRepository creates an in-memory SQLite repository for testing
func NewTestRepository() (Repository, error) {
	config := sqlite.Config{
		Path: ":memory:",
	}
	repo, err := sqlite.NewSQliteRepository(config)
	if err != nil {
		return nil, err
	}

	// Run migrations for test database
	if err := repo.Migrate(); err != nil {
		repo.Close()
		return nil, fmt.Errorf("failed to migrate test database: %w", err)
	}

	return repo, nil
}

// NewPostgresRepository creates a PostgreSQL repository (backward compatibility)
func NewPostgresRepository(dsn string) (Repository, error) {
	// Parse DSN to extract connection parameters (simplified)
	config := postgres.Config{
		DSN: dsn,
	}
	return postgres.NewPostgresRepository(config)
}

// NewSQLiteRepository creates a SQLite repository
func NewSQLiteRepository(path string) (Repository, error) {
	config := sqlite.Config{
		Path: path,
	}
	return sqlite.NewSQliteRepository(config)
}
