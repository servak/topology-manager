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
# Device classification and hierarchy management system

database:
  # Database type: postgres
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

# Environment Variable Examples:
# export DB_HOST=production-db.example.com
# export DB_PASSWORD=secure-password-from-vault
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