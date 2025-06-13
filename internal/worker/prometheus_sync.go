package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/prometheus"
)

// PrometheusSync handles synchronization of topology data from Prometheus
type PrometheusSync struct {
	promClient       *prometheus.Client
	metricsExtractor *prometheus.MetricsExtractor
	lldpParser       *prometheus.LLDPParser
	repository       topology.Repository
	scheduler        *Scheduler
	logger           *log.Logger
	config           PrometheusSyncConfig
}

// PrometheusSyncConfig holds configuration for Prometheus synchronization
type PrometheusSyncConfig struct {
	// Collection intervals
	LLDPSyncInterval      time.Duration `yaml:"lldp_sync_interval"`
	DeviceSyncInterval    time.Duration `yaml:"device_sync_interval"`
	CleanupInterval       time.Duration `yaml:"cleanup_interval"`
	
	// Sync behavior
	EnableLLDPSync        bool `yaml:"enable_lldp_sync"`
	EnableDeviceSync      bool `yaml:"enable_device_sync"`
	EnableCleanup         bool `yaml:"enable_cleanup"`
	
	// Data management
	MaxDeviceAge          time.Duration `yaml:"max_device_age"`
	MaxLinkAge            time.Duration `yaml:"max_link_age"`
	
	// Batch settings
	BatchSize             int `yaml:"batch_size"`
	SyncTimeout           time.Duration `yaml:"sync_timeout"`
}

// DefaultPrometheusSyncConfig returns default configuration
func DefaultPrometheusSyncConfig() PrometheusSyncConfig {
	return PrometheusSyncConfig{
		LLDPSyncInterval:   5 * time.Minute,
		DeviceSyncInterval: 10 * time.Minute,
		CleanupInterval:    1 * time.Hour,
		EnableLLDPSync:     true,
		EnableDeviceSync:   true,
		EnableCleanup:      true,
		MaxDeviceAge:       24 * time.Hour,
		MaxLinkAge:         12 * time.Hour,
		BatchSize:          100,
		SyncTimeout:        10 * time.Minute,
	}
}

// NewPrometheusSync creates a new Prometheus synchronization worker
func NewPrometheusSync(
	promClient *prometheus.Client,
	metricsConfig *prometheus.MetricsConfig,
	repository topology.Repository,
	config PrometheusSyncConfig,
	logger *log.Logger,
) *PrometheusSync {
	if logger == nil {
		logger = log.Default()
	}

	metricsExtractor := prometheus.NewMetricsExtractor(promClient, metricsConfig)
	lldpParser := prometheus.NewLLDPParser(promClient)
	scheduler := NewScheduler(logger)

	return &PrometheusSync{
		promClient:       promClient,
		metricsExtractor: metricsExtractor,
		lldpParser:       lldpParser,
		repository:       repository,
		scheduler:        scheduler,
		logger:           logger,
		config:           config,
	}
}

// Start starts the Prometheus synchronization worker
func (ps *PrometheusSync) Start() error {
	ps.logger.Println("Starting Prometheus synchronization worker...")

	// Add combined topology synchronization task (devices + LLDP)
	if ps.config.EnableLLDPSync || ps.config.EnableDeviceSync {
		// Use the shorter interval for combined sync
		syncInterval := ps.config.LLDPSyncInterval
		if ps.config.EnableDeviceSync && ps.config.DeviceSyncInterval < syncInterval {
			syncInterval = ps.config.DeviceSyncInterval
		}

		topologyTask := NewTaskBuilder("topology_sync", "Complete Topology Sync").
			Description("Synchronizes devices and LLDP topology from Prometheus in proper order").
			Interval(syncInterval).
			Timeout(ps.config.SyncTimeout).
			Function(ps.syncCompleteTopology).
			Build()

		if err := ps.scheduler.AddTask(topologyTask); err != nil {
			return fmt.Errorf("failed to add topology sync task: %w", err)
		}
	}

	// Add cleanup task
	if ps.config.EnableCleanup {
		cleanupTask := NewTaskBuilder("cleanup", "Data Cleanup").
			Description("Cleans up old topology data").
			Interval(ps.config.CleanupInterval).
			Timeout(ps.config.SyncTimeout).
			Function(ps.cleanupOldData).
			Build()

		if err := ps.scheduler.AddTask(cleanupTask); err != nil {
			return fmt.Errorf("failed to add cleanup task: %w", err)
		}
	}

	// Start the scheduler
	ps.scheduler.Start()

	ps.logger.Println("Prometheus synchronization worker started successfully")
	return nil
}

// Stop stops the Prometheus synchronization worker
func (ps *PrometheusSync) Stop() {
	ps.logger.Println("Stopping Prometheus synchronization worker...")
	ps.scheduler.Stop()
	ps.logger.Println("Prometheus synchronization worker stopped")
}

// GetStatus returns the status of all synchronization tasks
func (ps *PrometheusSync) GetStatus() []TaskStatus {
	return ps.scheduler.GetTaskStatus()
}

// SyncNow triggers an immediate synchronization of all enabled tasks
func (ps *PrometheusSync) SyncNow() error {
	var errors []error

	if ps.config.EnableLLDPSync || ps.config.EnableDeviceSync {
		if err := ps.scheduler.RunTaskNow("topology_sync"); err != nil {
			errors = append(errors, fmt.Errorf("topology sync: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %v", errors)
	}

	return nil
}

// Private synchronization methods

func (ps *PrometheusSync) syncCompleteTopology(ctx context.Context) error {
	ps.logger.Println("Starting complete topology synchronization...")

	var allErrors []error

	// Step 1: Synchronize device information first
	if ps.config.EnableDeviceSync {
		ps.logger.Println("Phase 1: Synchronizing device information...")
		if err := ps.syncDeviceInfo(ctx); err != nil {
			allErrors = append(allErrors, fmt.Errorf("device sync failed: %w", err))
			ps.logger.Printf("Device sync failed, but continuing with LLDP sync: %v", err)
		} else {
			ps.logger.Println("Phase 1: Device synchronization completed successfully")
		}
	}

	// Step 2: Synchronize LLDP topology (with placeholder device creation)
	if ps.config.EnableLLDPSync {
		ps.logger.Println("Phase 2: Synchronizing LLDP topology...")
		if err := ps.syncLLDPTopology(ctx); err != nil {
			allErrors = append(allErrors, fmt.Errorf("LLDP sync failed: %w", err))
			ps.logger.Printf("LLDP sync failed: %v", err)
		} else {
			ps.logger.Println("Phase 2: LLDP synchronization completed successfully")
		}
	}

	if len(allErrors) > 0 {
		ps.logger.Printf("Complete topology synchronization finished with %d errors", len(allErrors))
		return fmt.Errorf("topology sync errors: %v", allErrors)
	}

	ps.logger.Println("Complete topology synchronization finished successfully")
	return nil
}

func (ps *PrometheusSync) syncLLDPTopology(ctx context.Context) error {
	ps.logger.Println("Starting LLDP topology synchronization...")

	// Extract links using MetricsExtractor with fallback support
	links, warnings := ps.metricsExtractor.ExtractLinks(ctx)
	
	// Log warnings (data missing scenarios)
	for _, warning := range warnings {
		ps.logger.Printf("Info: %v", warning)
	}

	if len(links) == 0 {
		ps.logger.Println("No links extracted from Prometheus - skipping this cycle")
		return nil
	}

	ps.logger.Printf("Successfully extracted %d links using metrics mapping", len(links))

	// Ensure all devices referenced by links exist before inserting links
	if err := ps.ensureReferencedDevicesExist(ctx, links); err != nil {
		return fmt.Errorf("failed to ensure referenced devices exist: %w", err)
	}

	// Batch process links
	if err := ps.batchAddLinks(ctx, links); err != nil {
		return fmt.Errorf("failed to add links: %w", err)
	}

	ps.logger.Printf("LLDP topology synchronization completed, processed %d links", len(links))
	return nil
}

func (ps *PrometheusSync) syncDeviceInfo(ctx context.Context) error {
	ps.logger.Println("Starting device information synchronization...")

	// Extract devices using MetricsExtractor with fallback support
	devices, warnings := ps.metricsExtractor.ExtractDevices(ctx)
	
	// Log warnings (data missing scenarios)
	for _, warning := range warnings {
		ps.logger.Printf("Info: %v", warning)
	}

	if len(devices) == 0 {
		ps.logger.Println("No devices extracted from Prometheus - skipping this cycle")
		return nil
	}

	ps.logger.Printf("Successfully extracted %d devices using metrics mapping", len(devices))

	// Batch process devices
	if err := ps.batchAddDevices(ctx, devices); err != nil {
		return fmt.Errorf("failed to add/update devices: %w", err)
	}

	ps.logger.Printf("Device information synchronization completed, processed %d devices", len(devices))
	return nil
}

func (ps *PrometheusSync) cleanupOldData(ctx context.Context) error {
	ps.logger.Println("Starting data cleanup...")

	// Note: This is a simplified cleanup implementation
	// In a real implementation, you would want to:
	// 1. Find devices/links not seen for MaxDeviceAge/MaxLinkAge
	// 2. Remove them from the database
	// 3. Handle cascade deletions properly

	ps.logger.Println("Data cleanup completed")
	return nil
}

func (ps *PrometheusSync) batchAddDevices(ctx context.Context, devices []topology.Device) error {
	batchSize := ps.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(devices); i += batchSize {
		end := i + batchSize
		if end > len(devices) {
			end = len(devices)
		}

		batch := devices[i:end]
		if err := ps.repository.BulkAddDevices(ctx, batch); err != nil {
			return fmt.Errorf("failed to add device batch %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

func (ps *PrometheusSync) batchAddLinks(ctx context.Context, links []topology.Link) error {
	batchSize := ps.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(links); i += batchSize {
		end := i + batchSize
		if end > len(links) {
			end = len(links)
		}

		batch := links[i:end]
		if err := ps.repository.BulkAddLinks(ctx, batch); err != nil {
			return fmt.Errorf("failed to add link batch %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

// ensureReferencedDevicesExist creates placeholder devices for any device IDs referenced in links but not yet in the database
func (ps *PrometheusSync) ensureReferencedDevicesExist(ctx context.Context, links []topology.Link) error {
	// Collect all unique device IDs referenced in links
	deviceIDSet := make(map[string]bool)
	for _, link := range links {
		if link.SourceID != "" {
			deviceIDSet[link.SourceID] = true
		}
		if link.TargetID != "" {
			deviceIDSet[link.TargetID] = true
		}
	}

	// Convert to slice
	var deviceIDs []string
	for deviceID := range deviceIDSet {
		deviceIDs = append(deviceIDs, deviceID)
	}

	if len(deviceIDs) == 0 {
		return nil
	}

	ps.logger.Printf("Checking existence of %d devices referenced in links", len(deviceIDs))

	// Check which devices already exist
	existingDevices := make(map[string]bool)
	for _, deviceID := range deviceIDs {
		if device, err := ps.repository.GetDevice(ctx, deviceID); err == nil && device != nil {
			existingDevices[deviceID] = true
		}
	}

	// Create placeholder devices for missing ones
	var missingDevices []topology.Device
	now := time.Now()
	
	for _, deviceID := range deviceIDs {
		if !existingDevices[deviceID] {
			device := topology.Device{
				ID:        deviceID,
				Type:      "unknown",
				Hardware:  "unknown", 
				Instance:  "unknown",
				Location:  "unknown",
				Status:    "unknown",
				Layer:     99,
				Metadata:  make(map[string]string),
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			missingDevices = append(missingDevices, device)
		}
	}

	if len(missingDevices) > 0 {
		ps.logger.Printf("Creating %d placeholder devices for LLDP-discovered devices not in Prometheus monitoring", len(missingDevices))
		for _, device := range missingDevices {
			ps.logger.Printf("  - Creating placeholder for device: %s (likely managed by another team)", device.ID)
		}
		if err := ps.batchAddDevices(ctx, missingDevices); err != nil {
			return fmt.Errorf("failed to create placeholder devices: %w", err)
		}
		ps.logger.Printf("Successfully created %d placeholder devices for LLDP-discovered neighbors", len(missingDevices))
	}

	return nil
}

// Health check for the worker
func (ps *PrometheusSync) Health(ctx context.Context) error {
	// Check Prometheus connectivity
	if err := ps.promClient.Health(ctx); err != nil {
		return fmt.Errorf("prometheus health check failed: %w", err)
	}

	// Check repository health
	if err := ps.repository.Health(ctx); err != nil {
		return fmt.Errorf("repository health check failed: %w", err)
	}

	return nil
}