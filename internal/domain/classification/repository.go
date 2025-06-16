package classification

import "context"

// Repository defines the interface for device classification data access
type Repository interface {
	// Device Classifications
	GetDeviceClassification(ctx context.Context, deviceID string) (*DeviceClassification, error)
	ListDeviceClassifications(ctx context.Context) ([]DeviceClassification, error)
	ListUnclassifiedDevices(ctx context.Context) ([]string, error)
	ListUnclassifiedDevicesWithPagination(ctx context.Context, limit, offset int) ([]string, error)
	CountUnclassifiedDevices(ctx context.Context) (int, error)
	SaveDeviceClassification(ctx context.Context, classification DeviceClassification) error
	DeleteDeviceClassification(ctx context.Context, deviceID string) error

	// Classification Rules
	GetClassificationRule(ctx context.Context, ruleID string) (*ClassificationRule, error)
	ListClassificationRules(ctx context.Context) ([]ClassificationRule, error)
	ListActiveClassificationRules(ctx context.Context) ([]ClassificationRule, error)
	SaveClassificationRule(ctx context.Context, rule ClassificationRule) error
	UpdateClassificationRule(ctx context.Context, rule ClassificationRule) error
	DeleteClassificationRule(ctx context.Context, ruleID string) error

	// Classification Suggestions
	GetClassificationSuggestion(ctx context.Context, suggestionID string) (*ClassificationSuggestion, error)
	ListPendingClassificationSuggestions(ctx context.Context) ([]ClassificationSuggestion, error)
	SaveClassificationSuggestion(ctx context.Context, suggestion ClassificationSuggestion) error
	UpdateClassificationSuggestionStatus(ctx context.Context, suggestionID string, status SuggestionStatus) error
	DeleteClassificationSuggestion(ctx context.Context, suggestionID string) error

	// Hierarchy Layers
	GetHierarchyLayer(ctx context.Context, layerID int) (*HierarchyLayer, error)
	ListHierarchyLayers(ctx context.Context) ([]HierarchyLayer, error)
	SaveHierarchyLayer(ctx context.Context, layer HierarchyLayer) error
	UpdateHierarchyLayer(ctx context.Context, layer HierarchyLayer) error
	DeleteHierarchyLayer(ctx context.Context, layerID int) error

	// Utilities
	Close() error
}
