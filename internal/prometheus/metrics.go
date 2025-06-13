package prometheus

import (
	"fmt"
	"time"
)

// Common Prometheus metric queries for network topology discovery

const (
	// LLDP Metrics
	// These queries assume SNMP exporter or similar tools are collecting LLDP data
	
	// LLDP neighbor information - returns devices with their neighbors
	QueryLLDPNeighbors = `lldp_neighbor_info`
	
	// LLDP neighbor with detailed information including remote system details
	QueryLLDPNeighborDetails = `
		lldp_neighbor_info{
			remote_chassis_id!="",
			remote_system_name!=""
		}
	`
	
	// Active LLDP neighbors (recently seen)
	QueryActiveLLDPNeighbors = `
		lldp_neighbor_info[5m]
	`

	// Device Information Metrics
	
	// Basic device information from SNMP
	QueryDeviceInfo = `
		{__name__=~"device_info|snmp_.*", instance!=""}
	`
	
	// System information including hostname, description, location
	QuerySystemInfo = `
		{__name__=~"snmp_sysName|snmp_sysDescr|snmp_sysLocation|snmp_sysContact"}
	`
	
	// Device uptime information
	QueryDeviceUptime = `
		snmp_sysUpTime / 100
	`

	// Interface Metrics
	
	// Interface operational status (1=up, 2=down)
	QueryInterfaceStatus = `
		snmp_ifOperStatus
	`
	
	// Interface administrative status
	QueryInterfaceAdminStatus = `
		snmp_ifAdminStatus
	`
	
	// Interface names and descriptions
	QueryInterfaceInfo = `
		{__name__=~"snmp_ifName|snmp_ifDescr|snmp_ifAlias"}
	`
	
	// Interface speed in bits per second
	QueryInterfaceSpeed = `
		snmp_ifSpeed
	`
	
	// Interface traffic metrics
	QueryInterfaceTraffic = `
		rate(snmp_ifInOctets[5m]) * 8 or rate(snmp_ifOutOctets[5m]) * 8
	`

	// CDP Metrics (for Cisco devices)
	
	// CDP neighbor information
	QueryCDPNeighbors = `
		cdp_neighbor_info
	`

	// Network Discovery Queries
	
	// All network devices discovered
	QueryAllNetworkDevices = `
		{job=~"snmp.*", __name__=~"up|snmp_up"}
	`
	
	// Devices that are currently reachable
	QueryReachableDevices = `
		up == 1
	`
	
	// Recently discovered devices (last 10 minutes)
	QueryRecentDevices = `
		changes(up[10m]) > 0
	`
)

// MetricQuery represents a structured Prometheus query
type MetricQuery struct {
	Name        string
	Query       string
	Description string
	Labels      []string
	Interval    time.Duration
}

// PredefinedQueries contains commonly used queries for topology discovery
var PredefinedQueries = map[string]MetricQuery{
	"lldp_neighbors": {
		Name:        "lldp_neighbors",
		Query:       QueryLLDPNeighbors,
		Description: "LLDP neighbor information for topology discovery",
		Labels:      []string{"instance", "local_port", "remote_chassis_id", "remote_system_name"},
		Interval:    5 * time.Minute,
	},
	"device_info": {
		Name:        "device_info",
		Query:       QueryDeviceInfo,
		Description: "Basic device information from SNMP",
		Labels:      []string{"instance", "hostname", "system_desc", "location"},
		Interval:    10 * time.Minute,
	},
	"interface_status": {
		Name:        "interface_status",
		Query:       QueryInterfaceStatus,
		Description: "Interface operational status",
		Labels:      []string{"instance", "ifIndex", "ifName"},
		Interval:    1 * time.Minute,
	},
	"reachable_devices": {
		Name:        "reachable_devices",
		Query:       QueryReachableDevices,
		Description: "Currently reachable network devices",
		Labels:      []string{"instance", "job"},
		Interval:    30 * time.Second,
	},
}

// LLDPMetricLabels defines expected labels for LLDP metrics
type LLDPMetricLabels struct {
	// Local device information
	Instance  string `json:"instance"`   // Local device IP/hostname
	LocalPort string `json:"local_port"` // Local port identifier
	
	// Remote device information
	RemoteChassisID    string `json:"remote_chassis_id"`    // Remote device chassis ID
	RemoteSystemName   string `json:"remote_system_name"`   // Remote device hostname
	RemotePortID       string `json:"remote_port_id"`       // Remote port identifier
	RemotePortDesc     string `json:"remote_port_desc"`     // Remote port description
	RemoteSystemDesc   string `json:"remote_system_desc"`   // Remote system description
	RemoteMgmtAddr     string `json:"remote_mgmt_addr"`     // Remote management address
	
	// Additional LLDP information
	RemotePortType     string `json:"remote_port_type"`     // Remote port type
	RemoteChassisType  string `json:"remote_chassis_type"`  // Remote chassis type
}

// DeviceMetricLabels defines expected labels for device metrics
type DeviceMetricLabels struct {
	Instance    string `json:"instance"`     // Device IP/hostname
	Job         string `json:"job"`          // Prometheus job name
	Hostname    string `json:"hostname"`     // Device hostname from SNMP
	SystemDesc  string `json:"system_desc"`  // System description
	Location    string `json:"location"`     // Physical location
	Contact     string `json:"contact"`      // System contact
	ObjectID    string `json:"object_id"`    // SNMP system object ID
}

// InterfaceMetricLabels defines expected labels for interface metrics
type InterfaceMetricLabels struct {
	Instance  string `json:"instance"`   // Device IP/hostname
	IfIndex   string `json:"ifIndex"`    // Interface index
	IfName    string `json:"ifName"`     // Interface name
	IfDescr   string `json:"ifDescr"`    // Interface description
	IfAlias   string `json:"ifAlias"`    // Interface alias
	IfType    string `json:"ifType"`     // Interface type
}

// QueryTemplate represents a template for building dynamic queries
type QueryTemplate struct {
	Template    string
	Description string
	Parameters  []string
}

// Query templates for dynamic query building
var QueryTemplates = map[string]QueryTemplate{
	"device_neighbors": {
		Template:    `lldp_neighbor_info{instance="%s"}`,
		Description: "Get LLDP neighbors for a specific device",
		Parameters:  []string{"instance"},
	},
	"interface_status_by_device": {
		Template:    `snmp_ifOperStatus{instance="%s"}`,
		Description: "Get interface status for a specific device",
		Parameters:  []string{"instance"},
	},
	"devices_by_location": {
		Template:    `snmp_sysLocation{location=~".*%s.*"}`,
		Description: "Find devices in a specific location",
		Parameters:  []string{"location_pattern"},
	},
	"devices_by_type": {
		Template:    `device_info{system_desc=~".*%s.*"}`,
		Description: "Find devices by system description pattern",
		Parameters:  []string{"type_pattern"},
	},
}

// MetricCollectionConfig defines configuration for metric collection
type MetricCollectionConfig struct {
	// Collection intervals
	LLDPInterval      time.Duration `yaml:"lldp_interval"`
	DeviceInterval    time.Duration `yaml:"device_interval"`
	InterfaceInterval time.Duration `yaml:"interface_interval"`
	
	// Query timeouts
	QueryTimeout time.Duration `yaml:"query_timeout"`
	
	// Data retention
	MaxAge time.Duration `yaml:"max_age"`
	
	// Filter configuration
	IncludePatterns []string `yaml:"include_patterns"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
}

// DefaultMetricCollectionConfig returns default collection configuration
func DefaultMetricCollectionConfig() MetricCollectionConfig {
	return MetricCollectionConfig{
		LLDPInterval:      5 * time.Minute,
		DeviceInterval:    10 * time.Minute,
		InterfaceInterval: 2 * time.Minute,
		QueryTimeout:      30 * time.Second,
		MaxAge:            24 * time.Hour,
		IncludePatterns:   []string{".*"},
		ExcludePatterns:   []string{},
	}
}

// BuildQuery builds a query from a template with parameters
func BuildQuery(templateName string, params ...string) (string, error) {
	template, exists := QueryTemplates[templateName]
	if !exists {
		return "", fmt.Errorf("query template '%s' not found", templateName)
	}
	
	if len(params) != len(template.Parameters) {
		return "", fmt.Errorf("expected %d parameters for template '%s', got %d", 
			len(template.Parameters), templateName, len(params))
	}
	
	query := template.Template
	for _, param := range params {
		query = fmt.Sprintf(query, param)
	}
	
	return query, nil
}

// ValidateLabels checks if required labels are present in the metric
func ValidateLabels(labels map[string]string, required []string) error {
	for _, label := range required {
		if _, exists := labels[label]; !exists {
			return fmt.Errorf("required label '%s' not found", label)
		}
	}
	return nil
}