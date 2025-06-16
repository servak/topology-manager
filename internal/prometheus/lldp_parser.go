package prometheus

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/servak/topology-manager/internal/domain/topology"
)

// LLDPParser parses LLDP information from Prometheus metrics
type LLDPParser struct {
	client *Client
}

// NewLLDPParser creates a new LLDP parser
func NewLLDPParser(client *Client) *LLDPParser {
	return &LLDPParser{
		client: client,
	}
}

// LLDPNeighbor represents LLDP neighbor information
type LLDPNeighbor struct {
	LocalDevice      string
	LocalPort        string
	RemoteChassisID  string
	RemotePortID     string
	RemoteSystemName string
	RemoteSystemDesc string
	RemotePortDesc   string
	LastSeen         time.Time
}

// DeviceInfo represents device information from SNMP/other sources
type DeviceInfo struct {
	DeviceID   string
	Hostname   string
	SystemDesc string
	Location   string
	Contact    string
	Uptime     time.Duration
	LastSeen   time.Time
}

// ParseLLDPNeighbors parses LLDP neighbor information from Prometheus
func (p *LLDPParser) ParseLLDPNeighbors(ctx context.Context) ([]LLDPNeighbor, error) {
	result, err := p.client.GetLLDPNeighbors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLDP neighbors: %w", err)
	}

	var neighbors []LLDPNeighbor

	for _, r := range result.Data.Result {
		neighbor := LLDPNeighbor{
			LocalDevice:      r.Metric["instance"],
			LocalPort:        r.Metric["local_port"],
			RemoteChassisID:  r.Metric["remote_chassis_id"],
			RemotePortID:     r.Metric["remote_port_id"],
			RemoteSystemName: r.Metric["remote_system_name"],
			RemoteSystemDesc: r.Metric["remote_system_desc"],
			RemotePortDesc:   r.Metric["remote_port_desc"],
		}

		// Parse timestamp if available
		if len(r.Value) >= 2 {
			if timestamp, ok := r.Value[0].(float64); ok {
				neighbor.LastSeen = time.Unix(int64(timestamp), 0)
			}
		}

		// Clean up remote system name (remove domain suffixes, etc.)
		neighbor.RemoteSystemName = p.cleanSystemName(neighbor.RemoteSystemName)

		neighbors = append(neighbors, neighbor)
	}

	return neighbors, nil
}

// ParseDeviceInfo parses device information from Prometheus
func (p *LLDPParser) ParseDeviceInfo(ctx context.Context) ([]DeviceInfo, error) {
	result, err := p.client.GetDeviceInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	var devices []DeviceInfo

	for _, r := range result.Data.Result {
		device := DeviceInfo{
			DeviceID:   r.Metric["instance"],
			Hostname:   r.Metric["hostname"],
			SystemDesc: r.Metric["system_desc"],
			Location:   r.Metric["location"],
			Contact:    r.Metric["contact"],
		}

		// Parse timestamp
		if len(r.Value) >= 2 {
			if timestamp, ok := r.Value[0].(float64); ok {
				device.LastSeen = time.Unix(int64(timestamp), 0)
			}
		}

		// Parse uptime if available
		if uptimeStr, ok := r.Metric["uptime"]; ok {
			if uptime, err := strconv.ParseFloat(uptimeStr, 64); err == nil {
				device.Uptime = time.Duration(uptime) * time.Second
			}
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// BuildTopologyFromLLDP builds topology entities from LLDP data
func (p *LLDPParser) BuildTopologyFromLLDP(ctx context.Context) ([]topology.Device, []topology.Link, error) {
	// Get LLDP neighbors and device info
	neighbors, err := p.ParseLLDPNeighbors(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse LLDP neighbors: %w", err)
	}

	deviceInfos, err := p.ParseDeviceInfo(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse device info: %w", err)
	}

	// Create device map for quick lookup
	deviceMap := make(map[string]DeviceInfo)
	for _, device := range deviceInfos {
		// Index by hostname and device ID
		if device.Hostname != "" {
			deviceMap[device.Hostname] = device
		}
		if device.DeviceID != "" {
			deviceMap[device.DeviceID] = device
		}
	}

	// Track unique devices and links
	uniqueDevices := make(map[string]topology.Device)
	var links []topology.Link
	linkID := 0

	now := time.Now()

	for _, neighbor := range neighbors {
		// Create or update local device
		localDeviceID := p.resolveDeviceID(neighbor.LocalDevice, deviceMap)
		if localDevice, exists := uniqueDevices[localDeviceID]; !exists {
			device := p.createDeviceFromInfo(localDeviceID, neighbor.LocalDevice, deviceMap, now)
			uniqueDevices[localDeviceID] = device
		} else {
			// Update last seen time
			if neighbor.LastSeen.After(localDevice.LastSeen) {
				localDevice.LastSeen = neighbor.LastSeen
				localDevice.UpdatedAt = now
				uniqueDevices[localDeviceID] = localDevice
			}
		}

		// Create or update remote device
		remoteDeviceID := p.resolveDeviceID(neighbor.RemoteSystemName, deviceMap)
		if remoteDeviceID == "" {
			// Use chassis ID as fallback
			remoteDeviceID = p.normalizeChassisID(neighbor.RemoteChassisID)
		}

		if remoteDevice, exists := uniqueDevices[remoteDeviceID]; !exists {
			device := p.createDeviceFromInfo(remoteDeviceID, neighbor.RemoteSystemName, deviceMap, now)
			// Fill in additional info from LLDP if device info is not available
			if device.Hardware == "" && neighbor.RemoteSystemDesc != "" {
				device.Hardware = p.extractHardwareFromDesc(neighbor.RemoteSystemDesc)
			}
			uniqueDevices[remoteDeviceID] = device
		} else {
			// Update last seen time
			if neighbor.LastSeen.After(remoteDevice.LastSeen) {
				remoteDevice.LastSeen = neighbor.LastSeen
				remoteDevice.UpdatedAt = now
				uniqueDevices[remoteDeviceID] = remoteDevice
			}
		}

		// Create link
		linkID++
		link := topology.Link{
			ID:         fmt.Sprintf("lldp-link-%d", linkID),
			SourceID:   localDeviceID,
			TargetID:   remoteDeviceID,
			SourcePort: p.normalizePortName(neighbor.LocalPort),
			TargetPort: p.normalizePortName(neighbor.RemotePortID),
			Weight:     1.0,
			Metadata: map[string]string{
				"discovery_method":  "lldp",
				"remote_chassis_id": neighbor.RemoteChassisID,
				"remote_port_desc":  neighbor.RemotePortDesc,
			},
			LastSeen:  neighbor.LastSeen,
			CreatedAt: now,
			UpdatedAt: now,
		}

		links = append(links, link)
	}

	// Convert map to slice
	var devices []topology.Device
	for _, device := range uniqueDevices {
		devices = append(devices, device)
	}

	return devices, links, nil
}

// Helper methods

func (p *LLDPParser) cleanSystemName(name string) string {
	if name == "" {
		return name
	}

	// Remove common domain suffixes
	domainSuffixes := []string{".local", ".example.com", ".corp"}
	for _, suffix := range domainSuffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}

	// Remove FQDN if it contains dots
	if idx := strings.Index(name, "."); idx != -1 {
		name = name[:idx]
	}

	return strings.TrimSpace(name)
}

func (p *LLDPParser) resolveDeviceID(identifier string, deviceMap map[string]DeviceInfo) string {
	// Try direct lookup
	if device, exists := deviceMap[identifier]; exists {
		if device.Hostname != "" {
			return device.Hostname
		}
		return device.DeviceID
	}

	// Clean and return identifier
	return p.cleanSystemName(identifier)
}

func (p *LLDPParser) createDeviceFromInfo(deviceID, identifier string, deviceMap map[string]DeviceInfo, now time.Time) topology.Device {
	device := topology.Device{
		ID:        deviceID,
		Type:      "unknown",
		LayerID:   nil, // will be set by classification
		Metadata:  make(map[string]string),
		LastSeen:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Fill in additional info if available
	if deviceInfo, exists := deviceMap[identifier]; exists {
		if deviceInfo.SystemDesc != "" {
			device.Hardware = p.extractHardwareFromDesc(deviceInfo.SystemDesc)
		}
		if deviceInfo.Location != "" {
			device.Metadata["location"] = deviceInfo.Location
		}
		if deviceInfo.Contact != "" {
			device.Metadata["contact"] = deviceInfo.Contact
		}
		if !deviceInfo.LastSeen.IsZero() {
			device.LastSeen = deviceInfo.LastSeen
		}
	}

	// Determine device type from device ID patterns
	deviceType, _ := p.determineDeviceType(device.ID)
	device.Type = deviceType
	// LayerID will be set by classification service

	return device
}

func (p *LLDPParser) normalizeChassisID(chassisID string) string {
	// Remove common prefixes and clean up chassis ID
	chassisID = strings.TrimSpace(chassisID)

	// Remove MAC address formatting
	chassisID = strings.ReplaceAll(chassisID, ":", "")
	chassisID = strings.ReplaceAll(chassisID, "-", "")

	// Convert to lowercase
	return strings.ToLower(chassisID)
}

func (p *LLDPParser) normalizePortName(portName string) string {
	if portName == "" {
		return portName
	}

	// Common port name normalizations
	portName = strings.TrimSpace(portName)

	// Handle common variations
	patterns := map[string]string{
		`^GigabitEthernet(\d+/\d+)$`:    "Gi$1",
		`^TenGigabitEthernet(\d+/\d+)$`: "Te$1",
		`^FastEthernet(\d+/\d+)$`:       "Fa$1",
		`^Ethernet(\d+/\d+)$`:           "Eth$1",
	}

	for pattern, replacement := range patterns {
		if matched, _ := regexp.MatchString(pattern, portName); matched {
			re := regexp.MustCompile(pattern)
			portName = re.ReplaceAllString(portName, replacement)
			break
		}
	}

	// Truncate to fit database constraints (VARCHAR(255))
	if len(portName) > 255 {
		portName = portName[:252] + "..."
	}

	return portName
}

func (p *LLDPParser) extractHardwareFromDesc(systemDesc string) string {
	if systemDesc == "" {
		return ""
	}

	// Common hardware patterns
	patterns := []string{
		`Cisco\s+(\w+\s*\d+\w*)`,
		`Arista\s+DCS-(\d+\w*)`,
		`Juniper\s+(\w+\s*\d+\w*)`,
		`HP\s+(\w+\s*\d+\w*)`,
		`Dell\s+(\w+\s*\d+\w*)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(systemDesc); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// Extract first meaningful part of description
	parts := strings.Fields(systemDesc)
	if len(parts) > 0 {
		return parts[0]
	}

	return "Unknown"
}

func (p *LLDPParser) determineDeviceType(hostname string) (string, int) {
	if hostname == "" {
		return "unknown", 99
	}

	hostname = strings.ToLower(hostname)

	// Define patterns for device types
	patterns := map[string]struct {
		deviceType string
		layer      int
	}{
		`core`:         {"core", 1},
		`spine`:        {"core", 1},
		`dist`:         {"distribution", 2},
		`distribution`: {"distribution", 2},
		`agg`:          {"distribution", 2},
		`access`:       {"access", 3},
		`tor`:          {"access", 3},
		`leaf`:         {"access", 3},
		`server`:       {"server", 4},
		`srv`:          {"server", 4},
		`host`:         {"server", 4},
		`sw`:           {"switch", 3},
		`switch`:       {"switch", 3},
		`router`:       {"router", 2},
		`rt`:           {"router", 2},
		`fw`:           {"firewall", 2},
		`firewall`:     {"firewall", 2},
	}

	for pattern, info := range patterns {
		if strings.Contains(hostname, pattern) {
			return info.deviceType, info.layer
		}
	}

	return "unknown", 99
}
