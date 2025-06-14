package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/servak/topology-manager/internal/domain/classification"
)

// ClassificationRepository provides in-memory implementation for testing
type ClassificationRepository struct {
	mu                 sync.RWMutex
	classifications    map[string]classification.DeviceClassification
	rules              map[string]classification.ClassificationRule
	suggestions        map[string]classification.ClassificationSuggestion
	layers             map[int]classification.HierarchyLayer
	unclassifiedDevices []string
}

func NewClassificationRepository() *ClassificationRepository {
	repo := &ClassificationRepository{
		classifications: make(map[string]classification.DeviceClassification),
		rules:          make(map[string]classification.ClassificationRule),
		suggestions:    make(map[string]classification.ClassificationSuggestion),
		layers:         make(map[int]classification.HierarchyLayer),
		unclassifiedDevices: []string{
			"fw1.edge", "sw1.core", "srv1.web",
			"router1.main", "switch1.access", "server1.db", 
			"firewall1.dmz", "ap1.wifi", "lb1.frontend",
			"router2.backup", "switch2.mgmt", "server2.app"
		}, // テスト用
	}

	// デフォルト階層を設定
	for _, layer := range classification.DefaultHierarchyLayers() {
		repo.layers[layer.ID] = layer
	}

	return repo
}

// Device Classifications
func (r *ClassificationRepository) GetDeviceClassification(ctx context.Context, deviceID string) (*classification.DeviceClassification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if c, exists := r.classifications[deviceID]; exists {
		return &c, nil
	}
	return nil, nil
}

func (r *ClassificationRepository) ListDeviceClassifications(ctx context.Context) ([]classification.DeviceClassification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]classification.DeviceClassification, 0, len(r.classifications))
	for _, c := range r.classifications {
		result = append(result, c)
	}
	return result, nil
}

func (r *ClassificationRepository) ListUnclassifiedDevices(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 分類済みデバイスを除外
	var unclassified []string
	for _, deviceID := range r.unclassifiedDevices {
		if _, exists := r.classifications[deviceID]; !exists {
			unclassified = append(unclassified, deviceID)
		}
	}
	return unclassified, nil
}

func (r *ClassificationRepository) SaveDeviceClassification(ctx context.Context, classification classification.DeviceClassification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if classification.CreatedAt.IsZero() {
		classification.CreatedAt = time.Now()
	}
	classification.UpdatedAt = time.Now()

	r.classifications[classification.DeviceID] = classification
	return nil
}

func (r *ClassificationRepository) DeleteDeviceClassification(ctx context.Context, deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.classifications, deviceID)
	return nil
}

// Classification Rules
func (r *ClassificationRepository) GetClassificationRule(ctx context.Context, ruleID string) (*classification.ClassificationRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if rule, exists := r.rules[ruleID]; exists {
		return &rule, nil
	}
	return nil, fmt.Errorf("rule not found: %s", ruleID)
}

func (r *ClassificationRepository) ListClassificationRules(ctx context.Context) ([]classification.ClassificationRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]classification.ClassificationRule, 0, len(r.rules))
	for _, rule := range r.rules {
		result = append(result, rule)
	}
	return result, nil
}

func (r *ClassificationRepository) ListActiveClassificationRules(ctx context.Context) ([]classification.ClassificationRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []classification.ClassificationRule
	for _, rule := range r.rules {
		if rule.IsActive {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (r *ClassificationRepository) SaveClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	rule.UpdatedAt = time.Now()

	r.rules[rule.ID] = rule
	return nil
}

func (r *ClassificationRepository) UpdateClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	rule.UpdatedAt = time.Now()
	r.rules[rule.ID] = rule
	return nil
}

func (r *ClassificationRepository) DeleteClassificationRule(ctx context.Context, ruleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.rules, ruleID)
	return nil
}

// Classification Suggestions
func (r *ClassificationRepository) GetClassificationSuggestion(ctx context.Context, suggestionID string) (*classification.ClassificationSuggestion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if suggestion, exists := r.suggestions[suggestionID]; exists {
		return &suggestion, nil
	}
	return nil, fmt.Errorf("suggestion not found: %s", suggestionID)
}

func (r *ClassificationRepository) ListPendingClassificationSuggestions(ctx context.Context) ([]classification.ClassificationSuggestion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []classification.ClassificationSuggestion
	for _, suggestion := range r.suggestions {
		if suggestion.Status == classification.SuggestionStatusPending {
			result = append(result, suggestion)
		}
	}
	return result, nil
}

func (r *ClassificationRepository) SaveClassificationSuggestion(ctx context.Context, suggestion classification.ClassificationSuggestion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if suggestion.CreatedAt.IsZero() {
		suggestion.CreatedAt = time.Now()
	}

	r.suggestions[suggestion.ID] = suggestion
	return nil
}

func (r *ClassificationRepository) UpdateClassificationSuggestionStatus(ctx context.Context, suggestionID string, status classification.SuggestionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	suggestion, exists := r.suggestions[suggestionID]
	if !exists {
		return fmt.Errorf("suggestion not found: %s", suggestionID)
	}

	suggestion.Status = status
	r.suggestions[suggestionID] = suggestion
	return nil
}

func (r *ClassificationRepository) DeleteClassificationSuggestion(ctx context.Context, suggestionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.suggestions, suggestionID)
	return nil
}

// Hierarchy Layers
func (r *ClassificationRepository) GetHierarchyLayer(ctx context.Context, layerID int) (*classification.HierarchyLayer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if layer, exists := r.layers[layerID]; exists {
		return &layer, nil
	}
	return nil, fmt.Errorf("layer not found: %d", layerID)
}

func (r *ClassificationRepository) ListHierarchyLayers(ctx context.Context) ([]classification.HierarchyLayer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]classification.HierarchyLayer, 0, len(r.layers))
	for _, layer := range r.layers {
		result = append(result, layer)
	}
	return result, nil
}

func (r *ClassificationRepository) SaveHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if layer.CreatedAt.IsZero() {
		layer.CreatedAt = time.Now()
	}
	layer.UpdatedAt = time.Now()

	r.layers[layer.ID] = layer
	return nil
}

func (r *ClassificationRepository) UpdateHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.layers[layer.ID]; !exists {
		return fmt.Errorf("layer not found: %d", layer.ID)
	}

	layer.UpdatedAt = time.Now()
	r.layers[layer.ID] = layer
	return nil
}

func (r *ClassificationRepository) DeleteHierarchyLayer(ctx context.Context, layerID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.layers, layerID)
	return nil
}

func (r *ClassificationRepository) Close() error {
	return nil
}