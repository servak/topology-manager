package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/servak/topology-manager/internal/prometheus"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/repository/postgres"
	"gopkg.in/yaml.v3"
)

// Config represents the main application configuration
type Config struct {
	Hierarchy  HierarchyConfig           `yaml:"hierarchy"`
	Database   repository.DatabaseConfig `yaml:"database"`
	Prometheus PrometheusConfig          `yaml:"prometheus"`
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	URL             string                         `yaml:"url"`
	Timeout         time.Duration                  `yaml:"timeout"`
	MetricsMapping  map[string]prometheus.MetricConfigGroup   `yaml:"metrics_mapping"`
	FieldRequirements map[string]prometheus.FieldRequirement `yaml:"field_requirements"`
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

	// Prometheus defaults
	if c.Prometheus.URL == "" {
		c.Prometheus.URL = "http://localhost:9090"
	}
	if c.Prometheus.Timeout == 0 {
		c.Prometheus.Timeout = 30 * time.Second
	}
	
	// Set default metrics mapping
	c.setDefaultMetricsMapping()

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

	// Validate Prometheus
	if err := c.validatePrometheus(); err != nil {
		return fmt.Errorf("prometheus configuration error: %w", err)
	}


	return nil
}

// setDefaultMetricsMapping sets default metrics mapping configuration
func (c *Config) setDefaultMetricsMapping() {
	if c.Prometheus.MetricsMapping == nil {
		c.Prometheus.MetricsMapping = map[string]prometheus.MetricConfigGroup{
			"device_info": {
				Primary: prometheus.MetricMapping{
					MetricName: "snmp_device_info",
					Labels: map[string]string{
						"device_id":  "instance",
						"ip_address": "instance",
						"hardware":   "sysDescr",
						"location":   "sysLocation",
					},
				},
				Fallbacks: []prometheus.MetricMapping{
					{
						MetricName: "node_uname_info",
						Labels: map[string]string{
							"device_id":  "instance",
							"ip_address": "instance",
							"hardware":   "machine",
							"location":   "",
						},
					},
					{
						MetricName: "lldp_local_info",
						Labels: map[string]string{
							"device_id":  "chassis_id",
							"ip_address": "mgmt_address",
							"hardware":   "system_description",
							"location":   "system_location",
						},
					},
				},
			},
			"lldp_neighbors": {
				Primary: prometheus.MetricMapping{
					MetricName: "snmp_lldp_neighbor_info",
					Labels: map[string]string{
						"source_device": "instance",
						"source_port":   "lldpLocalPortId",
						"target_device": "lldpRemSysName",
						"target_port":   "lldpRemPortId",
					},
				},
				Fallbacks: []prometheus.MetricMapping{
					{
						MetricName: "lldp_remote_info",
						Labels: map[string]string{
							"source_device": "local_chassis",
							"source_port":   "local_port_id",
							"target_device": "remote_chassis",
							"target_port":   "remote_port_id",
						},
					},
				},
			},
		}
	}

	if c.Prometheus.FieldRequirements == nil {
		c.Prometheus.FieldRequirements = map[string]prometheus.FieldRequirement{
			"device_info": {
				Required: []string{"device_id"},
				Optional: []string{"ip_address", "hardware", "location"},
			},
			"lldp_neighbors": {
				Required: []string{"source_device", "target_device"},
				Optional: []string{"source_port", "target_port"},
			},
		}
	}
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


	// Expand Prometheus configuration
	c.Prometheus.URL = expandEnvVar(c.Prometheus.URL)
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

// validatePrometheus validates Prometheus configuration
func (c *Config) validatePrometheus() error {
	if c.Prometheus.URL == "" {
		return fmt.Errorf("prometheus URL is required")
	}
	if c.Prometheus.Timeout <= 0 {
		return fmt.Errorf("prometheus timeout must be positive")
	}
	return nil
}


// GetPrometheusConfig returns Prometheus client configuration
func (c *Config) GetPrometheusConfig() prometheus.Config {
	return prometheus.Config{
		URL:     expandEnvVar(c.Prometheus.URL),
		Timeout: c.Prometheus.Timeout,
	}
}

// GetMetricsConfig returns metrics configuration for MetricsExtractor
func (c *Config) GetMetricsConfig() *prometheus.MetricsConfig {
	return &prometheus.MetricsConfig{
		MetricsMapping:    c.Prometheus.MetricsMapping,
		FieldRequirements: c.Prometheus.FieldRequirements,
	}
}
