package cmd

import (
	"fmt"
	"os"

	"github.com/servak/topology-manager/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Configuration management commands for the topology manager`,
}

// configValidateCmd validates the configuration file
var configValidateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate configuration file",
	Long:  `Validate the syntax and content of a configuration file`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.GetDefaultConfigPath()
		if len(args) > 0 {
			configPath = args[0]
		}

		fmt.Printf("Validating configuration file: %s\n", configPath)

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		fmt.Println("✅ Configuration is valid")
		fmt.Printf("Database type: %s\n", cfg.Database.Type)
		fmt.Printf("Device types: %d configured\n", len(cfg.Hierarchy.DeviceTypes))
		fmt.Printf("Naming rules: %d configured\n", len(cfg.Hierarchy.NamingRules))
		fmt.Printf("Manual overrides: %d configured\n", len(cfg.Hierarchy.ManualOverrides))

		return nil
	},
}

// configShowCmd shows the current configuration
var configShowCmd = &cobra.Command{
	Use:   "show [config-file]",
	Short: "Show current configuration",
	Long:  `Display the current configuration with all default values applied`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.GetDefaultConfigPath()
		if len(args) > 0 {
			configPath = args[0]
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Mask sensitive information
		if cfg.Database.Postgres != nil && cfg.Database.Postgres.Password != "" {
			cfg.Database.Postgres.Password = "***"
		}
		if cfg.Database.Neo4j != nil && cfg.Database.Neo4j.Password != "" {
			cfg.Database.Neo4j.Password = "***"
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to serialize configuration: %w", err)
		}

		fmt.Printf("Configuration from: %s\n\n", configPath)
		fmt.Print(string(data))

		return nil
	},
}

// configExampleCmd generates an example configuration file
var configExampleCmd = &cobra.Command{
	Use:   "example [output-file]",
	Short: "Generate example configuration file",
	Long:  `Generate an example configuration file with comments and default values`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath := "tm.example.yaml"
		if len(args) > 0 {
			outputPath = args[0]
		}

		exampleConfig := `# Network Topology Manager Configuration
# This is an example configuration file with all available options

hierarchy:
  # Device type hierarchy levels (lower numbers = higher in hierarchy)
  device_types:
    core: 1
    distribution: 2
    access: 3
    server: 4
    unknown: 99
  
  # Automatic device type detection rules based on naming patterns
  naming_rules:
    - pattern: "^core-.*"
      type: "core"
    - pattern: "^dist-.*"
      type: "distribution" 
    - pattern: "^access-.*"
      type: "access"
    - pattern: "^server-.*"
      type: "server"
  
  # Manual device type overrides
  manual_overrides:
    # "special-device-001": "core"

database:
  # Database type: postgres or neo4j
  type: postgres
  
  # PostgreSQL configuration
  # Environment variables are supported: ${VAR} or ${VAR:default}
  postgres:
    host: ${DB_HOST:localhost}              # Environment: DB_HOST
    port: 5432
    user: ${DB_USER:tm}                     # Environment: DB_USER
    password: ${DB_PASSWORD:tm_password}    # Environment: DB_PASSWORD
    dbname: ${DB_NAME:topology_manager}     # Environment: DB_NAME
    sslmode: ${DB_SSLMODE:disable}          # Environment: DB_SSLMODE
  
  # Neo4j configuration (when type: neo4j)
  # Environment variables are supported: ${VAR} or ${VAR:default}
  neo4j:
    uri: ${NEO4J_URI:bolt://localhost:7687}       # Environment: NEO4J_URI
    username: ${NEO4J_USERNAME:neo4j}             # Environment: NEO4J_USERNAME
    password: ${NEO4J_PASSWORD:neo4j_password}    # Environment: NEO4J_PASSWORD
    database: ${NEO4J_DATABASE:neo4j}             # Environment: NEO4J_DATABASE

# Environment Variable Examples:
# export DB_HOST=production-db.example.com
# export DB_PASSWORD=secure-password-from-vault
# export NEO4J_PASSWORD=secure-neo4j-password
`

		err := os.WriteFile(outputPath, []byte(exampleConfig), 0644)
		if err != nil {
			return fmt.Errorf("failed to write example config: %w", err)
		}

		fmt.Printf("Example configuration written to: %s\n", outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configExampleCmd)
}