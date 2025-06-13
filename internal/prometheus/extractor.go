package prometheus

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/servak/topology-manager/internal/domain/topology"
)

// MetricConfigGroup holds primary and fallback configurations for a metric type
type MetricConfigGroup struct {
	Primary   MetricMapping   `yaml:"primary"`
	Fallbacks []MetricMapping `yaml:"fallbacks"`
}

// MetricMapping defines how to extract data from a specific metric
type MetricMapping struct {
	MetricName string            `yaml:"metric_name"`
	Labels     map[string]string `yaml:"labels"`
}

// FieldRequirement defines required and optional fields for validation
type FieldRequirement struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// MetricsConfig holds metrics mapping configuration
type MetricsConfig struct {
	MetricsMapping    map[string]MetricConfigGroup   `yaml:"metrics_mapping"`
	FieldRequirements map[string]FieldRequirement   `yaml:"field_requirements"`
}

// MetricsExtractor extracts network topology data from Prometheus metrics
type MetricsExtractor struct {
	client *Client
	config *MetricsConfig
}

// NewMetricsExtractor creates a new MetricsExtractor instance
func NewMetricsExtractor(client *Client, config *MetricsConfig) *MetricsExtractor {
	return &MetricsExtractor{
		client: client,
		config: config,
	}
}

// ExtractDevices extracts device information from Prometheus metrics
func (e *MetricsExtractor) ExtractDevices(ctx context.Context) ([]topology.Device, []error) {
	var warnings []error

	deviceConfig, exists := e.config.MetricsMapping["device_info"]
	if !exists {
		return nil, []error{fmt.Errorf("device_info mapping not found in configuration")}
	}

	// Try primary metric first
	devices, err := e.tryExtractDevices(ctx, deviceConfig.Primary, "device_info")
	if err == nil && len(devices) > 0 {
		log.Printf("Successfully extracted %d devices using primary metric '%s'", len(devices), deviceConfig.Primary.MetricName)
		return e.validateAndCleanDevices(devices, "device_info"), warnings
	}
	warnings = append(warnings, fmt.Errorf("primary metric '%s' failed: %w", deviceConfig.Primary.MetricName, err))

	// Try fallback metrics
	for i, fallback := range deviceConfig.Fallbacks {
		devices, err := e.tryExtractDevices(ctx, fallback, "device_info")
		if err == nil && len(devices) > 0 {
			log.Printf("Successfully extracted %d devices using fallback %d metric '%s'", len(devices), i+1, fallback.MetricName)
			return e.validateAndCleanDevices(devices, "device_info"), warnings
		}
		warnings = append(warnings, fmt.Errorf("fallback %d metric '%s' failed: %w", i+1, fallback.MetricName, err))
	}

	return nil, warnings
}

// ExtractLinks extracts link information from Prometheus metrics
func (e *MetricsExtractor) ExtractLinks(ctx context.Context) ([]topology.Link, []error) {
	var warnings []error

	linkConfig, exists := e.config.MetricsMapping["lldp_neighbors"]
	if !exists {
		return nil, []error{fmt.Errorf("lldp_neighbors mapping not found in configuration")}
	}

	// Try primary metric first
	links, err := e.tryExtractLinks(ctx, linkConfig.Primary, "lldp_neighbors")
	if err == nil && len(links) > 0 {
		log.Printf("Successfully extracted %d links using primary metric '%s'", len(links), linkConfig.Primary.MetricName)
		return e.validateAndCleanLinks(links, "lldp_neighbors"), warnings
	}
	warnings = append(warnings, fmt.Errorf("primary metric '%s' failed: %w", linkConfig.Primary.MetricName, err))

	// Try fallback metrics
	for i, fallback := range linkConfig.Fallbacks {
		links, err := e.tryExtractLinks(ctx, fallback, "lldp_neighbors")
		if err == nil && len(links) > 0 {
			log.Printf("Successfully extracted %d links using fallback %d metric '%s'", len(links), i+1, fallback.MetricName)
			return e.validateAndCleanLinks(links, "lldp_neighbors"), warnings
		}
		warnings = append(warnings, fmt.Errorf("fallback %d metric '%s' failed: %w", i+1, fallback.MetricName, err))
	}

	return nil, warnings
}

// tryExtractDevices attempts to extract devices from a specific metric configuration
func (e *MetricsExtractor) tryExtractDevices(ctx context.Context, mapping MetricMapping, configKey string) ([]topology.Device, error) {
	query := fmt.Sprintf(`{__name__="%s"}`, mapping.MetricName)
	
	result, err := e.client.Query(ctx, query, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("failed to query metric '%s': %w", mapping.MetricName, err)
	}

	var devices []topology.Device
	now := time.Now()

	for _, sample := range result.Data.Result {
		device := topology.Device{
			LastSeen:  now,
			CreatedAt: now,
			UpdatedAt: now,
			Metadata:  make(map[string]string),
		}

		// Extract fields based on label mapping
		if deviceID, exists := e.extractLabelValue(sample.Metric, mapping.Labels, "device_id"); exists && deviceID != "" {
			device.ID = deviceID
		} else {
			continue // Skip if no device ID
		}

		if hardware, exists := e.extractLabelValue(sample.Metric, mapping.Labels, "hardware"); exists && hardware != "" {
			device.Hardware = hardware
		}

		if location, exists := e.extractLabelValue(sample.Metric, mapping.Labels, "location"); exists && location != "" {
			device.Location = location
		}

		devices = append(devices, device)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices found in metric '%s'", mapping.MetricName)
	}

	return devices, nil
}

// tryExtractLinks attempts to extract links from a specific metric configuration
func (e *MetricsExtractor) tryExtractLinks(ctx context.Context, mapping MetricMapping, configKey string) ([]topology.Link, error) {
	query := fmt.Sprintf(`{__name__="%s"}`, mapping.MetricName)
	
	result, err := e.client.Query(ctx, query, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("failed to query metric '%s': %w", mapping.MetricName, err)
	}

	var links []topology.Link
	now := time.Now()

	for i, sample := range result.Data.Result {
		link := topology.Link{
			ID:        fmt.Sprintf("lldp-link-%d", i),
			LastSeen:  now,
			CreatedAt: now,
			UpdatedAt: now,
			Weight:    1.0,
			Status:    "up",
			Metadata:  make(map[string]string),
		}

		// Extract fields based on label mapping
		if sourceDevice, exists := e.extractLabelValue(sample.Metric, mapping.Labels, "source_device"); exists && sourceDevice != "" {
			link.SourceID = sourceDevice
		} else {
			continue // Skip if no source device
		}

		if targetDevice, exists := e.extractLabelValue(sample.Metric, mapping.Labels, "target_device"); exists && targetDevice != "" {
			link.TargetID = targetDevice
		} else {
			continue // Skip if no target device
		}

		if sourcePort, exists := e.extractLabelValue(sample.Metric, mapping.Labels, "source_port"); exists && sourcePort != "" {
			// Truncate to fit database constraints (VARCHAR(255))
			if len(sourcePort) > 255 {
				sourcePort = sourcePort[:252] + "..."
			}
			link.SourcePort = sourcePort
		}

		if targetPort, exists := e.extractLabelValue(sample.Metric, mapping.Labels, "target_port"); exists && targetPort != "" {
			// Truncate to fit database constraints (VARCHAR(255))
			if len(targetPort) > 255 {
				targetPort = targetPort[:252] + "..."
			}
			link.TargetPort = targetPort
		}

		links = append(links, link)
	}

	if len(links) == 0 {
		return nil, fmt.Errorf("no links found in metric '%s'", mapping.MetricName)
	}

	return links, nil
}

// extractLabelValue extracts a label value based on mapping configuration
func (e *MetricsExtractor) extractLabelValue(labels map[string]string, mapping map[string]string, field string) (string, bool) {
	prometheusLabel, exists := mapping[field]
	if !exists || prometheusLabel == "" {
		return "", false
	}

	value, exists := labels[prometheusLabel]
	return value, exists && value != ""
}

// validateAndCleanDevices validates devices and fills missing optional fields
func (e *MetricsExtractor) validateAndCleanDevices(devices []topology.Device, configKey string) []topology.Device {
	requirements, exists := e.config.FieldRequirements[configKey]
	if !exists {
		return devices // No validation rules, return as-is
	}

	var validDevices []topology.Device

	for _, device := range devices {
		// Check required fields
		if !e.hasRequiredFields(device, requirements.Required) {
			log.Printf("Skipping device '%s': missing required fields", device.ID)
			continue
		}

		// Fill missing optional fields
		cleanDevice := e.fillMissingDeviceFields(device)
		validDevices = append(validDevices, cleanDevice)
	}

	log.Printf("Validated %d/%d devices", len(validDevices), len(devices))
	return validDevices
}

// validateAndCleanLinks validates links and fills missing optional fields
func (e *MetricsExtractor) validateAndCleanLinks(links []topology.Link, configKey string) []topology.Link {
	requirements, exists := e.config.FieldRequirements[configKey]
	if !exists {
		return links // No validation rules, return as-is
	}

	var validLinks []topology.Link

	for _, link := range links {
		// Check required fields
		if !e.hasRequiredLinkFields(link, requirements.Required) {
			log.Printf("Skipping link '%s': missing required fields", link.ID)
			continue
		}

		// Fill missing optional fields
		cleanLink := e.fillMissingLinkFields(link)
		validLinks = append(validLinks, cleanLink)
	}

	log.Printf("Validated %d/%d links", len(validLinks), len(links))
	return validLinks
}

// hasRequiredFields checks if device has all required fields
func (e *MetricsExtractor) hasRequiredFields(device topology.Device, required []string) bool {
	for _, field := range required {
		switch field {
		case "device_id":
			if device.ID == "" {
				return false
			}
		case "hardware":
			if device.Hardware == "" {
				return false
			}
		case "location":
			if device.Location == "" {
				return false
			}
		}
	}
	return true
}

// hasRequiredLinkFields checks if link has all required fields
func (e *MetricsExtractor) hasRequiredLinkFields(link topology.Link, required []string) bool {
	for _, field := range required {
		switch field {
		case "source_device":
			if link.SourceID == "" {
				return false
			}
		case "target_device":
			if link.TargetID == "" {
				return false
			}
		case "source_port":
			if link.SourcePort == "" {
				return false
			}
		case "target_port":
			if link.TargetPort == "" {
				return false
			}
		}
	}
	return true
}

// fillMissingDeviceFields fills missing optional fields with default values
func (e *MetricsExtractor) fillMissingDeviceFields(device topology.Device) topology.Device {
	// Skip setting IPAddress to "unknown" since PostgreSQL INET type doesn't accept it
	// Leave it empty if not available from metrics
	if device.Hardware == "" {
		device.Hardware = "unknown"
	}
	if device.Location == "" {
		device.Location = "unknown"
	}
	if device.Status == "" {
		device.Status = "up"
	}
	if device.Type == "" {
		device.Type = "unknown"
	}
	if device.Metadata == nil {
		device.Metadata = make(map[string]string)
	}

	return device
}

// fillMissingLinkFields fills missing optional fields with default values
func (e *MetricsExtractor) fillMissingLinkFields(link topology.Link) topology.Link {
	if link.SourcePort == "" {
		link.SourcePort = "unknown"
	}
	if link.TargetPort == "" {
		link.TargetPort = "unknown"
	}
	if link.Status == "" {
		link.Status = "up"
	}
	if link.Weight == 0 {
		link.Weight = 1.0
	}
	if link.Metadata == nil {
		link.Metadata = make(map[string]string)
	}

	return link
}