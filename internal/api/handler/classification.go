package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/service"
	"github.com/servak/topology-manager/pkg/logger"
)

type ClassificationHandler struct {
	classificationService *service.ClassificationService
	logger                *logger.Logger
}

func NewClassificationHandler(classificationService *service.ClassificationService, appLogger *logger.Logger) *ClassificationHandler {
	return &ClassificationHandler{
		classificationService: classificationService,
		logger:                appLogger.WithComponent("classification_handler"),
	}
}

// Request/Response types for device classification
type ClassifyDeviceRequest struct {
	Body struct {
		DeviceID   string `json:"device_id" doc:"Device ID to classify"`
		Layer      int    `json:"layer" doc:"Network layer (0-5)"`
		DeviceType string `json:"device_type" doc:"Device type (e.g., router, switch, server)"`
	}
}

type DeviceClassificationResponse struct {
	Body classification.DeviceClassification
}

type UnclassifiedDevicesResponse struct {
	Body struct {
		Devices []UnclassifiedDevice `json:"devices"`
		Count   int                  `json:"count"`
		Total   int                  `json:"total"`
		Limit   int                  `json:"limit"`
		Offset  int                  `json:"offset"`
	}
}

type UnclassifiedDevice struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Hardware string `json:"hardware"`
}

// Request/Response types for classification rules
type CreateRuleRequest struct {
	Body struct {
		Name          string                         `json:"name" doc:"Rule name"`
		Description   string                         `json:"description" doc:"Rule description"`
		LogicOperator string                         `json:"logic" doc:"Logic operator for multiple conditions (AND, OR)" default:"AND"`
		Conditions    []classification.RuleCondition `json:"conditions" doc:"Multiple conditions for the rule"`
		Layer         int                            `json:"layer" doc:"Target layer"`
		DeviceType    string                         `json:"device_type" doc:"Target device type"`
		Priority      int                            `json:"priority" doc:"Rule priority (higher = applied first)"`
		IsActive      bool                           `json:"is_active" doc:"Whether rule is active"`
	}
}

type UpdateRuleRequest struct {
	RuleID string `path:"rule_id" doc:"Rule ID"`
	Body   struct {
		Name          string                         `json:"name" doc:"Rule name"`
		Description   string                         `json:"description" doc:"Rule description"`
		LogicOperator string                         `json:"logic" doc:"Logic operator for multiple conditions (AND, OR)" default:"AND"`
		Conditions    []classification.RuleCondition `json:"conditions" doc:"Multiple conditions for the rule"`
		Layer         int                            `json:"layer" doc:"Target layer"`
		DeviceType    string                         `json:"device_type" doc:"Target device type"`
		Priority      int                            `json:"priority" doc:"Rule priority (higher = applied first)"`
		IsActive      bool                           `json:"is_active" doc:"Whether rule is active"`
	}
}

type ClassificationRuleResponse struct {
	Body classification.ClassificationRule
}

type ClassificationRulesResponse struct {
	Body struct {
		Rules []classification.ClassificationRule `json:"rules"`
		Count int                                 `json:"count"`
	}
}

type DeviceClassificationsResponse struct {
	Body struct {
		Classifications []classification.DeviceClassification `json:"classifications"`
		Count           int                                   `json:"count"`
	}
}

// Request/Response types for suggestions
type ClassificationSuggestionsResponse struct {
	Body struct {
		Suggestions []classification.ClassificationSuggestion `json:"suggestions"`
		Count       int                                       `json:"count"`
	}
}

type SuggestionActionRequest struct {
	Body struct {
		Action string `json:"action" doc:"Action to take (accept, reject)"`
	}
}

// Request/Response types for hierarchy layers
type HierarchyLayersResponse struct {
	Body struct {
		Layers []classification.HierarchyLayer `json:"layers"`
		Count  int                             `json:"count"`
	}
}

type HierarchyLayerResponse struct {
	Body classification.HierarchyLayer
}

type CreateHierarchyLayerRequest struct {
	Body struct {
		Name        string `json:"name" doc:"Layer name"`
		Description string `json:"description" doc:"Layer description"`
		Order       int    `json:"order" doc:"Display order"`
		Color       string `json:"color" doc:"Display color (hex format)"`
	}
}

type UpdateHierarchyLayerRequest struct {
	LayerID int `path:"layer_id" doc:"Layer ID"`
	Body    struct {
		Name        string `json:"name" doc:"Layer name"`
		Description string `json:"description" doc:"Layer description"`
		Order       int    `json:"order" doc:"Display order"`
		Color       string `json:"color" doc:"Display color (hex format)"`
	}
}

// RegisterClassificationRoutes registers all classification-related routes
func (h *ClassificationHandler) RegisterRoutes(api huma.API) {
	// Device classification endpoints
	huma.Register(api, huma.Operation{
		OperationID: "classify-device",
		Method:      http.MethodPost,
		Path:        "/api/v1/classification/devices",
		Summary:     "Manually classify a device",
		Description: "Assign a device to a specific network layer and type",
		Tags:        []string{"classification"},
	}, h.ClassifyDevice)

	huma.Register(api, huma.Operation{
		OperationID: "get-device-classification",
		Method:      http.MethodGet,
		Path:        "/api/v1/classification/devices/{device_id}",
		Summary:     "Get device classification",
		Description: "Retrieve classification information for a specific device",
		Tags:        []string{"classification"},
	}, h.GetDeviceClassification)

	huma.Register(api, huma.Operation{
		OperationID: "list-unclassified-devices",
		Method:      http.MethodGet,
		Path:        "/api/v1/classification/devices/unclassified",
		Summary:     "List unclassified devices",
		Description: "Get all devices that haven't been classified yet",
		Tags:        []string{"classification"},
	}, h.ListUnclassifiedDevices)

	huma.Register(api, huma.Operation{
		OperationID: "list-device-classifications",
		Method:      http.MethodGet,
		Path:        "/api/v1/classification/devices/classified",
		Summary:     "List all device classifications",
		Description: "Get all device classifications",
		Tags:        []string{"classification"},
	}, h.ListDeviceClassifications)

	huma.Register(api, huma.Operation{
		OperationID: "delete-device-classification",
		Method:      http.MethodDelete,
		Path:        "/api/v1/classification/devices/{device_id}",
		Summary:     "Delete device classification",
		Description: "Remove classification for a specific device",
		Tags:        []string{"classification"},
	}, h.DeleteDeviceClassification)

	// Classification rules endpoints
	huma.Register(api, huma.Operation{
		OperationID: "create-classification-rule",
		Method:      http.MethodPost,
		Path:        "/api/v1/classification/rules",
		Summary:     "Create a classification rule",
		Description: "Create a new rule for automatic device classification",
		Tags:        []string{"classification"},
	}, h.CreateClassificationRule)

	huma.Register(api, huma.Operation{
		OperationID: "list-classification-rules",
		Method:      http.MethodGet,
		Path:        "/api/v1/classification/rules",
		Summary:     "List classification rules",
		Description: "Get all classification rules",
		Tags:        []string{"classification"},
	}, h.ListClassificationRules)

	huma.Register(api, huma.Operation{
		OperationID: "update-classification-rule",
		Method:      http.MethodPut,
		Path:        "/api/v1/classification/rules/{rule_id}",
		Summary:     "Update a classification rule",
		Description: "Update an existing classification rule",
		Tags:        []string{"classification"},
	}, h.UpdateClassificationRule)

	huma.Register(api, huma.Operation{
		OperationID: "delete-classification-rule",
		Method:      http.MethodDelete,
		Path:        "/api/v1/classification/rules/{rule_id}",
		Summary:     "Delete a classification rule",
		Description: "Remove a classification rule",
		Tags:        []string{"classification"},
	}, h.DeleteClassificationRule)

	huma.Register(api, huma.Operation{
		OperationID: "apply-classification-rules",
		Method:      http.MethodPost,
		Path:        "/api/v1/classification/rules/apply",
		Summary:     "Apply classification rules",
		Description: "Apply all active rules to classify unclassified devices",
		Tags:        []string{"classification"},
	}, h.ApplyClassificationRules)

	// Suggestions endpoints
	huma.Register(api, huma.Operation{
		OperationID: "generate-rule-suggestions",
		Method:      http.MethodPost,
		Path:        "/api/v1/classification/suggestions/generate",
		Summary:     "Generate rule suggestions",
		Description: "Analyze manual classifications and generate suggested rules",
		Tags:        []string{"classification"},
	}, h.GenerateRuleSuggestions)

	huma.Register(api, huma.Operation{
		OperationID: "list-rule-suggestions",
		Method:      http.MethodGet,
		Path:        "/api/v1/classification/suggestions",
		Summary:     "List rule suggestions",
		Description: "Get all pending rule suggestions",
		Tags:        []string{"classification"},
	}, h.ListRuleSuggestions)

	huma.Register(api, huma.Operation{
		OperationID: "handle-suggestion",
		Method:      http.MethodPost,
		Path:        "/api/v1/classification/suggestions/{suggestion_id}/action",
		Summary:     "Accept or reject a rule suggestion",
		Description: "Accept or reject a classification rule suggestion",
		Tags:        []string{"classification"},
	}, h.HandleSuggestion)

	// Hierarchy layers endpoints
	huma.Register(api, huma.Operation{
		OperationID: "list-hierarchy-layers",
		Method:      http.MethodGet,
		Path:        "/api/v1/classification/layers",
		Summary:     "List hierarchy layers",
		Description: "Get all network hierarchy layer definitions",
		Tags:        []string{"classification"},
	}, h.ListHierarchyLayers)

	huma.Register(api, huma.Operation{
		OperationID: "get-hierarchy-layer",
		Method:      http.MethodGet,
		Path:        "/api/v1/classification/layers/{layer_id}",
		Summary:     "Get hierarchy layer",
		Description: "Get a specific hierarchy layer by ID",
		Tags:        []string{"classification"},
	}, h.GetHierarchyLayer)

	huma.Register(api, huma.Operation{
		OperationID: "create-hierarchy-layer",
		Method:      http.MethodPost,
		Path:        "/api/v1/classification/layers",
		Summary:     "Create hierarchy layer",
		Description: "Create a new network hierarchy layer",
		Tags:        []string{"classification"},
	}, h.CreateHierarchyLayer)

	huma.Register(api, huma.Operation{
		OperationID: "update-hierarchy-layer",
		Method:      http.MethodPut,
		Path:        "/api/v1/classification/layers/{layer_id}",
		Summary:     "Update hierarchy layer",
		Description: "Update an existing hierarchy layer",
		Tags:        []string{"classification"},
	}, h.UpdateHierarchyLayer)

	huma.Register(api, huma.Operation{
		OperationID: "delete-hierarchy-layer",
		Method:      http.MethodDelete,
		Path:        "/api/v1/classification/layers/{layer_id}",
		Summary:     "Delete hierarchy layer",
		Description: "Delete a hierarchy layer",
		Tags:        []string{"classification"},
	}, h.DeleteHierarchyLayer)
}

// Device classification handlers

func (h *ClassificationHandler) ClassifyDevice(ctx context.Context, req *ClassifyDeviceRequest) (*DeviceClassificationResponse, error) {
	// TODO: Get user ID from context/auth
	userID := "admin"

	err := h.classificationService.ClassifyDevice(ctx, req.Body.DeviceID, req.Body.Layer, req.Body.DeviceType, userID)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to classify device", err)
	}

	classification, err := h.classificationService.GetDeviceClassification(ctx, req.Body.DeviceID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve classification", err)
	}

	return &DeviceClassificationResponse{Body: *classification}, nil
}

func (h *ClassificationHandler) GetDeviceClassification(ctx context.Context, req *struct {
	DeviceID string `path:"device_id" doc:"Device ID"`
}) (*DeviceClassificationResponse, error) {
	classification, err := h.classificationService.GetDeviceClassification(ctx, req.DeviceID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get device classification", err)
	}
	if classification == nil {
		return nil, huma.Error404NotFound("Device classification not found")
	}

	return &DeviceClassificationResponse{Body: *classification}, nil
}

func (h *ClassificationHandler) ListUnclassifiedDevices(ctx context.Context, req *struct {
	Limit  int `query:"limit" doc:"Maximum number of devices to return (default: 100, max: 1000)" default:"100"`
	Offset int `query:"offset" doc:"Number of devices to skip (default: 0)" default:"0"`
}) (*UnclassifiedDevicesResponse, error) {
	// デフォルト値とバリデーション
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	devices, total, err := h.classificationService.ListUnclassifiedDevicesWithPagination(ctx, limit, offset)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list unclassified devices", err)
	}

	unclassifiedDevices := make([]UnclassifiedDevice, len(devices))
	for i, device := range devices {
		unclassifiedDevices[i] = UnclassifiedDevice{
			ID:       device.ID,
			Name:     device.ID, // DeviceにNameがないため、IDを使用
			Type:     device.Type,
			Hardware: device.Hardware,
		}
	}

	return &UnclassifiedDevicesResponse{
		Body: struct {
			Devices []UnclassifiedDevice `json:"devices"`
			Count   int                  `json:"count"`
			Total   int                  `json:"total"`
			Limit   int                  `json:"limit"`
			Offset  int                  `json:"offset"`
		}{
			Devices: unclassifiedDevices,
			Count:   len(unclassifiedDevices),
			Total:   total,
			Limit:   limit,
			Offset:  offset,
		},
	}, nil
}

func (h *ClassificationHandler) ListDeviceClassifications(ctx context.Context, req *struct{}) (*DeviceClassificationsResponse, error) {
	classifications, err := h.classificationService.ListDeviceClassifications(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list device classifications", err)
	}

	return &DeviceClassificationsResponse{
		Body: struct {
			Classifications []classification.DeviceClassification `json:"classifications"`
			Count           int                                   `json:"count"`
		}{
			Classifications: classifications,
			Count:           len(classifications),
		},
	}, nil
}

func (h *ClassificationHandler) DeleteDeviceClassification(ctx context.Context, req *struct {
	DeviceID string `path:"device_id" doc:"Device ID"`
}) (*struct{}, error) {
	err := h.classificationService.DeleteDeviceClassification(ctx, req.DeviceID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete device classification", err)
	}

	return &struct{}{}, nil
}

// Classification rules handlers

func (h *ClassificationHandler) CreateClassificationRule(ctx context.Context, req *CreateRuleRequest) (*ClassificationRuleResponse, error) {
	// バリデーション
	if len(req.Body.Conditions) == 0 {
		return nil, huma.Error400BadRequest("At least one condition is required")
	}

	logic := req.Body.LogicOperator
	if logic == "" {
		logic = "AND"
	}

	rule := classification.ClassificationRule{
		Name:          req.Body.Name,
		Description:   req.Body.Description,
		LogicOperator: logic,
		Conditions:    req.Body.Conditions,
		Layer:         req.Body.Layer,
		DeviceType:    req.Body.DeviceType,
		Priority:      req.Body.Priority,
		IsActive:      req.Body.IsActive,
		CreatedBy:     "admin", // TODO: Get from auth context
	}

	err := h.classificationService.SaveClassificationRule(ctx, rule)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create classification rule", err)
	}

	return &ClassificationRuleResponse{Body: rule}, nil
}

func (h *ClassificationHandler) UpdateClassificationRule(ctx context.Context, req *UpdateRuleRequest) (*ClassificationRuleResponse, error) {
	// バリデーション
	if len(req.Body.Conditions) == 0 {
		return nil, huma.Error400BadRequest("At least one condition is required")
	}

	logic := req.Body.LogicOperator
	if logic == "" {
		logic = "AND"
	}

	rule := classification.ClassificationRule{
		ID:            req.RuleID,
		Name:          req.Body.Name,
		Description:   req.Body.Description,
		LogicOperator: logic,
		Conditions:    req.Body.Conditions,
		Layer:         req.Body.Layer,
		DeviceType:    req.Body.DeviceType,
		Priority:      req.Body.Priority,
		IsActive:      req.Body.IsActive,
	}

	err := h.classificationService.UpdateClassificationRule(ctx, rule)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update classification rule", err)
	}

	// 更新されたルールを取得して返す
	updatedRule, err := h.classificationService.GetClassificationRule(ctx, req.RuleID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve updated rule", err)
	}
	if updatedRule == nil {
		return nil, huma.Error404NotFound("Rule not found")
	}

	return &ClassificationRuleResponse{Body: *updatedRule}, nil
}

func (h *ClassificationHandler) DeleteClassificationRule(ctx context.Context, req *struct {
	RuleID string `path:"rule_id" doc:"Rule ID"`
}) (*struct{}, error) {
	err := h.classificationService.DeleteClassificationRule(ctx, req.RuleID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete classification rule", err)
	}

	return &struct{}{}, nil
}

func (h *ClassificationHandler) ListClassificationRules(ctx context.Context, req *struct{}) (*ClassificationRulesResponse, error) {
	rules, err := h.classificationService.ListClassificationRules(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list classification rules", err)
	}

	return &ClassificationRulesResponse{
		Body: struct {
			Rules []classification.ClassificationRule `json:"rules"`
			Count int                                 `json:"count"`
		}{
			Rules: rules,
			Count: len(rules),
		},
	}, nil
}

func (h *ClassificationHandler) ApplyClassificationRules(ctx context.Context, req *struct{}) (*struct{}, error) {
	// Get all unclassified devices
	devices, err := h.classificationService.ListUnclassifiedDevices(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get unclassified devices", err)
	}

	deviceIDs := make([]string, len(devices))
	for i, device := range devices {
		deviceIDs[i] = device.ID
	}

	// Apply rules
	_, err = h.classificationService.ApplyClassificationRules(ctx, deviceIDs)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to apply classification rules", err)
	}

	return &struct{}{}, nil
}

// Suggestions handlers

func (h *ClassificationHandler) GenerateRuleSuggestions(ctx context.Context, req *struct{}) (*ClassificationSuggestionsResponse, error) {
	suggestions, err := h.classificationService.GenerateRuleSuggestions(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to generate rule suggestions", err)
	}

	return &ClassificationSuggestionsResponse{
		Body: struct {
			Suggestions []classification.ClassificationSuggestion `json:"suggestions"`
			Count       int                                       `json:"count"`
		}{
			Suggestions: suggestions,
			Count:       len(suggestions),
		},
	}, nil
}

func (h *ClassificationHandler) ListRuleSuggestions(ctx context.Context, req *struct{}) (*ClassificationSuggestionsResponse, error) {
	suggestions, err := h.classificationService.ListPendingSuggestions(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list rule suggestions", err)
	}

	return &ClassificationSuggestionsResponse{
		Body: struct {
			Suggestions []classification.ClassificationSuggestion `json:"suggestions"`
			Count       int                                       `json:"count"`
		}{
			Suggestions: suggestions,
			Count:       len(suggestions),
		},
	}, nil
}

func (h *ClassificationHandler) HandleSuggestion(ctx context.Context, req *struct {
	SuggestionID string `path:"suggestion_id" doc:"Suggestion ID"`
	Body         struct {
		Action string `json:"action" doc:"Action to take (accept, reject)"`
	}
}) (*struct{}, error) {
	switch req.Body.Action {
	case "accept":
		err := h.classificationService.AcceptSuggestion(ctx, req.SuggestionID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to accept suggestion", err)
		}
	case "reject":
		err := h.classificationService.RejectSuggestion(ctx, req.SuggestionID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to reject suggestion", err)
		}
	default:
		return nil, huma.Error400BadRequest("Invalid action", nil)
	}

	return &struct{}{}, nil
}

// Hierarchy layers handlers

func (h *ClassificationHandler) ListHierarchyLayers(ctx context.Context, req *struct{}) (*HierarchyLayersResponse, error) {
	h.logger.Info("Listing hierarchy layers")
	layers, err := h.classificationService.ListHierarchyLayers(ctx)
	if err != nil {
		h.logger.Error("Failed to list hierarchy layers", "error", err)
		return nil, huma.Error500InternalServerError("Failed to list hierarchy layers", err)
	}

	return &HierarchyLayersResponse{
		Body: struct {
			Layers []classification.HierarchyLayer `json:"layers"`
			Count  int                             `json:"count"`
		}{
			Layers: layers,
			Count:  len(layers),
		},
	}, nil
}

func (h *ClassificationHandler) GetHierarchyLayer(ctx context.Context, req *struct {
	LayerID int `path:"layer_id" doc:"Layer ID"`
}) (*HierarchyLayerResponse, error) {
	layer, err := h.classificationService.GetHierarchyLayer(ctx, req.LayerID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get hierarchy layer", err)
	}
	if layer == nil {
		return nil, huma.Error404NotFound("Hierarchy layer not found")
	}

	return &HierarchyLayerResponse{Body: *layer}, nil
}

func (h *ClassificationHandler) CreateHierarchyLayer(ctx context.Context, req *CreateHierarchyLayerRequest) (*HierarchyLayerResponse, error) {
	layer := classification.HierarchyLayer{
		Name:        req.Body.Name,
		Description: req.Body.Description,
		Order:       req.Body.Order,
		Color:       req.Body.Color,
	}

	err := h.classificationService.SaveHierarchyLayer(ctx, layer)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create hierarchy layer", err)
	}

	// Get the created layer to return the generated ID
	layers, err := h.classificationService.ListHierarchyLayers(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve created layer", err)
	}

	// Find the created layer (assuming it's the last one with matching name)
	var createdLayer *classification.HierarchyLayer
	for i := len(layers) - 1; i >= 0; i-- {
		if layers[i].Name == req.Body.Name {
			createdLayer = &layers[i]
			break
		}
	}

	if createdLayer == nil {
		return nil, huma.Error500InternalServerError("Failed to find created layer", nil)
	}

	return &HierarchyLayerResponse{Body: *createdLayer}, nil
}

func (h *ClassificationHandler) UpdateHierarchyLayer(ctx context.Context, req *UpdateHierarchyLayerRequest) (*HierarchyLayerResponse, error) {
	layer := classification.HierarchyLayer{
		ID:          req.LayerID,
		Name:        req.Body.Name,
		Description: req.Body.Description,
		Order:       req.Body.Order,
		Color:       req.Body.Color,
	}

	err := h.classificationService.UpdateHierarchyLayer(ctx, layer)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update hierarchy layer", err)
	}

	updatedLayer, err := h.classificationService.GetHierarchyLayer(ctx, req.LayerID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve updated layer", err)
	}

	return &HierarchyLayerResponse{Body: *updatedLayer}, nil
}

func (h *ClassificationHandler) DeleteHierarchyLayer(ctx context.Context, req *struct {
	LayerID int `path:"layer_id" doc:"Layer ID"`
}) (*struct{}, error) {
	err := h.classificationService.DeleteHierarchyLayer(ctx, req.LayerID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete hierarchy layer", err)
	}

	return &struct{}{}, nil
}
