package repository

import (
	"fmt"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository/neo4j"
	"github.com/servak/topology-manager/internal/repository/postgres"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string                   `yaml:"type"`
	Postgres *postgres.PostgresConfig `yaml:"postgres,omitempty"`
	Neo4j    *neo4j.Neo4jConfig       `yaml:"neo4j,omitempty"`
}

// Validate checks if the database configuration is valid
func (c *DatabaseConfig) Validate() error {
	switch c.Type {
	case "postgres":
		if c.Postgres == nil {
			return fmt.Errorf("postgres configuration is required when type is postgres")
		}
		return c.Postgres.Validate()
	case "neo4j":
		if c.Neo4j == nil {
			return fmt.Errorf("neo4j configuration is required when type is neo4j")
		}
		return c.Neo4j.Validate()
	default:
		return fmt.Errorf("unsupported database type: %s (supported: postgres, neo4j)", c.Type)
	}
}

// NewDatabase creates a new database repository based on configuration
func NewDatabase(config *DatabaseConfig) (topology.Repository, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	switch config.Type {
	case "postgres":
		repo, err := postgres.NewPostgresRepository(config.Postgres.DSN())
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres repository: %w", err)
		}
		return repo, nil
	case "neo4j":
		// TODO: Implement Neo4j repository
		return nil, fmt.Errorf("neo4j repository not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
