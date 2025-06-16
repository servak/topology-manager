package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/servak/topology-manager/internal/domain/classification"
)

// Classification repository methods for SQLite

// SaveDeviceClassification saves a device classification (not used in new schema)
func (r *sqliteRepository) SaveDeviceClassification(ctx context.Context, dc classification.DeviceClassification) error {
	// In the new schema, classification data is stored in the devices table
	// This method is kept for backward compatibility but not actively used
	return fmt.Errorf("SaveDeviceClassification is deprecated - use UpdateDevice instead")
}

// GetDeviceClassification retrieves classification for a specific device
func (r *sqliteRepository) GetDeviceClassification(ctx context.Context, deviceID string) (*classification.DeviceClassification, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, created_at, updated_at
		FROM devices
		WHERE id = ?`

	var id, deviceType, hardware string
	var layerID *int
	var classifiedBy sql.NullString
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&id, &deviceType, &hardware, &layerID, &deviceType, &classifiedBy, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// If device is not classified, return nil
	if layerID == nil && deviceType == "" && !classifiedBy.Valid {
		return nil, nil
	}

	layer := 0
	if layerID != nil {
		layer = *layerID
	}

	isManual := false
	createdBy := ""
	if classifiedBy.Valid {
		if classifiedBy.String[:5] == "user:" {
			isManual = true
			createdBy = classifiedBy.String[5:]
		} else if classifiedBy.String[:5] == "rule:" {
			createdBy = "system"
		}
	}

	return &classification.DeviceClassification{
		ID:         id,
		DeviceID:   deviceID,
		Layer:      layer,
		DeviceType: deviceType,
		IsManual:   isManual,
		CreatedBy:  createdBy,
		// CreatedAt and UpdatedAt would need proper time parsing
	}, nil
}

// ListDeviceClassifications retrieves all device classifications
func (r *sqliteRepository) ListDeviceClassifications(ctx context.Context) ([]classification.DeviceClassification, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, created_at, updated_at
		FROM devices
		WHERE layer_id IS NOT NULL OR device_type != '' OR classified_by IS NOT NULL
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classifications []classification.DeviceClassification
	for rows.Next() {
		var id, deviceType, hardware string
		var layerID *int
		var classifiedBy sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(&id, &deviceType, &hardware, &layerID, &deviceType, &classifiedBy, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		layer := 0
		if layerID != nil {
			layer = *layerID
		}

		isManual := false
		createdByStr := ""
		if classifiedBy.Valid {
			if len(classifiedBy.String) > 5 && classifiedBy.String[:5] == "user:" {
				isManual = true
				createdByStr = classifiedBy.String[5:]
			} else if len(classifiedBy.String) > 5 && classifiedBy.String[:5] == "rule:" {
				createdByStr = "system"
			}
		}

		classifications = append(classifications, classification.DeviceClassification{
			ID:         id,
			DeviceID:   id,
			Layer:      layer,
			DeviceType: deviceType,
			IsManual:   isManual,
			CreatedBy:  createdByStr,
		})
	}

	return classifications, rows.Err()
}

// DeleteDeviceClassification removes classification for a specific device (not used in new schema)
func (r *sqliteRepository) DeleteDeviceClassification(ctx context.Context, deviceID string) error {
	// In the new schema, we clear the classification fields in the devices table
	query := `
		UPDATE devices SET
			layer_id = NULL,
			device_type = '',
			classified_by = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("device with ID %s not found", deviceID)
	}

	return nil
}

// ListUnclassifiedDevices returns device IDs that haven't been classified
func (r *sqliteRepository) ListUnclassifiedDevices(ctx context.Context) ([]string, error) {
	query := `
		SELECT id
		FROM devices
		WHERE layer_id IS NULL OR classified_by IS NULL OR classified_by = ''
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		deviceIDs = append(deviceIDs, id)
	}

	return deviceIDs, rows.Err()
}

// CountUnclassifiedDevices returns the count of unclassified devices
func (r *sqliteRepository) CountUnclassifiedDevices(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM devices
		WHERE layer_id IS NULL OR classified_by IS NULL OR classified_by = ''`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// ListUnclassifiedDevicesWithPagination returns unclassified device IDs with pagination
func (r *sqliteRepository) ListUnclassifiedDevicesWithPagination(ctx context.Context, limit, offset int) ([]string, error) {
	query := `
		SELECT id
		FROM devices
		WHERE layer_id IS NULL OR classified_by IS NULL OR classified_by = ''
		ORDER BY id
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		deviceIDs = append(deviceIDs, id)
	}

	return deviceIDs, rows.Err()
}

// Classification Rules methods

// SaveClassificationRule saves a classification rule
func (r *sqliteRepository) SaveClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		INSERT INTO classification_rules (id, name, description, conditions, logic_operator, layer, device_type, priority, is_active, confidence, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			conditions = EXCLUDED.conditions,
			logic_operator = EXCLUDED.logic_operator,
			layer = EXCLUDED.layer,
			device_type = EXCLUDED.device_type,
			priority = EXCLUDED.priority,
			is_active = EXCLUDED.is_active,
			confidence = EXCLUDED.confidence,
			updated_at = CURRENT_TIMESTAMP`

	_, err = r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, string(conditionsJSON), rule.LogicOperator,
		rule.Layer, rule.DeviceType, rule.Priority, rule.IsActive, rule.Confidence,
		rule.CreatedBy, rule.CreatedAt, rule.UpdatedAt)

	return err
}

// GetClassificationRule retrieves a specific classification rule
func (r *sqliteRepository) GetClassificationRule(ctx context.Context, ruleID string) (*classification.ClassificationRule, error) {
	var rule classification.ClassificationRule
	var conditionsJSON string

	query := `
		SELECT id, name, description, conditions, logic_operator, layer, device_type, priority, is_active, confidence, created_by, created_at, updated_at
		FROM classification_rules
		WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, ruleID).Scan(
		&rule.ID, &rule.Name, &rule.Description, &conditionsJSON, &rule.LogicOperator,
		&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.Confidence,
		&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal conditions
	if err := json.Unmarshal([]byte(conditionsJSON), &rule.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	return &rule, nil
}

// UpdateClassificationRule updates an existing classification rule
func (r *sqliteRepository) UpdateClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		UPDATE classification_rules SET
			name = ?,
			description = ?,
			conditions = ?,
			logic_operator = ?,
			layer = ?,
			device_type = ?,
			priority = ?,
			is_active = ?,
			confidence = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		rule.Name, rule.Description, string(conditionsJSON), rule.LogicOperator,
		rule.Layer, rule.DeviceType, rule.Priority, rule.IsActive, rule.Confidence,
		rule.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("classification rule with ID %s not found", rule.ID)
	}

	return nil
}

// DeleteClassificationRule deletes a classification rule
func (r *sqliteRepository) DeleteClassificationRule(ctx context.Context, ruleID string) error {
	query := "DELETE FROM classification_rules WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("classification rule with ID %s not found", ruleID)
	}

	return nil
}

// ListClassificationRules lists all classification rules
func (r *sqliteRepository) ListClassificationRules(ctx context.Context) ([]classification.ClassificationRule, error) {
	query := `
		SELECT id, name, description, conditions, logic_operator, layer, device_type, priority, is_active, confidence, created_by, created_at, updated_at
		FROM classification_rules
		ORDER BY priority DESC, name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []classification.ClassificationRule
	for rows.Next() {
		var rule classification.ClassificationRule
		var conditionsJSON string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &conditionsJSON, &rule.LogicOperator,
			&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.Confidence,
			&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal conditions
		if err := json.Unmarshal([]byte(conditionsJSON), &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// ListActiveClassificationRules lists all active classification rules
func (r *sqliteRepository) ListActiveClassificationRules(ctx context.Context) ([]classification.ClassificationRule, error) {
	query := `
		SELECT id, name, description, conditions, logic_operator, layer, device_type, priority, is_active, confidence, created_by, created_at, updated_at
		FROM classification_rules
		WHERE is_active = true
		ORDER BY priority DESC, name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []classification.ClassificationRule
	for rows.Next() {
		var rule classification.ClassificationRule
		var conditionsJSON string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &conditionsJSON, &rule.LogicOperator,
			&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.Confidence,
			&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal conditions
		if err := json.Unmarshal([]byte(conditionsJSON), &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// Hierarchy Layers methods

// GetHierarchyLayer retrieves a specific hierarchy layer
func (r *sqliteRepository) GetHierarchyLayer(ctx context.Context, layerID int) (*classification.HierarchyLayer, error) {
	var layer classification.HierarchyLayer

	query := `
		SELECT id, name, description, created_at, updated_at
		FROM hierarchy_layers
		WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, layerID).Scan(
		&layer.ID, &layer.Name, &layer.Description, &layer.CreatedAt, &layer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &layer, nil
}

// ListHierarchyLayers retrieves all hierarchy layers
func (r *sqliteRepository) ListHierarchyLayers(ctx context.Context) ([]classification.HierarchyLayer, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM hierarchy_layers
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []classification.HierarchyLayer
	for rows.Next() {
		var layer classification.HierarchyLayer

		err := rows.Scan(
			&layer.ID, &layer.Name, &layer.Description, &layer.CreatedAt, &layer.UpdatedAt)
		if err != nil {
			return nil, err
		}

		layers = append(layers, layer)
	}

	return layers, rows.Err()
}

// SaveHierarchyLayer creates or updates a hierarchy layer
func (r *sqliteRepository) SaveHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	query := `
		INSERT INTO hierarchy_layers (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.ExecContext(ctx, query,
		layer.ID, layer.Name, layer.Description,
		layer.CreatedAt, layer.UpdatedAt)

	return err
}

// UpdateHierarchyLayer updates an existing hierarchy layer
func (r *sqliteRepository) UpdateHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	query := `
		UPDATE hierarchy_layers SET
			name = ?,
			description = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		layer.Name, layer.Description, layer.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("hierarchy layer with ID %d not found", layer.ID)
	}

	return nil
}

// DeleteHierarchyLayer deletes a hierarchy layer
func (r *sqliteRepository) DeleteHierarchyLayer(ctx context.Context, layerID int) error {
	query := "DELETE FROM hierarchy_layers WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, layerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("hierarchy layer with ID %d not found", layerID)
	}

	return nil
}

// Placeholder methods for classification suggestions (can be implemented as needed)

func (r *sqliteRepository) SaveClassificationSuggestion(ctx context.Context, suggestion classification.ClassificationSuggestion) error {
	// Implementation for saving classification suggestions
	return fmt.Errorf("SaveClassificationSuggestion not implemented")
}

func (r *sqliteRepository) GetClassificationSuggestion(ctx context.Context, suggestionID string) (*classification.ClassificationSuggestion, error) {
	// Implementation for getting classification suggestions
	return nil, fmt.Errorf("GetClassificationSuggestion not implemented")
}

func (r *sqliteRepository) ListPendingClassificationSuggestions(ctx context.Context) ([]classification.ClassificationSuggestion, error) {
	// Implementation for listing pending suggestions
	return []classification.ClassificationSuggestion{}, nil
}

func (r *sqliteRepository) UpdateClassificationSuggestionStatus(ctx context.Context, suggestionID string, status classification.SuggestionStatus) error {
	// Implementation for updating suggestion status
	return fmt.Errorf("UpdateClassificationSuggestionStatus not implemented")
}

func (r *sqliteRepository) DeleteClassificationSuggestion(ctx context.Context, suggestionID string) error {
	// Implementation for deleting classification suggestions
	return fmt.Errorf("DeleteClassificationSuggestion not implemented")
}
