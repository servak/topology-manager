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

	// グループ展開用API
	huma.Register(api, huma.Operation{
		OperationID: "expand-group",
		Method:      http.MethodPost,
		Path:        "/api/topology/expand-group",
		Summary:     "Expand group node and get detailed topology",
		Tags:        []string{"visualization"},
	}, h.ExpandGroup)
}

func (h *VisualizationHandler) GetTopology(ctx context.Context, input *struct {
	DeviceID         string `path:"deviceId"`
	Depth            int    `query:"depth" default:"3"`
	EnableGrouping   bool   `query:"enable_grouping" default:"true"`
	MinGroupSize     int    `query:"min_group_size" default:"3"`
	MaxGroupDepth    int    `query:"max_group_depth" default:"2"`
	GroupByPrefix    bool   `query:"group_by_prefix" default:"true"`
	GroupByType      bool   `query:"group_by_type" default:"false"`
	PrefixMinLen     int    `query:"prefix_min_len" default:"3"`
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

func (h *VisualizationHandler) ExpandGroup(ctx context.Context, input *struct {
	Body struct {
		GroupID          string   `json:"group_id" doc:"Group ID to expand"`
		RootDeviceID     string   `json:"root_device_id" doc:"Original root device ID"`
		GroupDeviceIDs   []string `json:"group_device_ids" doc:"Device IDs within the group"`
		CurrentTopology  visualization.VisualTopology `json:"current_topology" doc:"Current topology state"`
		GroupingOptions  visualization.GroupingOptions `json:"grouping_options" doc:"Grouping configuration"`
		ExpandDepth      int      `json:"expand_depth" default:"2" doc:"Depth to expand from group devices"`
	} `json:"body"`
}) (*struct {
	Body struct {
		ExpandedTopology visualization.VisualTopology `json:"expanded_topology" doc:"Topology with expanded group"`
		NewNodes         []visualization.VisualNode   `json:"new_nodes" doc:"Newly added nodes"`
		NewEdges         []visualization.VisualEdge   `json:"new_edges" doc:"Newly added edges"`
	} `json:"body"`
}, error) {
	expandedTopology, newNodes, newEdges, err := h.visualizationService.ExpandGroupInTopology(
		ctx,
		input.Body.GroupID,
		input.Body.RootDeviceID,
		input.Body.GroupDeviceIDs,
		input.Body.CurrentTopology,
		input.Body.GroupingOptions,
		input.Body.ExpandDepth,
	)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to expand group", err)
	}

	return &struct {
		Body struct {
			ExpandedTopology visualization.VisualTopology `json:"expanded_topology" doc:"Topology with expanded group"`
			NewNodes         []visualization.VisualNode   `json:"new_nodes" doc:"Newly added nodes"`
			NewEdges         []visualization.VisualEdge   `json:"new_edges" doc:"Newly added edges"`
		} `json:"body"`
	}{
		Body: struct {
			ExpandedTopology visualization.VisualTopology `json:"expanded_topology" doc:"Topology with expanded group"`
			NewNodes         []visualization.VisualNode   `json:"new_nodes" doc:"Newly added nodes"`
			NewEdges         []visualization.VisualEdge   `json:"new_edges" doc:"Newly added edges"`
		}{
			ExpandedTopology: *expandedTopology,
			NewNodes:         newNodes,
			NewEdges:         newEdges,
		},
	}, nil
}