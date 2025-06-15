package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/domain/topology"
)

type ClassificationService struct {
	classificationRepo classification.Repository
	topologyRepo       topology.Repository
}

func NewClassificationService(classificationRepo classification.Repository, topologyRepo topology.Repository) *ClassificationService {
	return &ClassificationService{
		classificationRepo: classificationRepo,
		topologyRepo:       topologyRepo,
	}
}

// ClassifyDevice manually classifies a device
func (s *ClassificationService) ClassifyDevice(ctx context.Context, deviceID string, layer int, deviceType string, userID string) error {
	// Verify device exists
	device, err := s.topologyRepo.GetDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	// Update device with classification information in new schema
	device.LayerID = &layer
	device.DeviceType = deviceType
	device.ClassifiedBy = fmt.Sprintf("user:%s", userID) // user:username format

	// Update the device in the topology repository
	return s.topologyRepo.UpdateDevice(ctx, *device)
}

// GetDeviceClassification retrieves classification for a specific device
func (s *ClassificationService) GetDeviceClassification(ctx context.Context, deviceID string) (*classification.DeviceClassification, error) {
	// Get device from topology repository
	device, err := s.topologyRepo.GetDevice(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return nil, nil // Device not found
	}

	// If device is not classified, return nil
	if device.LayerID == nil && device.DeviceType == "" && device.ClassifiedBy == "" {
		return nil, nil
	}

	// Construct DeviceClassification from device data
	layer := 0
	if device.LayerID != nil {
		layer = *device.LayerID
	}

	isManual := false
	createdBy := ""
	if strings.HasPrefix(device.ClassifiedBy, "user:") {
		isManual = true
		createdBy = strings.TrimPrefix(device.ClassifiedBy, "user:")
	}

	return &classification.DeviceClassification{
		ID:         device.ID, // Use device ID as classification ID
		DeviceID:   device.ID,
		Layer:      layer,
		DeviceType: device.DeviceType,
		IsManual:   isManual,
		CreatedBy:  createdBy,
		CreatedAt:  device.CreatedAt,
		UpdatedAt:  device.UpdatedAt,
	}, nil
}

// ListDeviceClassifications retrieves all device classifications from the new schema
func (s *ClassificationService) ListDeviceClassifications(ctx context.Context) ([]classification.DeviceClassification, error) {
	// Get all devices with classification information
	paginationOpts := topology.PaginationOptions{
		Page:     1,
		PageSize: 10000, // 大きめに取得
		OrderBy:  "id",
		SortDir:  "ASC",
	}

	allDevices, _, err := s.topologyRepo.GetDevices(ctx, paginationOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var classifications []classification.DeviceClassification
	for _, device := range allDevices {
		// Only include classified devices
		if device.LayerID != nil || device.DeviceType != "" || device.ClassifiedBy != "" {
			layer := 0
			if device.LayerID != nil {
				layer = *device.LayerID
			}

			isManual := false
			createdBy := ""
			if strings.HasPrefix(device.ClassifiedBy, "user:") {
				isManual = true
				createdBy = strings.TrimPrefix(device.ClassifiedBy, "user:")
			} else if strings.HasPrefix(device.ClassifiedBy, "rule:") {
				createdBy = "system"
			}

			classifications = append(classifications, classification.DeviceClassification{
				ID:         device.ID,
				DeviceID:   device.ID,
				Layer:      layer,
				DeviceType: device.DeviceType,
				IsManual:   isManual,
				CreatedBy:  createdBy,
				CreatedAt:  device.CreatedAt,
				UpdatedAt:  device.UpdatedAt,
			})
		}
	}

	return classifications, nil
}

// DeleteDeviceClassification removes classification for a specific device
func (s *ClassificationService) DeleteDeviceClassification(ctx context.Context, deviceID string) error {
	// Get device from topology repository
	device, err := s.topologyRepo.GetDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	// Clear classification fields
	device.LayerID = nil
	device.DeviceType = ""
	device.ClassifiedBy = ""

	// Update the device in the topology repository
	return s.topologyRepo.UpdateDevice(ctx, *device)
}

// ListUnclassifiedDevices returns devices that haven't been classified
func (s *ClassificationService) ListUnclassifiedDevices(ctx context.Context) ([]topology.Device, error) {
	// Get all devices from topology repository
	paginationOpts := topology.PaginationOptions{
		Page:     1,
		PageSize: 10000, // 大きめに取得
		OrderBy:  "id",
		SortDir:  "ASC",
	}

	allDevices, _, err := s.topologyRepo.GetDevices(ctx, paginationOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	// Filter unclassified devices
	var unclassifiedDevices []topology.Device
	for _, device := range allDevices {
		if s.isUnclassified(device) {
			unclassifiedDevices = append(unclassifiedDevices, device)
		}
	}

	return unclassifiedDevices, nil
}

// ListUnclassifiedDevicesWithPagination returns devices that haven't been classified with pagination
func (s *ClassificationService) ListUnclassifiedDevicesWithPagination(ctx context.Context, limit, offset int) ([]topology.Device, int, error) {
	// 新しいスキーマでは、layer_idがNULLまたはclassified_byがNULL/空のデバイスが未分類
	// 簡易実装: topologyリポジトリから全デバイスを取得してフィルタリング
	paginationOpts := topology.PaginationOptions{
		Page:     1,
		PageSize: 1000, // 大きめに取得してフィルタリング
		OrderBy:  "id",
		SortDir:  "ASC",
	}

	allDevices, _, err := s.topologyRepo.GetDevices(ctx, paginationOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get devices: %w", err)
	}

	// 未分類デバイスをフィルタリング
	var unclassifiedDevices []topology.Device
	for _, device := range allDevices {
		if s.isUnclassified(device) {
			unclassifiedDevices = append(unclassifiedDevices, device)
		}
	}

	// ページネーション適用
	totalCount := len(unclassifiedDevices)
	start := offset
	if start >= len(unclassifiedDevices) {
		return []topology.Device{}, totalCount, nil
	}

	end := start + limit
	if end > len(unclassifiedDevices) {
		end = len(unclassifiedDevices)
	}

	return unclassifiedDevices[start:end], totalCount, nil
}

// isUnclassified checks if a device is unclassified in the new schema
func (s *ClassificationService) isUnclassified(device topology.Device) bool {
	// layer_idがNULLまたはclassified_byがNULL/空の場合は未分類
	return device.LayerID == nil || device.ClassifiedBy == ""
}

// ApplyClassificationRules applies all active rules to classify devices
func (s *ClassificationService) ApplyClassificationRules(ctx context.Context, deviceIDs []string) ([]classification.DeviceClassification, error) {
	rules, err := s.classificationRepo.ListActiveClassificationRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active rules: %w", err)
	}

	var results []classification.DeviceClassification

	for _, deviceID := range deviceIDs {
		// Get device details
		device, err := s.topologyRepo.GetDevice(ctx, deviceID)
		if err != nil || device == nil {
			continue
		}

		// Skip if device is already manually classified (user: prefix)
		if strings.HasPrefix(device.ClassifiedBy, "user:") {
			continue
		}

		// Apply rules in priority order
		for _, rule := range rules {
			if s.deviceMatchesRule(*device, rule) {
				// Update device with classification information
				device.LayerID = &rule.Layer
				device.DeviceType = rule.DeviceType
				device.ClassifiedBy = fmt.Sprintf("rule:%s", rule.Name)

				// Update device in topology repository
				if err := s.topologyRepo.UpdateDevice(ctx, *device); err == nil {
					// Create result object for return
					classification := classification.DeviceClassification{
						ID:         device.ID,
						DeviceID:   deviceID,
						Layer:      rule.Layer,
						DeviceType: rule.DeviceType,
						IsManual:   false,
						CreatedBy:  "system",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}
					results = append(results, classification)
				}
				break // Apply only the first matching rule
			}
		}
	}

	return results, nil
}

// deviceMatchesRule checks if a device matches a classification rule
func (s *ClassificationService) deviceMatchesRule(device topology.Device, rule classification.ClassificationRule) bool {
	if len(rule.Conditions) == 0 {
		return false
	}

	var results []bool
	for _, condition := range rule.Conditions {
		results = append(results, s.deviceMatchesCondition(device, condition))
	}

	// Apply logic operator
	if rule.LogicOperator == "OR" {
		// OR: at least one condition must be true
		for _, result := range results {
			if result {
				return true
			}
		}
		return false
	} else {
		// AND (default): all conditions must be true
		for _, result := range results {
			if !result {
				return false
			}
		}
		return true
	}
}

// deviceMatchesCondition checks if a device matches a single condition
func (s *ClassificationService) deviceMatchesCondition(device topology.Device, condition classification.RuleCondition) bool {
	var fieldValue string

	switch condition.Field {
	case "name":
		fieldValue = device.ID // DeviceにNameがないため、IDを使用
	case "hardware":
		fieldValue = device.Hardware
	case "type":
		fieldValue = device.Type
	default:
		return false
	}

	switch condition.Operator {
	case "contains":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(condition.Value))
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(fieldValue), strings.ToLower(condition.Value))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(fieldValue), strings.ToLower(condition.Value))
	case "equals":
		return strings.EqualFold(fieldValue, condition.Value)
	case "regex":
		if re, err := regexp.Compile(condition.Value); err == nil {
			return re.MatchString(fieldValue)
		}
		return false
	default:
		return false
	}
}

// GenerateRuleSuggestions analyzes manual classifications and suggests new rules
func (s *ClassificationService) GenerateRuleSuggestions(ctx context.Context) ([]classification.ClassificationSuggestion, error) {
	manualClassifications, err := s.getManualClassifications(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get manual classifications: %w", err)
	}

	if len(manualClassifications) < 2 {
		return []classification.ClassificationSuggestion{}, nil
	}

	var suggestions []classification.ClassificationSuggestion

	// Group by layer and device type
	groups := s.groupClassificationsByLayerAndType(manualClassifications)

	for key, group := range groups {
		if len(group) < 2 { // Need at least 2 devices to suggest a pattern
			continue
		}

		// Analyze naming patterns
		nameSuggestions := s.analyzeNamePatterns(ctx, group, key)
		suggestions = append(suggestions, nameSuggestions...)

		// Analyze hardware patterns
		hardwareSuggestions := s.analyzeHardwarePatterns(ctx, group, key)
		suggestions = append(suggestions, hardwareSuggestions...)
	}

	return suggestions, nil
}

// getManualClassifications retrieves all manual device classifications with device details
func (s *ClassificationService) getManualClassifications(ctx context.Context) ([]classificationWithDevice, error) {
	// Get all devices from topology repository
	paginationOpts := topology.PaginationOptions{
		Page:     1,
		PageSize: 10000, // 大きめに取得
		OrderBy:  "id",
		SortDir:  "ASC",
	}

	allDevices, _, err := s.topologyRepo.GetDevices(ctx, paginationOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var result []classificationWithDevice
	for _, device := range allDevices {
		// Only include devices with manual classification (user: prefix)
		if strings.HasPrefix(device.ClassifiedBy, "user:") {
			layer := 0
			if device.LayerID != nil {
				layer = *device.LayerID
			}

			createdBy := strings.TrimPrefix(device.ClassifiedBy, "user:")
			classification := classification.DeviceClassification{
				ID:         device.ID,
				DeviceID:   device.ID,
				Layer:      layer,
				DeviceType: device.DeviceType,
				IsManual:   true,
				CreatedBy:  createdBy,
				CreatedAt:  device.CreatedAt,
				UpdatedAt:  device.UpdatedAt,
			}

			result = append(result, classificationWithDevice{
				Classification: classification,
				Device:         device,
			})
		}
	}

	return result, nil
}

type classificationWithDevice struct {
	Classification classification.DeviceClassification
	Device         topology.Device
}

type classificationGroupKey struct {
	Layer      int
	DeviceType string
}

func (s *ClassificationService) groupClassificationsByLayerAndType(classifications []classificationWithDevice) map[classificationGroupKey][]classificationWithDevice {
	groups := make(map[classificationGroupKey][]classificationWithDevice)

	for _, c := range classifications {
		key := classificationGroupKey{
			Layer:      c.Classification.Layer,
			DeviceType: c.Classification.DeviceType,
		}
		groups[key] = append(groups[key], c)
	}

	return groups
}

// analyzeNamePatterns analyzes device names to suggest naming pattern rules
func (s *ClassificationService) analyzeNamePatterns(ctx context.Context, group []classificationWithDevice, key classificationGroupKey) []classification.ClassificationSuggestion {
	var suggestions []classification.ClassificationSuggestion

	deviceNames := make([]string, len(group))
	deviceIDs := make([]string, len(group))
	for i, c := range group {
		deviceNames[i] = c.Device.ID // DeviceにNameがないため、IDを使用
		deviceIDs[i] = c.Device.ID
	}

	// Find common prefixes
	commonPrefixes := s.findCommonPrefixes(deviceNames)
	for _, prefix := range commonPrefixes {
		if len(prefix) >= 2 { // Minimum prefix length
			confidence := s.calculateConfidence(prefix, deviceNames, "starts_with")
			if confidence >= 0.7 { // Minimum confidence threshold
				rule := classification.ClassificationRule{
					ID:            uuid.New().String(),
					Name:          fmt.Sprintf("Auto: Names starting with '%s'", prefix),
					Description:   fmt.Sprintf("Devices with names starting with '%s' should be classified as %s layer %d", prefix, key.DeviceType, key.Layer),
					LogicOperator: "AND",
					Conditions: []classification.RuleCondition{
						{
							Field:    "name",
							Operator: "starts_with",
							Value:    prefix,
						},
					},
					Layer:      key.Layer,
					DeviceType: key.DeviceType,
					Priority:   100,
					IsActive:   false, // Suggestions start as inactive
					Confidence: confidence,
					CreatedBy:  "system",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}

				suggestion := classification.ClassificationSuggestion{
					ID:              uuid.New().String(),
					Rule:            rule,
					AffectedDevices: s.findAffectedDevicesByRule(ctx, rule),
					BasedOnDevices:  deviceIDs,
					Confidence:      confidence,
					Status:          classification.SuggestionStatusPending,
					CreatedAt:       time.Now(),
				}

				suggestions = append(suggestions, suggestion)
			}
		}
	}

	// Find common keywords
	keywords := s.findCommonKeywords(deviceNames)
	for _, keyword := range keywords {
		if len(keyword) >= 2 && keyword != strings.ToLower(keyword) { // Skip single chars and already lowercase
			confidence := s.calculateConfidence(keyword, deviceNames, "contains")
			if confidence >= 0.7 {
				rule := classification.ClassificationRule{
					ID:            uuid.New().String(),
					Name:          fmt.Sprintf("Auto: Names containing '%s'", keyword),
					Description:   fmt.Sprintf("Devices with names containing '%s' should be classified as %s layer %d", keyword, key.DeviceType, key.Layer),
					LogicOperator: "AND",
					Conditions: []classification.RuleCondition{
						{
							Field:    "name",
							Operator: "contains",
							Value:    keyword,
						},
					},
					Layer:      key.Layer,
					DeviceType: key.DeviceType,
					Priority:   90,
					IsActive:   false,
					Confidence: confidence,
					CreatedBy:  "system",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}

				suggestion := classification.ClassificationSuggestion{
					ID:              uuid.New().String(),
					Rule:            rule,
					AffectedDevices: s.findAffectedDevicesByRule(ctx, rule),
					BasedOnDevices:  deviceIDs,
					Confidence:      confidence,
					Status:          classification.SuggestionStatusPending,
					CreatedAt:       time.Now(),
				}

				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions
}

// analyzeHardwarePatterns analyzes device hardware to suggest hardware-based rules
func (s *ClassificationService) analyzeHardwarePatterns(ctx context.Context, group []classificationWithDevice, key classificationGroupKey) []classification.ClassificationSuggestion {
	var suggestions []classification.ClassificationSuggestion

	hardwareMap := make(map[string][]string) // hardware -> device IDs
	for _, c := range group {
		hardwareMap[c.Device.Hardware] = append(hardwareMap[c.Device.Hardware], c.Device.ID)
	}

	for hardware, deviceIDs := range hardwareMap {
		if len(deviceIDs) >= 2 { // Need at least 2 devices with same hardware
			confidence := float64(len(deviceIDs)) / float64(len(group))
			if confidence >= 0.5 {
				rule := classification.ClassificationRule{
					ID:            uuid.New().String(),
					Name:          fmt.Sprintf("Auto: Hardware equals '%s'", hardware),
					Description:   fmt.Sprintf("Devices with hardware '%s' should be classified as %s layer %d", hardware, key.DeviceType, key.Layer),
					LogicOperator: "AND",
					Conditions: []classification.RuleCondition{
						{
							Field:    "hardware",
							Operator: "equals",
							Value:    hardware,
						},
					},
					Layer:      key.Layer,
					DeviceType: key.DeviceType,
					Priority:   80,
					IsActive:   false,
					Confidence: confidence,
					CreatedBy:  "system",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}

				suggestion := classification.ClassificationSuggestion{
					ID:              uuid.New().String(),
					Rule:            rule,
					AffectedDevices: s.findAffectedDevicesByRule(ctx, rule),
					BasedOnDevices:  deviceIDs,
					Confidence:      confidence,
					Status:          classification.SuggestionStatusPending,
					CreatedAt:       time.Now(),
				}

				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions
}

// Helper functions for pattern analysis

func (s *ClassificationService) findCommonPrefixes(names []string) []string {
	if len(names) < 2 {
		return []string{}
	}

	prefixCount := make(map[string]int)

	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			prefix := s.longestCommonPrefix(names[i], names[j])
			if len(prefix) >= 2 {
				prefixCount[prefix]++
			}
		}
	}

	var result []string
	for prefix, count := range prefixCount {
		if count >= len(names)/2 { // Prefix appears in at least half of the pairs
			result = append(result, prefix)
		}
	}

	return result
}

func (s *ClassificationService) findCommonKeywords(names []string) []string {
	wordCount := make(map[string]int)

	for _, name := range names {
		words := s.extractWords(name)
		for _, word := range words {
			wordCount[strings.ToLower(word)]++
		}
	}

	var result []string
	threshold := len(names) * 2 / 3 // Word must appear in at least 2/3 of names
	for word, count := range wordCount {
		if count >= threshold && len(word) >= 2 {
			result = append(result, word)
		}
	}

	return result
}

func (s *ClassificationService) longestCommonPrefix(s1, s2 string) string {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return s1[:i]
		}
	}
	return s1[:minLen]
}

func (s *ClassificationService) extractWords(name string) []string {
	// Split by common delimiters
	delimiters := []string{"-", "_", ".", " "}
	words := []string{name}

	for _, delimiter := range delimiters {
		var newWords []string
		for _, word := range words {
			newWords = append(newWords, strings.Split(word, delimiter)...)
		}
		words = newWords
	}

	// Filter out empty strings
	var result []string
	for _, word := range words {
		if len(word) > 0 {
			result = append(result, word)
		}
	}

	return result
}

func (s *ClassificationService) calculateConfidence(pattern string, names []string, operator string) float64 {
	matches := 0
	for _, name := range names {
		switch operator {
		case "starts_with":
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(pattern)) {
				matches++
			}
		case "contains":
			if strings.Contains(strings.ToLower(name), strings.ToLower(pattern)) {
				matches++
			}
		}
	}
	return float64(matches) / float64(len(names))
}

func (s *ClassificationService) findAffectedDevicesByRule(ctx context.Context, rule classification.ClassificationRule) []string {
	// This would need to query all devices and apply the rule
	// For now, return empty slice - implement when needed
	return []string{}
}

// SaveClassificationRule saves a new or updated classification rule
func (s *ClassificationService) SaveClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
		rule.CreatedAt = time.Now()
	}
	rule.UpdatedAt = time.Now()
	return s.classificationRepo.SaveClassificationRule(ctx, rule)
}

// GetClassificationRule retrieves a specific classification rule
func (s *ClassificationService) GetClassificationRule(ctx context.Context, ruleID string) (*classification.ClassificationRule, error) {
	return s.classificationRepo.GetClassificationRule(ctx, ruleID)
}

// UpdateClassificationRule updates an existing classification rule
func (s *ClassificationService) UpdateClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	rule.UpdatedAt = time.Now()
	return s.classificationRepo.UpdateClassificationRule(ctx, rule)
}

// DeleteClassificationRule deletes a classification rule
func (s *ClassificationService) DeleteClassificationRule(ctx context.Context, ruleID string) error {
	return s.classificationRepo.DeleteClassificationRule(ctx, ruleID)
}

// ListClassificationRules lists all classification rules
func (s *ClassificationService) ListClassificationRules(ctx context.Context) ([]classification.ClassificationRule, error) {
	return s.classificationRepo.ListClassificationRules(ctx)
}

// AcceptSuggestion accepts a classification suggestion and creates an active rule
func (s *ClassificationService) AcceptSuggestion(ctx context.Context, suggestionID string) error {
	suggestion, err := s.classificationRepo.GetClassificationSuggestion(ctx, suggestionID)
	if err != nil {
		return fmt.Errorf("failed to get suggestion: %w", err)
	}
	if suggestion == nil {
		return fmt.Errorf("suggestion not found: %s", suggestionID)
	}

	// Activate the rule
	rule := suggestion.Rule
	rule.IsActive = true
	if err := s.classificationRepo.SaveClassificationRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to save rule: %w", err)
	}

	// Update suggestion status
	if err := s.classificationRepo.UpdateClassificationSuggestionStatus(ctx, suggestionID, classification.SuggestionStatusAccepted); err != nil {
		return fmt.Errorf("failed to update suggestion status: %w", err)
	}

	return nil
}

// RejectSuggestion rejects a classification suggestion
func (s *ClassificationService) RejectSuggestion(ctx context.Context, suggestionID string) error {
	return s.classificationRepo.UpdateClassificationSuggestionStatus(ctx, suggestionID, classification.SuggestionStatusRejected)
}

// ListPendingSuggestions lists all pending classification suggestions
func (s *ClassificationService) ListPendingSuggestions(ctx context.Context) ([]classification.ClassificationSuggestion, error) {
	return s.classificationRepo.ListPendingClassificationSuggestions(ctx)
}

// Hierarchy Layer management

// GetHierarchyLayer retrieves a specific hierarchy layer
func (s *ClassificationService) GetHierarchyLayer(ctx context.Context, layerID int) (*classification.HierarchyLayer, error) {
	return s.classificationRepo.GetHierarchyLayer(ctx, layerID)
}

// ListHierarchyLayers retrieves all hierarchy layers
func (s *ClassificationService) ListHierarchyLayers(ctx context.Context) ([]classification.HierarchyLayer, error) {
	return s.classificationRepo.ListHierarchyLayers(ctx)
}

// SaveHierarchyLayer creates or updates a hierarchy layer
func (s *ClassificationService) SaveHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	// Generate ID if not provided (for new layers)
	if layer.ID == 0 {
		// Find the next available ID
		layers, err := s.classificationRepo.ListHierarchyLayers(ctx)
		if err != nil {
			return fmt.Errorf("failed to list layers for ID generation: %w", err)
		}

		maxID := -1
		for _, l := range layers {
			if l.ID > maxID {
				maxID = l.ID
			}
		}
		layer.ID = maxID + 1
	}

	return s.classificationRepo.SaveHierarchyLayer(ctx, layer)
}

// UpdateHierarchyLayer updates an existing hierarchy layer
func (s *ClassificationService) UpdateHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	// Verify layer exists
	existing, err := s.classificationRepo.GetHierarchyLayer(ctx, layer.ID)
	if err != nil {
		return fmt.Errorf("failed to check existing layer: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("layer with ID %d not found", layer.ID)
	}

	return s.classificationRepo.UpdateHierarchyLayer(ctx, layer)
}

// DeleteHierarchyLayer deletes a hierarchy layer
func (s *ClassificationService) DeleteHierarchyLayer(ctx context.Context, layerID int) error {
	// Check if layer is being used by any classifications
	classifications, err := s.classificationRepo.ListDeviceClassifications(ctx)
	if err != nil {
		return fmt.Errorf("failed to check layer usage: %w", err)
	}

	for _, c := range classifications {
		if c.Layer == layerID {
			return fmt.Errorf("cannot delete layer %d: it is currently being used by device classifications", layerID)
		}
	}

	return s.classificationRepo.DeleteHierarchyLayer(ctx, layerID)
}
