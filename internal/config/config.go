package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Hierarchy HierarchyConfig `yaml:"hierarchy"`
}

type HierarchyConfig struct {
	DeviceTypes     map[string]int    `yaml:"device_types"`
	NamingRules     []NamingRule      `yaml:"naming_rules"`
	ManualOverrides map[string]string `yaml:"manual_overrides"`
}

type NamingRule struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = "config/hierarchy.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
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

func GetDefaultConfigPath() string {
	if path := os.Getenv("TOPOLOGY_CONFIG_PATH"); path != "" {
		return path
	}
	
	wd, _ := os.Getwd()
	return filepath.Join(wd, "config", "hierarchy.yaml")
}