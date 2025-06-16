package service

import (
	"context"

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

// トポロジー検索メソッド（フロントエンドで使用中）
func (s *TopologyService) FindReachableDevices(ctx context.Context, deviceID string, opts topology.ReachabilityOptions) ([]topology.Device, error) {
	return s.repo.FindReachableDevices(ctx, deviceID, opts)
}

func (s *TopologyService) FindShortestPath(ctx context.Context, fromID, toID string, opts topology.PathOptions) (*topology.Path, error) {
	return s.repo.FindShortestPath(ctx, fromID, toID, opts)
}

// SearchDevices searches for devices with the given query
func (s *TopologyService) SearchDevices(ctx context.Context, query string, limit int) ([]topology.Device, error) {
	if query == "" {
		return []topology.Device{}, nil
	}
	return s.repo.SearchDevices(ctx, query, limit)
}
