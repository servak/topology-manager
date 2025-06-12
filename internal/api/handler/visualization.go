package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/servak/topology-manager/internal/domain/visualization"
	"github.com/servak/topology-manager/internal/service"
)

type VisualizationHandler struct {
	visualizationService *service.VisualizationService
}

func NewVisualizationHandler(visualizationService *service.VisualizationService) *VisualizationHandler {
	return &VisualizationHandler{
		visualizationService: visualizationService,
	}
}

func (h *VisualizationHandler) Register(api huma.API) {
	// 新しい形式: /api/topology/{deviceId}
	huma.Register(api, huma.Operation{
		OperationID: "get-topology",
		Method:      http.MethodGet,
		Path:        "/api/topology/{deviceId}",
		Summary:     "Get visual topology",
		Tags:        []string{"visualization"},
	}, h.GetTopology)

	// 旧来の形式: /api/topology?hostname=xxx (UI互換性のため)
	huma.Register(api, huma.Operation{
		OperationID: "get-topology-legacy",
		Method:      http.MethodGet,
		Path:        "/api/topology",
		Summary:     "Get visual topology (legacy format)",
		Tags:        []string{"visualization"},
	}, h.GetTopologyLegacy)
}

func (h *VisualizationHandler) GetTopology(ctx context.Context, input *struct {
	DeviceID string `path:"deviceId"`
	Depth    int    `query:"depth" default:"3"`
}) (*struct {
	Body visualization.VisualTopology
}, error) {
	visualTopology, err := h.visualizationService.GetVisualTopology(ctx, input.DeviceID, input.Depth)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get visual topology", err)
	}

	return &struct {
		Body visualization.VisualTopology
	}{
		Body: *visualTopology,
	}, nil
}

func (h *VisualizationHandler) GetTopologyLegacy(ctx context.Context, input *struct {
	Hostname string `query:"hostname"`
	Depth    int    `query:"depth" default:"3"`
}) (*struct {
	Body visualization.VisualTopology
}, error) {
	if input.Hostname == "" {
		return nil, huma.Error400BadRequest("hostname parameter is required")
	}

	visualTopology, err := h.visualizationService.GetVisualTopology(ctx, input.Hostname, input.Depth)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get visual topology", err)
	}

	return &struct {
		Body visualization.VisualTopology
	}{
		Body: *visualTopology,
	}, nil
}