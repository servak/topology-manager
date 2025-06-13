package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/repository/postgres"
	"gopkg.in/yaml.v3"
)

// Config represents the main application configuration
type Config struct {
	Hierarchy HierarchyConfig           `yaml:"hierarchy"`
	Database  repository.DatabaseConfig `yaml:"database"`
}

// HierarchyConfig holds device hierarchy configuration
type HierarchyConfig struct {
	DeviceTypes     map[string]int    `yaml:"device_types"`
	NamingRules     []NamingRule      `yaml:"naming_rules"`
	ManualOverrides map[string]string `yaml:"manual_overrides"`
}

// NamingRule defines a pattern-based device type detection rule
type NamingRule struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	
	// Set defaults
	config.setDefaults()
	
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand environment variables
	config.expandEnvironmentVariables()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// setDefaults sets default values for the configuration
func (c *Config) setDefaults() {
	// Hierarchy defaults
	if c.Hierarchy.DeviceTypes == nil {
		c.Hierarchy.DeviceTypes = map[string]int{
			"core":         1,
			"distribution": 2,
			"access":       3,
			"server":       4,
			"unknown":      99,
		}
	}
	if c.Hierarchy.NamingRules == nil {
		c.Hierarchy.NamingRules = []NamingRule{
			{Pattern: "^core-.*", Type: "core"},
			{Pattern: "^dist-.*", Type: "distribution"},
			{Pattern: "^access-.*", Type: "access"},
			{Pattern: "^server-.*", Type: "server"},
		}
	}
	if c.Hierarchy.ManualOverrides == nil {
		c.Hierarchy.ManualOverrides = make(map[string]string)
	}

	// Database defaults
	if c.Database.Type == "" {
		c.Database.Type = "postgres"
	}
	if c.Database.Type == "postgres" && c.Database.Postgres == nil {
		c.Database.Postgres = &postgres.PostgresConfig{
			Host:    "localhost",
			Port:    5432,
			User:    "topology",
			DBName:  "topology_manager",
			SSLMode: "disable",
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate hierarchy
	if err := c.validateHierarchy(); err != nil {
		return fmt.Errorf("hierarchy configuration error: %w", err)
	}

	// Validate database
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database configuration error: %w", err)
	}

	return nil
}

// validateHierarchy validates the hierarchy configuration
func (c *Config) validateHierarchy() error {
	if len(c.Hierarchy.DeviceTypes) == 0 {
		return fmt.Errorf("device_types cannot be empty")
	}

	// Validate naming rules
	for i, rule := range c.Hierarchy.NamingRules {
		if rule.Pattern == "" {
			return fmt.Errorf("naming rule %d: pattern cannot be empty", i)
		}
		if rule.Type == "" {
			return fmt.Errorf("naming rule %d: type cannot be empty", i)
		}
		
		// Validate regex pattern
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("naming rule %d: invalid regex pattern '%s': %w", i, rule.Pattern, err)
		}
		
		// Check if type exists in device types
		if _, exists := c.Hierarchy.DeviceTypes[rule.Type]; !exists {
			return fmt.Errorf("naming rule %d: type '%s' not found in device_types", i, rule.Type)
		}
	}

	// Validate manual overrides
	for device, deviceType := range c.Hierarchy.ManualOverrides {
		if device == "" {
			return fmt.Errorf("manual override: device name cannot be empty")
		}
		if deviceType == "" {
			return fmt.Errorf("manual override for device '%s': type cannot be empty", device)
		}
		if _, exists := c.Hierarchy.DeviceTypes[deviceType]; !exists {
			return fmt.Errorf("manual override for device '%s': type '%s' not found in device_types", device, deviceType)
		}
	}

	return nil
}

// expandEnvironmentVariables expands environment variables in configuration values
func (c *Config) expandEnvironmentVariables() {
	// Expand database configuration
	if c.Database.Postgres != nil {
		c.Database.Postgres.Host = expandEnvVar(c.Database.Postgres.Host)
		c.Database.Postgres.User = expandEnvVar(c.Database.Postgres.User)
		c.Database.Postgres.Password = expandEnvVar(c.Database.Postgres.Password)
		c.Database.Postgres.DBName = expandEnvVar(c.Database.Postgres.DBName)
		c.Database.Postgres.SSLMode = expandEnvVar(c.Database.Postgres.SSLMode)
	}

	if c.Database.Neo4j != nil {
		c.Database.Neo4j.URI = expandEnvVar(c.Database.Neo4j.URI)
		c.Database.Neo4j.Username = expandEnvVar(c.Database.Neo4j.Username)
		c.Database.Neo4j.Password = expandEnvVar(c.Database.Neo4j.Password)
		c.Database.Neo4j.Database = expandEnvVar(c.Database.Neo4j.Database)
	}
}

// expandEnvVar expands environment variables in a string
// Supports ${VAR}, $VAR formats with optional default values ${VAR:default}
func expandEnvVar(s string) string {
	if s == "" {
		return s
	}

	// Handle ${VAR:default} format
	if strings.Contains(s, "${") && strings.Contains(s, "}") {
		return os.Expand(s, func(key string) string {
			// Handle default values: VAR:default
			if colonIndex := strings.Index(key, ":"); colonIndex != -1 {
				envKey := key[:colonIndex]
				defaultValue := key[colonIndex+1:]
				if value := os.Getenv(envKey); value != "" {
					return value
				}
				return defaultValue
			}
			return os.Getenv(key)
		})
	}

	// Handle simple $VAR format
	if strings.HasPrefix(s, "$") {
		envKey := s[1:]
		if value := os.Getenv(envKey); value != "" {
			return value
		}
	}

	return s
}

func (c *Config) GetDeviceType(deviceName string) (string, int, error) {
	if override, exists := c.Hierarchy.ManualOverrides[deviceName]; exists {
		if level, exists := c.Hierarchy.DeviceTypes[override]; exists {
			return override, level, nil
		}
	}

	for _, rule := range c.Hierarchy.NamingRules {
		matched, err := regexp.MatchString(rule.Pattern, deviceName)
		if err != nil {
			return "", 0, fmt.Errorf("invalid regex pattern %s: %w", rule.Pattern, err)
		}
		if matched {
			if level, exists := c.Hierarchy.DeviceTypes[rule.Type]; exists {
				return rule.Type, level, nil
			}
		}
	}

	return "unknown", c.Hierarchy.DeviceTypes["unknown"], nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	// Check environment variable first
	if path := os.Getenv("TOPOLOGY_CONFIG_PATH"); path != "" {
		return path
	}

	// Check common locations
	locations := []string{
		"tm.yaml",              // Current directory
		"config/tm.yaml",       // Config subdirectory
		"topology.yaml",        // Legacy name
		"config/topology.yaml", // Legacy location
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}

	// Default fallback
	return "tm.yaml"
}
