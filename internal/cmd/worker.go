package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/prometheus"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/repository/postgres"
	"github.com/servak/topology-manager/internal/worker"
	"github.com/spf13/cobra"
)

var (
	// Worker flags
	workerInterval     int
	prometheusURL      string
	deviceInterval     int
	cleanupInterval    int
	batchSize          int
	syncTimeout        int
	enableLLDPSync     bool
	enableDeviceSync   bool
	enableCleanup      bool
	enableAutoClassify bool
	maxDeviceAge       int
	maxLinkAge         int
	prometheusTimeout  int
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run topology synchronization worker",
	Long:  `Run background worker to synchronize network topology data from Prometheus`,
	RunE:  runWorker,
}

func init() {
	// Required flags
	workerCmd.Flags().IntVarP(&workerInterval, "interval", "i", 300, "LLDP sync interval in seconds")
	workerCmd.Flags().StringVarP(&prometheusURL, "prometheus-url", "p", "http://localhost:9090", "Prometheus server URL")

	// Optional flags with defaults
	workerCmd.Flags().IntVar(&deviceInterval, "device-interval", 600, "Device info sync interval in seconds")
	workerCmd.Flags().IntVar(&cleanupInterval, "cleanup-interval", 3600, "Data cleanup interval in seconds")
	workerCmd.Flags().IntVar(&batchSize, "batch-size", 100, "Batch size for bulk operations")
	workerCmd.Flags().IntVar(&syncTimeout, "sync-timeout", 600, "Sync operation timeout in seconds")
	workerCmd.Flags().IntVar(&prometheusTimeout, "prometheus-timeout", 30, "Prometheus query timeout in seconds")
	workerCmd.Flags().IntVar(&maxDeviceAge, "max-device-age", 86400, "Maximum device age in seconds before cleanup")
	workerCmd.Flags().IntVar(&maxLinkAge, "max-link-age", 43200, "Maximum link age in seconds before cleanup")

	// Feature toggles
	workerCmd.Flags().BoolVar(&enableLLDPSync, "enable-lldp", true, "Enable LLDP topology synchronization")
	workerCmd.Flags().BoolVar(&enableDeviceSync, "enable-device", true, "Enable device info synchronization")
	workerCmd.Flags().BoolVar(&enableCleanup, "enable-cleanup", true, "Enable old data cleanup")
	workerCmd.Flags().BoolVar(&enableAutoClassify, "enable-auto-classify", true, "Enable automatic device classification")

	// Add to root command
	rootCmd.AddCommand(workerCmd)
}

func runWorker(cmd *cobra.Command, args []string) error {
	logger := log.New(os.Stdout, "[WORKER] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting topology synchronization worker...")

	// Load base configuration for database settings
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override Prometheus settings from CLI flags
	cfg.Prometheus.URL = prometheusURL
	cfg.Prometheus.Timeout = time.Duration(prometheusTimeout) * time.Second

	// Create worker configuration from CLI flags
	workerConfig := worker.PrometheusSyncConfig{
		LLDPSyncInterval:   time.Duration(workerInterval) * time.Second,
		DeviceSyncInterval: time.Duration(deviceInterval) * time.Second,
		CleanupInterval:    time.Duration(cleanupInterval) * time.Second,
		EnableLLDPSync:     enableLLDPSync,
		EnableDeviceSync:   enableDeviceSync,
		EnableCleanup:      enableCleanup,
		EnableAutoClassify: enableAutoClassify,
		MaxDeviceAge:       time.Duration(maxDeviceAge) * time.Second,
		MaxLinkAge:         time.Duration(maxLinkAge) * time.Second,
		BatchSize:          batchSize,
		SyncTimeout:        time.Duration(syncTimeout) * time.Second,
	}

	// Validate worker configuration
	if err := validateWorkerConfig(workerConfig); err != nil {
		return fmt.Errorf("invalid worker configuration: %w", err)
	}

	// Create database repository
	repo, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer repo.Close()

	// Test database connection
	ctx := context.Background()
	if err := repo.Health(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	logger.Printf("Connected to %s database", cfg.Database.Type)

	// Create Prometheus client
	promClient := prometheus.NewClient(cfg.GetPrometheusConfig())

	// Test Prometheus connection
	if err := promClient.Health(ctx); err != nil {
		return fmt.Errorf("prometheus health check failed: %w", err)
	}
	logger.Printf("Connected to Prometheus at %s", prometheusURL)

	// Load configuration for metrics mapping
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create classification repository if auto-classification is enabled
	var classificationRepo classification.Repository
	if enableAutoClassify {
		// PostgreSQL specific implementation
		pgRepo, ok := repo.(*postgres.PostgresRepository)
		if !ok {
			return fmt.Errorf("auto-classification requires PostgreSQL, got %T", repo)
		}
		classificationRepo = postgres.NewClassificationRepository(pgRepo.GetDB())
		logger.Println("Auto-classification enabled")
	}

	// Create and start worker
	worker := worker.NewPrometheusSync(promClient, appConfig.GetMetricsConfig(), repo, classificationRepo, workerConfig, logger)

	if err := worker.Start(); err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}
	defer worker.Stop()

	// Log worker configuration
	logWorkerConfig(logger, workerConfig)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Println("Worker started successfully. Press Ctrl+C to stop.")

	// Block until signal received
	sig := <-sigChan
	logger.Printf("Received signal %s, shutting down...", sig)

	return nil
}

func validateWorkerConfig(config worker.PrometheusSyncConfig) error {
	if config.LLDPSyncInterval <= 0 {
		return fmt.Errorf("LLDP sync interval must be positive")
	}
	if config.DeviceSyncInterval <= 0 {
		return fmt.Errorf("device sync interval must be positive")
	}
	if config.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup interval must be positive")
	}
	if config.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}
	if config.SyncTimeout <= 0 {
		return fmt.Errorf("sync timeout must be positive")
	}
	if config.MaxDeviceAge <= 0 {
		return fmt.Errorf("max device age must be positive")
	}
	if config.MaxLinkAge <= 0 {
		return fmt.Errorf("max link age must be positive")
	}

	// Sanity checks
	if config.LLDPSyncInterval < 30*time.Second {
		return fmt.Errorf("LLDP sync interval too short (minimum 30 seconds)")
	}
	if config.DeviceSyncInterval < 60*time.Second {
		return fmt.Errorf("device sync interval too short (minimum 60 seconds)")
	}
	if config.CleanupInterval < 600*time.Second {
		return fmt.Errorf("cleanup interval too short (minimum 10 minutes)")
	}

	return nil
}

func logWorkerConfig(logger *log.Logger, config worker.PrometheusSyncConfig) {
	logger.Println("Worker Configuration:")
	logger.Printf("  LLDP Sync Interval: %s (enabled: %t)", config.LLDPSyncInterval, config.EnableLLDPSync)
	logger.Printf("  Device Sync Interval: %s (enabled: %t)", config.DeviceSyncInterval, config.EnableDeviceSync)
	logger.Printf("  Cleanup Interval: %s (enabled: %t)", config.CleanupInterval, config.EnableCleanup)
	logger.Printf("  Auto-Classification: enabled: %t", config.EnableAutoClassify)
	logger.Printf("  Batch Size: %d", config.BatchSize)
	logger.Printf("  Sync Timeout: %s", config.SyncTimeout)
	logger.Printf("  Max Device Age: %s", config.MaxDeviceAge)
	logger.Printf("  Max Link Age: %s", config.MaxLinkAge)
}