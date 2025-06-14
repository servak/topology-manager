package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/servak/topology-manager/internal/domain/visualization"
	"github.com/servak/topology-manager/internal/service"
	"github.com/servak/topology-manager/pkg/logger"
)

type VisualizationHandler struct {
	visualizationService *service.VisualizationService
	logger               *logger.Logger
}

func NewVisualizationHandler(visualizationService *service.VisualizationService, appLogger *logger.Logger) *VisualizationHandler {
	return &VisualizationHandler{
		visualizationService: visualizationService,
		logger:               appLogger.WithComponent("visualization_handler"),
	}
}

func (h *VisualizationHandler) Register(api huma.API) {
	// フロントエンドで使用中: /api/topology/{deviceId}
	huma.Register(api, huma.Operation{
		OperationID: "get-topology",
		Method:      http.MethodGet,
		Path:        "/api/v1/topology/{deviceId}",
		Summary:     "Get visual topology",
		Tags:        []string{"visualization"},
	}, h.GetTopology)

	// 階層表示用API - フロントエンドが期待するパス
	huma.Register(api, huma.Operation{
		OperationID: "get-visual-topology",
		Method:      http.MethodGet,
		Path:        "/api/v1/topology/visual/{deviceId}",
		Summary:     "Get hierarchical visual topology",
		Tags:        []string{"visualization"},
	}, h.GetVisualTopology)

	// グループ展開用API（新しいシンプル設計）
	huma.Register(api, huma.Operation{
		OperationID: "expand-from-device",
		Method:      http.MethodGet,
		Path:        "/api/v1/topology/{deviceId}/expand",
		Summary:     "Get topology expanding from specific device",
		Tags:        []string{"visualization"},
	}, h.ExpandFromDevice)
}

func (h *VisualizationHandler) GetTopology(ctx context.Context, input *struct {
	DeviceID       string `path:"deviceId"`
	Depth          int    `query:"depth" default:"3"`
	EnableGrouping bool   `query:"enable_grouping" default:"true"`
	MinGroupSize   int    `query:"min_group_size" default:"3"`
	MaxGroupDepth  int    `query:"max_group_depth" default:"2"`
	GroupByPrefix  bool   `query:"group_by_prefix" default:"true"`
	GroupByType    bool   `query:"group_by_type" default:"false"`
	PrefixMinLen   int    `query:"prefix_min_len" default:"3"`
}) (*struct {
	Body visualization.VisualTopology
}, error) {
	groupingOpts := visualization.GroupingOptions{
		Enabled:       input.EnableGrouping,
		MinGroupSize:  input.MinGroupSize,
		MaxDepth:      input.MaxGroupDepth,
		GroupByPrefix: input.GroupByPrefix,
		GroupByType:   input.GroupByType,
		PrefixMinLen:  input.PrefixMinLen,
	}

	visualTopology, err := h.visualizationService.GetVisualTopologyWithGrouping(ctx, input.DeviceID, input.Depth, groupingOpts)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get visual topology", err)
	}

	return &struct {
		Body visualization.VisualTopology
	}{
		Body: *visualTopology,
	}, nil
}

// GetVisualTopology returns topology data optimized for hierarchical display
func (h *VisualizationHandler) GetVisualTopology(ctx context.Context, input *struct {
	DeviceID string `path:"deviceId"`
	Depth    int    `query:"depth" default:"3"`
}) (*struct {
	Body visualization.VisualTopology
}, error) {
	// シンプルなビジュアルトポロジー取得（グループ化なし）
	visualTopology, err := h.visualizationService.GetSimpleVisualTopology(ctx, input.DeviceID, input.Depth)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get visual topology", err)
	}

	return &struct {
		Body visualization.VisualTopology
	}{
		Body: *visualTopology,
	}, nil
}

func (h *VisualizationHandler) ExpandFromDevice(ctx context.Context, input *struct {
	DeviceID       string `path:"deviceId"`
	Depth          int    `query:"depth" default:"2"`
	EnableGrouping bool   `query:"enable_grouping" default:"true"`
	MinGroupSize   int    `query:"min_group_size" default:"3"`
	MaxGroupDepth  int    `query:"max_group_depth" default:"2"`
	GroupByPrefix  bool   `query:"group_by_prefix" default:"true"`
	GroupByType    bool   `query:"group_by_type" default:"false"`
	GroupByDepth   bool   `query:"group_by_depth" default:"false"`
	PrefixMinLen   int    `query:"prefix_min_len" default:"3"`
}) (*struct {
	Body visualization.VisualTopology
}, error) {
	groupingOpts := visualization.GroupingOptions{
		Enabled:       input.EnableGrouping,
		MinGroupSize:  input.MinGroupSize,
		MaxDepth:      input.MaxGroupDepth,
		GroupByPrefix: input.GroupByPrefix,
		GroupByType:   input.GroupByType,
		GroupByDepth:  input.GroupByDepth,
		PrefixMinLen:  input.PrefixMinLen,
	}

	visualTopology, err := h.visualizationService.GetVisualTopologyWithGrouping(ctx, input.DeviceID, input.Depth, groupingOpts)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get visual topology", err)
	}

	return &struct {
		Body visualization.VisualTopology
	}{
		Body: *visualTopology,
	}, nil
}
