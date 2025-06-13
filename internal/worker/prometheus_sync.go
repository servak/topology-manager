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
	promClient   *prometheus.Client
	lldpParser   *prometheus.LLDPParser
	repository   topology.Repository
	scheduler    *Scheduler
	logger       *log.Logger
	config       PrometheusSyncConfig
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
	repository topology.Repository,
	config PrometheusSyncConfig,
	logger *log.Logger,
) *PrometheusSync {
	if logger == nil {
		logger = log.Default()
	}

	lldpParser := prometheus.NewLLDPParser(promClient)
	scheduler := NewScheduler(logger)

	return &PrometheusSync{
		promClient: promClient,
		lldpParser: lldpParser,
		repository: repository,
		scheduler:  scheduler,
		logger:     logger,
		config:     config,
	}
}

// Start starts the Prometheus synchronization worker
func (ps *PrometheusSync) Start() error {
	ps.logger.Println("Starting Prometheus synchronization worker...")

	// Add LLDP synchronization task
	if ps.config.EnableLLDPSync {
		lldpTask := NewTaskBuilder("lldp_sync", "LLDP Topology Sync").
			Description("Synchronizes network topology from LLDP data in Prometheus").
			Interval(ps.config.LLDPSyncInterval).
			Timeout(ps.config.SyncTimeout).
			Function(ps.syncLLDPTopology).
			Build()

		if err := ps.scheduler.AddTask(lldpTask); err != nil {
			return fmt.Errorf("failed to add LLDP sync task: %w", err)
		}
	}

	// Add device synchronization task
	if ps.config.EnableDeviceSync {
		deviceTask := NewTaskBuilder("device_sync", "Device Info Sync").
			Description("Synchronizes device information from Prometheus").
			Interval(ps.config.DeviceSyncInterval).
			Timeout(ps.config.SyncTimeout).
			Function(ps.syncDeviceInfo).
			Build()

		if err := ps.scheduler.AddTask(deviceTask); err != nil {
			return fmt.Errorf("failed to add device sync task: %w", err)
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

	if ps.config.EnableLLDPSync {
		if err := ps.scheduler.RunTaskNow("lldp_sync"); err != nil {
			errors = append(errors, fmt.Errorf("LLDP sync: %w", err))
		}
	}

	if ps.config.EnableDeviceSync {
		if err := ps.scheduler.RunTaskNow("device_sync"); err != nil {
			errors = append(errors, fmt.Errorf("device sync: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %v", errors)
	}

	return nil
}

// Private synchronization methods

func (ps *PrometheusSync) syncLLDPTopology(ctx context.Context) error {
	ps.logger.Println("Starting LLDP topology synchronization...")

	// Build topology from LLDP data
	devices, links, err := ps.lldpParser.BuildTopologyFromLLDP(ctx)
	if err != nil {
		return fmt.Errorf("failed to build topology from LLDP: %w", err)
	}

	ps.logger.Printf("Found %d devices and %d links from LLDP data", len(devices), len(links))

	// Batch process devices
	if len(devices) > 0 {
		if err := ps.batchAddDevices(ctx, devices); err != nil {
			return fmt.Errorf("failed to add devices: %w", err)
		}
		ps.logger.Printf("Successfully added/updated %d devices", len(devices))
	}

	// Batch process links
	if len(links) > 0 {
		if err := ps.batchAddLinks(ctx, links); err != nil {
			return fmt.Errorf("failed to add links: %w", err)
		}
		ps.logger.Printf("Successfully added/updated %d links", len(links))
	}

	ps.logger.Println("LLDP topology synchronization completed successfully")
	return nil
}

func (ps *PrometheusSync) syncDeviceInfo(ctx context.Context) error {
	ps.logger.Println("Starting device information synchronization...")

	// Parse device information from Prometheus
	deviceInfos, err := ps.lldpParser.ParseDeviceInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse device info: %w", err)
	}

	ps.logger.Printf("Found %d devices with additional information", len(deviceInfos))

	// Update devices with additional information
	var updateCount int
	for _, deviceInfo := range deviceInfos {
		deviceID := deviceInfo.DeviceID
		if deviceInfo.Hostname != "" {
			deviceID = deviceInfo.Hostname
		}

		// Get existing device
		existingDevice, err := ps.repository.GetDevice(ctx, deviceID)
		if err != nil {
			ps.logger.Printf("Failed to get device %s: %v", deviceID, err)
			continue
		}

		if existingDevice == nil {
			// Device doesn't exist, skip
			continue
		}

		// Update device with new information
		updated := false
		if deviceInfo.IPAddress != "" && existingDevice.IPAddress != deviceInfo.IPAddress {
			existingDevice.IPAddress = deviceInfo.IPAddress
			updated = true
		}
		if deviceInfo.Location != "" && existingDevice.Location != deviceInfo.Location {
			existingDevice.Location = deviceInfo.Location
			updated = true
		}
		if deviceInfo.Contact != "" {
			if existingDevice.Metadata == nil {
				existingDevice.Metadata = make(map[string]string)
			}
			if existingDevice.Metadata["contact"] != deviceInfo.Contact {
				existingDevice.Metadata["contact"] = deviceInfo.Contact
				updated = true
			}
		}

		if updated {
			existingDevice.LastSeen = deviceInfo.LastSeen
			existingDevice.UpdatedAt = time.Now()

			if err := ps.repository.UpdateDevice(ctx, *existingDevice); err != nil {
				ps.logger.Printf("Failed to update device %s: %v", deviceID, err)
				continue
			}
			updateCount++
		}
	}

	ps.logger.Printf("Device information synchronization completed, updated %d devices", updateCount)
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