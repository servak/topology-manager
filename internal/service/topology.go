package service

import (
	"context"
	"fmt"

	"github.com/servak/topology-manager/internal/domain/topology"
)

type TopologyService struct {
	repo topology.Repository
}

func NewTopologyService(repo topology.Repository) *TopologyService {
	return &TopologyService{
		repo: repo,
	}
}

func (s *TopologyService) AddDevice(ctx context.Context, device topology.Device) error {
	if device.ID == "" {
		return fmt.Errorf("device ID is required")
	}
	if device.Name == "" {
		return fmt.Errorf("device name is required")
	}
	
	return s.repo.AddDevice(ctx, device)
}

func (s *TopologyService) AddLink(ctx context.Context, link topology.Link) error {
	if link.ID == "" {
		return fmt.Errorf("link ID is required")
	}
	if link.SourceID == "" || link.TargetID == "" {
		return fmt.Errorf("source and target device IDs are required")
	}
	
	// デバイスの存在確認
	sourceDevice, err := s.repo.GetDevice(ctx, link.SourceID)
	if err != nil {
		return fmt.Errorf("failed to check source device: %w", err)
	}
	if sourceDevice == nil {
		return fmt.Errorf("source device %s not found", link.SourceID)
	}
	
	targetDevice, err := s.repo.GetDevice(ctx, link.TargetID)
	if err != nil {
		return fmt.Errorf("failed to check target device: %w", err)
	}
	if targetDevice == nil {
		return fmt.Errorf("target device %s not found", link.TargetID)
	}
	
	return s.repo.AddLink(ctx, link)
}

func (s *TopologyService) GetDevice(ctx context.Context, deviceID string) (*topology.Device, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}
	
	return s.repo.GetDevice(ctx, deviceID)
}

func (s *TopologyService) GetDeviceWithNeighbors(ctx context.Context, deviceID string) (*DeviceWithNeighbors, error) {
	device, err := s.repo.GetDevice(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	
	links, err := s.repo.GetDeviceLinks(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device links: %w", err)
	}
	
	neighbors := make([]NeighborInfo, 0, len(links))
	visited := make(map[string]bool)
	
	for _, link := range links {
		var neighborID string
		var localPort, remotePort string
		
		if link.SourceID == deviceID {
			neighborID = link.TargetID
			localPort = link.SourcePort
			remotePort = link.TargetPort
		} else {
			neighborID = link.SourceID
			localPort = link.TargetPort
			remotePort = link.SourcePort
		}
		
		if visited[neighborID] {
			continue
		}
		visited[neighborID] = true
		
		neighbor, err := s.repo.GetDevice(ctx, neighborID)
		if err != nil {
			continue
		}
		if neighbor == nil {
			continue
		}
		
		neighbors = append(neighbors, NeighborInfo{
			Device:     *neighbor,
			LocalPort:  localPort,
			RemotePort: remotePort,
			Status:     link.Status,
		})
	}
	
	return &DeviceWithNeighbors{
		Device:    *device,
		Neighbors: neighbors,
	}, nil
}

func (s *TopologyService) FindDevicesByType(ctx context.Context, deviceType string) ([]topology.Device, error) {
	if deviceType == "" {
		return nil, fmt.Errorf("device type is required")
	}
	
	return s.repo.FindDevicesByType(ctx, deviceType)
}

func (s *TopologyService) FindDevicesByHardware(ctx context.Context, hardware string) ([]topology.Device, error) {
	if hardware == "" {
		return nil, fmt.Errorf("hardware is required")
	}
	
	return s.repo.FindDevicesByHardware(ctx, hardware)
}

func (s *TopologyService) FindDevicesByInstance(ctx context.Context, instance string) ([]topology.Device, error) {
	if instance == "" {
		return nil, fmt.Errorf("instance is required")
	}
	
	return s.repo.FindDevicesByInstance(ctx, instance)
}

func (s *TopologyService) RemoveDevice(ctx context.Context, deviceID string) error {
	if deviceID == "" {
		return fmt.Errorf("device ID is required")
	}
	
	// デバイスの存在確認
	device, err := s.repo.GetDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to check device: %w", err)
	}
	if device == nil {
		return fmt.Errorf("device %s not found", deviceID)
	}
	
	return s.repo.RemoveDevice(ctx, deviceID)
}

func (s *TopologyService) RemoveLink(ctx context.Context, linkID string) error {
	if linkID == "" {
		return fmt.Errorf("link ID is required")
	}
	
	// リンクの存在確認
	link, err := s.repo.GetLink(ctx, linkID)
	if err != nil {
		return fmt.Errorf("failed to check link: %w", err)
	}
	if link == nil {
		return fmt.Errorf("link %s not found", linkID)
	}
	
	return s.repo.RemoveLink(ctx, linkID)
}

func (s *TopologyService) BulkAddDevices(ctx context.Context, devices []topology.Device) error {
	if len(devices) == 0 {
		return nil
	}
	
	// バリデーション
	for i, device := range devices {
		if device.ID == "" {
			return fmt.Errorf("device at index %d: ID is required", i)
		}
		if device.Name == "" {
			return fmt.Errorf("device at index %d: name is required", i)
		}
	}
	
	return s.repo.BulkAddDevices(ctx, devices)
}

func (s *TopologyService) BulkAddLinks(ctx context.Context, links []topology.Link) error {
	if len(links) == 0 {
		return nil
	}
	
	// バリデーション
	for i, link := range links {
		if link.ID == "" {
			return fmt.Errorf("link at index %d: ID is required", i)
		}
		if link.SourceID == "" || link.TargetID == "" {
			return fmt.Errorf("link at index %d: source and target device IDs are required", i)
		}
	}
	
	return s.repo.BulkAddLinks(ctx, links)
}

func (s *TopologyService) UpdateDevice(ctx context.Context, device topology.Device) error {
	if device.ID == "" {
		return fmt.Errorf("device ID is required")
	}
	
	return s.repo.UpdateDevice(ctx, device)
}

func (s *TopologyService) UpdateLink(ctx context.Context, link topology.Link) error {
	if link.ID == "" {
		return fmt.Errorf("link ID is required")
	}
	
	return s.repo.UpdateLink(ctx, link)
}

func (s *TopologyService) GetLink(ctx context.Context, linkID string) (*topology.Link, error) {
	if linkID == "" {
		return nil, fmt.Errorf("link ID is required")
	}
	
	return s.repo.GetLink(ctx, linkID)
}

func (s *TopologyService) GetDevices(ctx context.Context, opts topology.PaginationOptions) ([]topology.Device, *topology.PaginationResult, error) {
	return s.repo.GetDevices(ctx, opts)
}

type DeviceWithNeighbors struct {
	Device    topology.Device `json:"device"`
	Neighbors []NeighborInfo  `json:"neighbors"`
}

type NeighborInfo struct {
	Device     topology.Device `json:"device"`
	LocalPort  string          `json:"local_port"`
	RemotePort string          `json:"remote_port"`
	Status     string          `json:"status"`
}