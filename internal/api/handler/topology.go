package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/service"
)

type TopologyHandler struct {
	topologyService *service.TopologyService
}

func NewTopologyHandler(topologyService *service.TopologyService) *TopologyHandler {
	return &TopologyHandler{
		topologyService: topologyService,
	}
}

func (h *TopologyHandler) Register(api huma.API) {
	// デバイス検索API（フロントエンドで使用中）
	huma.Register(api, huma.Operation{
		OperationID: "search-devices",
		Method:      http.MethodGet,
		Path:        "/api/v1/devices/search",
		Summary:     "Search devices by ID, name, or IP address",
		Tags:        []string{"devices"},
	}, h.SearchDevices)

	// トポロジー検索API（フロントエンドで使用中）
	huma.Register(api, huma.Operation{
		OperationID: "find-reachable-devices",
		Method:      http.MethodGet,
		Path:        "/api/v1/devices/{deviceId}/reachable",
		Summary:     "Find reachable devices using BFS/DFS",
		Tags:        []string{"topology-search"},
	}, h.FindReachableDevices)

	huma.Register(api, huma.Operation{
		OperationID: "find-shortest-path",
		Method:      http.MethodGet,
		Path:        "/api/v1/path/{fromId}/{toId}",
		Summary:     "Find shortest path between two devices",
		Tags:        []string{"topology-search"},
	}, h.FindShortestPath)
}

// トポロジー検索ハンドラー
func (h *TopologyHandler) FindReachableDevices(ctx context.Context, input *struct {
	DeviceID   string `path:"deviceId"`
	Algorithm  string `query:"algorithm" enum:"bfs,dfs" default:"bfs"`
	MaxHops    int    `query:"max_hops" default:"5"`
}) (*struct {
	Body struct {
		Devices   []topology.Device `json:"devices"`
		Algorithm string            `json:"algorithm"`
		MaxHops   int               `json:"max_hops"`
		Count     int               `json:"count"`
	}
}, error) {
	var algorithm topology.SearchAlgorithm
	switch input.Algorithm {
	case "dfs":
		algorithm = topology.AlgorithmDFS
	default:
		algorithm = topology.AlgorithmBFS
	}

	devices, err := h.topologyService.FindReachableDevices(ctx, input.DeviceID, topology.ReachabilityOptions{
		MaxHops:   input.MaxHops,
		Algorithm: algorithm,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to find reachable devices", err)
	}

	return &struct {
		Body struct {
			Devices   []topology.Device `json:"devices"`
			Algorithm string            `json:"algorithm"`
			MaxHops   int               `json:"max_hops"`
			Count     int               `json:"count"`
		}
	}{
		Body: struct {
			Devices   []topology.Device `json:"devices"`
			Algorithm string            `json:"algorithm"`
			MaxHops   int               `json:"max_hops"`
			Count     int               `json:"count"`
		}{
			Devices:   devices,
			Algorithm: input.Algorithm,
			MaxHops:   input.MaxHops,
			Count:     len(devices),
		},
	}, nil
}

func (h *TopologyHandler) FindShortestPath(ctx context.Context, input *struct {
	FromID    string `path:"fromId"`
	ToID      string `path:"toId"`
	Algorithm string `query:"algorithm" enum:"dijkstra,k_shortest" default:"dijkstra"`
}) (*struct {
	Body topology.Path
}, error) {
	var algorithm topology.PathAlgorithm
	switch input.Algorithm {
	case "k_shortest":
		algorithm = topology.PathAlgorithmKShortest
	default:
		algorithm = topology.PathAlgorithmDijkstra
	}

	path, err := h.topologyService.FindShortestPath(ctx, input.FromID, input.ToID, topology.PathOptions{
		Algorithm: algorithm,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to find shortest path", err)
	}

	return &struct {
		Body topology.Path
	}{
		Body: *path,
	}, nil
}

// SearchDevices searches for devices by ID, name, or IP address
func (h *TopologyHandler) SearchDevices(ctx context.Context, input *struct {
	Query string `query:"q"`
	Limit int    `query:"limit" default:"20"`
}) (*struct {
	Body struct {
		Devices []topology.Device `json:"devices"`
		Count   int               `json:"count"`
	}
}, error) {
	devices, err := h.topologyService.SearchDevices(ctx, input.Query, input.Limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to search devices", err)
	}

	return &struct {
		Body struct {
			Devices []topology.Device `json:"devices"`
			Count   int               `json:"count"`
		}
	}{
		Body: struct {
			Devices []topology.Device `json:"devices"`
			Count   int               `json:"count"`
		}{
			Devices: devices,
			Count:   len(devices),
		},
	}, nil
}