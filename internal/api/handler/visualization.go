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
	// フロントエンドで使用中: /api/topology/{deviceId}
	huma.Register(api, huma.Operation{
		OperationID: "get-topology",
		Method:      http.MethodGet,
		Path:        "/api/topology/{deviceId}",
		Summary:     "Get visual topology",
		Tags:        []string{"visualization"},
	}, h.GetTopology)
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