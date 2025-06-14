package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/servak/topology-manager/internal/domain/topology"
)

// TopologyRepository provides in-memory implementation for testing
type TopologyRepository struct {
	mu      sync.RWMutex
	devices map[string]topology.Device
	links   map[string]topology.Link
}

func NewTopologyRepository() *TopologyRepository {
	repo := &TopologyRepository{
		devices: make(map[string]topology.Device),
		links:   make(map[string]topology.Link),
	}
	
	// テスト用デバイスを追加
	testDevices := []topology.Device{
		{ID: "fw1.edge", Type: "firewall", Hardware: "Fortinet FortiGate", Instance: "edge.dmz"},
		{ID: "sw1.core", Type: "switch", Hardware: "Cisco Catalyst 9500", Instance: "core.dc1"},
		{ID: "srv1.web", Type: "server", Hardware: "Dell PowerEdge", Instance: "web.app"},
		{ID: "router1.main", Type: "router", Hardware: "Cisco ASR1000", Instance: "main.core"},
		{ID: "switch1.access", Type: "switch", Hardware: "Arista 7280", Instance: "access.floor1"},
		{ID: "server1.db", Type: "server", Hardware: "HPE ProLiant", Instance: "db.backend"},
		{ID: "firewall1.dmz", Type: "firewall", Hardware: "Palo Alto PA-3200", Instance: "dmz.edge"},
		{ID: "ap1.wifi", Type: "access_point", Hardware: "Ubiquiti UniFi", Instance: "wifi.office"},
		{ID: "lb1.frontend", Type: "load_balancer", Hardware: "F5 BIG-IP", Instance: "frontend.web"},
		{ID: "router2.backup", Type: "router", Hardware: "Juniper MX240", Instance: "backup.wan"},
		{ID: "switch2.mgmt", Type: "switch", Hardware: "Cisco 3750", Instance: "mgmt.oob"},
		{ID: "server2.app", Type: "server", Hardware: "IBM System x", Instance: "app.prod"},
	}
	
	for _, device := range testDevices {
		repo.devices[device.ID] = device
	}
	
	return repo
}

// Device operations
func (r *TopologyRepository) AddDevice(ctx context.Context, device topology.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[device.ID] = device
	return nil
}

func (r *TopologyRepository) GetDevice(ctx context.Context, deviceID string) (*topology.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if device, exists := r.devices[deviceID]; exists {
		return &device, nil
	}
	return nil, nil
}

func (r *TopologyRepository) SearchDevices(ctx context.Context, query string, limit int) ([]topology.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []topology.Device
	count := 0
	for _, device := range r.devices {
		if limit > 0 && count >= limit {
			break
		}
		// Simple search implementation - check if query is in device ID or type
		if query == "" || 
			device.ID == query || 
			device.Type == query ||
			device.Hardware == query {
			result = append(result, device)
			count++
		}
	}
	return result, nil
}

func (r *TopologyRepository) UpdateDevice(ctx context.Context, device topology.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.devices[device.ID]; !exists {
		return fmt.Errorf("device not found: %s", device.ID)
	}
	r.devices[device.ID] = device
	return nil
}

func (r *TopologyRepository) RemoveDevice(ctx context.Context, deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.devices, deviceID)
	return nil
}

// Link operations
func (r *TopologyRepository) AddLink(ctx context.Context, link topology.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ID] = link
	return nil
}

func (r *TopologyRepository) GetLink(ctx context.Context, linkID string) (*topology.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if link, exists := r.links[linkID]; exists {
		return &link, nil
	}
	return nil, nil
}

func (r *TopologyRepository) UpdateLink(ctx context.Context, link topology.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.links[link.ID]; !exists {
		return fmt.Errorf("link not found: %s", link.ID)
	}
	r.links[link.ID] = link
	return nil
}

func (r *TopologyRepository) RemoveLink(ctx context.Context, linkID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.links, linkID)
	return nil
}

// Search operations
func (r *TopologyRepository) FindReachableDevices(ctx context.Context, deviceID string, opts topology.ReachabilityOptions) ([]topology.Device, error) {
	// Simple implementation for testing
	return []topology.Device{}, nil
}

func (r *TopologyRepository) ExtractSubTopology(ctx context.Context, deviceID string, opts topology.SubTopologyOptions) ([]topology.Device, []topology.Link, error) {
	// Simple implementation for testing
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var devices []topology.Device
	var links []topology.Link
	
	// Just return all devices and links for simplicity
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	for _, link := range r.links {
		links = append(links, link)
	}
	
	return devices, links, nil
}

func (r *TopologyRepository) FindShortestPath(ctx context.Context, fromID, toID string, opts topology.PathOptions) (*topology.Path, error) {
	// Simple implementation for testing
	return &topology.Path{}, nil
}

// Filter operations
func (r *TopologyRepository) FindDevicesByType(ctx context.Context, deviceType string) ([]topology.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []topology.Device
	for _, device := range r.devices {
		if device.Type == deviceType {
			result = append(result, device)
		}
	}
	return result, nil
}

func (r *TopologyRepository) FindDevicesByHardware(ctx context.Context, hardware string) ([]topology.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []topology.Device
	for _, device := range r.devices {
		if device.Hardware == hardware {
			result = append(result, device)
		}
	}
	return result, nil
}

func (r *TopologyRepository) FindDevicesByInstance(ctx context.Context, instance string) ([]topology.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []topology.Device
	for _, device := range r.devices {
		if device.Instance == instance {
			result = append(result, device)
		}
	}
	return result, nil
}

// Link search
func (r *TopologyRepository) GetDeviceLinks(ctx context.Context, deviceID string) ([]topology.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []topology.Link
	for _, link := range r.links {
		if link.SourceID == deviceID || link.TargetID == deviceID {
			result = append(result, link)
		}
	}
	return result, nil
}

func (r *TopologyRepository) FindLinksByPort(ctx context.Context, deviceID, port string) ([]topology.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []topology.Link
	for _, link := range r.links {
		if (link.SourceID == deviceID && link.SourcePort == port) ||
			(link.TargetID == deviceID && link.TargetPort == port) {
			result = append(result, link)
		}
	}
	return result, nil
}

// Bulk operations
func (r *TopologyRepository) BulkAddDevices(ctx context.Context, devices []topology.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, device := range devices {
		r.devices[device.ID] = device
	}
	return nil
}

func (r *TopologyRepository) BulkAddLinks(ctx context.Context, links []topology.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, link := range links {
		r.links[link.ID] = link
	}
	return nil
}

func (r *TopologyRepository) Close() error {
	return nil
}

func (r *TopologyRepository) Health(ctx context.Context) error {
	return nil
}