package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "topology-manager",
	Short: "Network topology management system",
	Long: `A network topology management system that collects LLDP information 
from Prometheus and provides hierarchical topology visualization.`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("topology-manager version %s\n", rootCmd.Version)
	},
}