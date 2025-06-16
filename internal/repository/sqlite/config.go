package sqlite

import (
	"fmt"
	"path/filepath"
)

// Config represents SQLite database configuration
type Config struct {
	Path string `yaml:"path"`
}

// Validate checks if the SQLite configuration is valid
func (c *Config) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("sqlite path is required")
	}

	// Special case for in-memory database
	if c.Path == ":memory:" {
		return nil
	}

	// Validate file path
	if !filepath.IsAbs(c.Path) && c.Path != ":memory:" {
		// Convert to absolute path for consistency
		absPath, err := filepath.Abs(c.Path)
		if err != nil {
			return fmt.Errorf("invalid sqlite path: %w", err)
		}
		c.Path = absPath
	}

	return nil
}

// DSN returns the SQLite connection string
func (c *Config) DSN() string {
	if c.Path == ":memory:" {
		return ":memory:"
	}
	return c.Path
}
