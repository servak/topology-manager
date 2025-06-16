package postgres

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Config represents PostgreSQL database configuration
type Config struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	DSN      string `yaml:"dsn"` // Direct DSN string (takes precedence)
}

// BuildDSN returns the PostgreSQL connection string
func (c *Config) BuildDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// ParseDSN parses a PostgreSQL DSN and populates the config fields
func (c *Config) ParseDSN(dsn string) error {
	c.DSN = dsn

	// Try to parse the DSN to extract individual components
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("invalid PostgreSQL DSN: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("invalid PostgreSQL scheme: %s", u.Scheme)
	}

	// Extract host and port
	c.Host = u.Hostname()
	if portStr := u.Port(); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			c.Port = port
		}
	}

	// Extract user and password
	if u.User != nil {
		c.User = u.User.Username()
		if password, ok := u.User.Password(); ok {
			c.Password = password
		}
	}

	// Extract database name
	if len(u.Path) > 1 {
		c.DBName = strings.TrimPrefix(u.Path, "/")
	}

	// Extract SSL mode from query parameters
	if sslMode := u.Query().Get("sslmode"); sslMode != "" {
		c.SSLMode = sslMode
	}

	return nil
}

// Validate checks if the PostgreSQL configuration is valid
func (c *Config) Validate() error {
	// If DSN is provided, try to parse it
	if c.DSN != "" {
		return c.ParseDSN(c.DSN)
	}

	// Otherwise validate individual fields
	if c.Host == "" {
		return fmt.Errorf("postgres host is required")
	}
	if c.User == "" {
		return fmt.Errorf("postgres user is required")
	}
	if c.Password == "" {
		return fmt.Errorf("postgres password is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("postgres database name is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("postgres port must be between 1 and 65535, got %d", c.Port)
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable" // Default value
	}

	// Validate SSL mode
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	validMode := false
	for _, mode := range validSSLModes {
		if c.SSLMode == mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return fmt.Errorf("invalid SSL mode: %s (valid: %s)", c.SSLMode, strings.Join(validSSLModes, ", "))
	}

	return nil
}
