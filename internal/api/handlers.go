package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/servak/topology-manager/internal/storage"
	"github.com/servak/topology-manager/internal/topology"
)

type TopologyHandler struct {
	builder *topology.TopologyBuilder
	redis   *storage.RedisClient
}

func NewTopologyHandler(builder *topology.TopologyBuilder, redis *storage.RedisClient) *TopologyHandler {
	return &TopologyHandler{
		builder: builder,
		redis:   redis,
	}
}

type TopologyRequest struct {
	Hostname string `query:"hostname" required:"true" doc:"Root device hostname"`
	Depth    int    `query:"depth" default:"3" doc:"Exploration depth"`
}

type TopologyResponse struct {
	Body topology.Topology
}

type DeviceResponse struct {
	Body topology.DeviceInfo
}

type HealthResponse struct {
	Body struct {
		Status string `json:"status" doc:"Service status"`
		Redis  string `json:"redis" doc:"Redis connection status"`
	}
}

type ErrorResponse struct {
	Body struct {
		Message string `json:"message" doc:"Error message"`
	}
}

func (h *TopologyHandler) GetTopology(ctx context.Context, req *TopologyRequest) (*TopologyResponse, error) {
	if req.Hostname == "" {
		return nil, huma.Error400BadRequest("hostname parameter is required")
	}

	if req.Depth <= 0 {
		req.Depth = 3
	}

	topo, err := h.builder.BuildTopology(ctx, req.Hostname, req.Depth)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("Failed to build topology: %v", err))
	}

	return &TopologyResponse{Body: *topo}, nil
}

func (h *TopologyHandler) GetDevice(ctx context.Context, deviceName string) (*DeviceResponse, error) {
	if deviceName == "" {
		return nil, huma.Error400BadRequest("device name is required")
	}

	deviceInfo, err := h.builder.GetDeviceInfo(ctx, deviceName)
	if err != nil {
		return nil, huma.Error404NotFound(fmt.Sprintf("Device not found: %v", err))
	}

	return &DeviceResponse{Body: *deviceInfo}, nil
}

func (h *TopologyHandler) GetHealth(ctx context.Context, input *struct{}) (*HealthResponse, error) {
	response := &HealthResponse{}
	response.Body.Status = "ok"

	if err := h.redis.Health(ctx); err != nil {
		response.Body.Redis = fmt.Sprintf("unhealthy: %v", err)
		response.Body.Status = "degraded"
	} else {
		response.Body.Redis = "healthy"
	}

	return response, nil
}

func RegisterRoutes(api huma.API, handler *TopologyHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "get-topology",
		Method:      http.MethodGet,
		Path:        "/api/topology",
		Summary:     "Get network topology",
		Description: "Returns hierarchical network topology from a specified root device",
		Tags:        []string{"topology"},
	}, handler.GetTopology)

	huma.Register(api, huma.Operation{
		OperationID: "get-device",
		Method:      http.MethodGet,
		Path:        "/api/device/{name}",
		Summary:     "Get device information",
		Description: "Returns detailed information about a specific device including neighbors",
		Tags:        []string{"devices"},
	}, func(ctx context.Context, input *struct {
		Name string `path:"name" required:"true" doc:"Device name"`
	}) (*DeviceResponse, error) {
		return handler.GetDevice(ctx, input.Name)
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-health",
		Method:      http.MethodGet,
		Path:        "/api/health",
		Summary:     "Health check",
		Description: "Returns service health status including database connectivity",
		Tags:        []string{"system"},
	}, handler.GetHealth)
}