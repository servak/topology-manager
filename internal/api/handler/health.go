package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/pkg/logger"
)

type HealthHandler struct {
	topologyRepo topology.Repository
	logger       *logger.Logger
}

type HealthResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Database string `json:"database"`
}

func NewHealthHandler(topologyRepo topology.Repository, appLogger *logger.Logger) *HealthHandler {
	return &HealthHandler{
		topologyRepo: topologyRepo,
		logger:       appLogger.WithComponent("health_handler"),
	}
}

func (h *HealthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/v1/health",
		Summary:     "Health check",
		Tags:        []string{"health"},
	}, h.HealthCheck)
}

func (h *HealthHandler) HealthCheck(ctx context.Context, input *struct{}) (*struct {
	Body HealthResponse
}, error) {
	response := HealthResponse{
		Status:   "healthy",
		Database: "healthy",
	}

	if err := h.topologyRepo.Health(ctx); err != nil {
		response.Status = "unhealthy"
		response.Database = "unhealthy"
		response.Message = "Database connection failed"

		return &struct {
			Body HealthResponse
		}{
			Body: response,
		}, huma.Error503ServiceUnavailable("Service unhealthy", err)
	}

	return &struct {
		Body HealthResponse
	}{
		Body: response,
	}, nil
}
