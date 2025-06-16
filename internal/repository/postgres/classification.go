package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/servak/topology-manager/internal/domain/classification"
)

// Classification-related repository methods

// Device Classifications
func (r *postgresRepository) GetDeviceClassification(ctx context.Context, deviceID string) (*classification.DeviceClassification, error) {
	query := `
		SELECT id, device_id, layer, device_type, is_manual, created_by, created_at, updated_at
		FROM device_classifications 
		WHERE device_id = $1
	`

	var c classification.DeviceClassification
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&c.ID, &c.DeviceID, &c.Layer, &c.DeviceType,
		&c.IsManual, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device classification: %w", err)
	}

	return &c, nil
}

func (r *postgresRepository) ListDeviceClassifications(ctx context.Context) ([]classification.DeviceClassification, error) {
	query := `
		SELECT id, device_id, layer, device_type, is_manual, created_by, created_at, updated_at
		FROM device_classifications 
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list device classifications: %w", err)
	}
	defer rows.Close()

	var classifications []classification.DeviceClassification
	for rows.Next() {
		var c classification.DeviceClassification
		err := rows.Scan(
			&c.ID, &c.DeviceID, &c.Layer, &c.DeviceType,
			&c.IsManual, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device classification: %w", err)
		}
		classifications = append(classifications, c)
	}

	return classifications, nil
}

func (r *postgresRepository) ListUnclassifiedDevices(ctx context.Context) ([]string, error) {
	query := `
		SELECT d.id 
		FROM devices d 
		LEFT JOIN device_classifications dc ON d.id = dc.device_id 
		WHERE dc.device_id IS NULL
		ORDER BY d.id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list unclassified devices: %w", err)
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("failed to scan device ID: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs, nil
}

func (r *postgresRepository) ListUnclassifiedDevicesWithPagination(ctx context.Context, limit, offset int) ([]string, error) {
	query := `
		SELECT d.id 
		FROM devices d 
		LEFT JOIN device_classifications dc ON d.id = dc.device_id 
		WHERE dc.device_id IS NULL
		ORDER BY d.id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list unclassified devices with pagination: %w", err)
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("failed to scan device ID: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs, nil
}

func (r *postgresRepository) CountUnclassifiedDevices(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(d.id)
		FROM devices d 
		LEFT JOIN device_classifications dc ON d.id = dc.device_id 
		WHERE dc.device_id IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unclassified devices: %w", err)
	}

	return count, nil
}

func (r *postgresRepository) SaveDeviceClassification(ctx context.Context, c classification.DeviceClassification) error {
	// UUIDが設定されていない場合は生成
	if c.ID == "" {
		c.ID = uuid.New().String()
	}

	// 作成日時が設定されていない場合は現在時刻を設定
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	c.UpdatedAt = time.Now()

	query := `
		INSERT INTO device_classifications (id, device_id, layer, device_type, is_manual, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (device_id) DO UPDATE SET
			layer = EXCLUDED.layer,
			device_type = EXCLUDED.device_type,
			is_manual = EXCLUDED.is_manual,
			created_by = EXCLUDED.created_by,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.DeviceID, c.Layer, c.DeviceType,
		c.IsManual, c.CreatedBy, c.CreatedAt, c.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save device classification: %w", err)
	}

	return nil
}

func (r *postgresRepository) DeleteDeviceClassification(ctx context.Context, deviceID string) error {
	query := `DELETE FROM device_classifications WHERE device_id = $1`

	_, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete device classification: %w", err)
	}

	return nil
}

// Classification Rules
func (r *postgresRepository) GetClassificationRule(ctx context.Context, ruleID string) (*classification.ClassificationRule, error) {
	query := `
		SELECT id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at
		FROM classification_rules 
		WHERE id = $1
	`

	var rule classification.ClassificationRule
	var conditionsJSON []byte
	err := r.db.QueryRowContext(ctx, query, ruleID).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.LogicOperator, &conditionsJSON,
		&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get classification rule: %w", err)
	}

	// JSONBからConditionsをデシリアライズ
	if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	return &rule, nil
}

func (r *postgresRepository) ListClassificationRules(ctx context.Context) ([]classification.ClassificationRule, error) {
	query := `
		SELECT id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at
		FROM classification_rules 
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list classification rules: %w", err)
	}
	defer rows.Close()

	var rules []classification.ClassificationRule
	for rows.Next() {
		var rule classification.ClassificationRule
		var conditionsJSON []byte
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.LogicOperator, &conditionsJSON,
			&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification rule: %w", err)
		}

		// JSONBからConditionsをデシリアライズ
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *postgresRepository) ListActiveClassificationRules(ctx context.Context) ([]classification.ClassificationRule, error) {
	query := `
		SELECT id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at
		FROM classification_rules 
		WHERE is_active = true
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active classification rules: %w", err)
	}
	defer rows.Close()

	var rules []classification.ClassificationRule
	for rows.Next() {
		var rule classification.ClassificationRule
		var conditionsJSON []byte
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.LogicOperator, &conditionsJSON,
			&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification rule: %w", err)
		}

		// JSONBからConditionsをデシリアライズ
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *postgresRepository) SaveClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	// UUIDが設定されていない場合は生成
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	// 作成日時が設定されていない場合は現在時刻を設定
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	rule.UpdatedAt = time.Now()

	// JSONBにConditionsを変換
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		INSERT INTO classification_rules (id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.LogicOperator, conditionsJSON,
		rule.Layer, rule.DeviceType, rule.Priority, rule.IsActive, rule.CreatedBy, rule.CreatedAt, rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save classification rule: %w", err)
	}

	return nil
}

func (r *postgresRepository) UpdateClassificationRule(ctx context.Context, rule classification.ClassificationRule) error {
	rule.UpdatedAt = time.Now()

	// JSONBにConditionsを変換
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		UPDATE classification_rules 
		SET name = $2, description = $3, logic_operator = $4, conditions = $5, 
		    layer = $6, device_type = $7, priority = $8, is_active = $9, updated_at = $10
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.LogicOperator, conditionsJSON,
		rule.Layer, rule.DeviceType, rule.Priority, rule.IsActive, rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update classification rule: %w", err)
	}

	return nil
}

func (r *postgresRepository) DeleteClassificationRule(ctx context.Context, ruleID string) error {
	query := `DELETE FROM classification_rules WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete classification rule: %w", err)
	}

	return nil
}

// Classification Suggestions
func (r *postgresRepository) GetClassificationSuggestion(ctx context.Context, suggestionID string) (*classification.ClassificationSuggestion, error) {
	query := `
		SELECT s.id, s.rule_id, s.confidence, s.status, s.affected_devices, s.based_on_devices, s.created_at, s.updated_at,
		       r.id, r.name, r.description, r.logic_operator, r.conditions, r.layer, r.device_type, r.priority, r.is_active, r.created_by, r.created_at, r.updated_at
		FROM classification_suggestions s
		JOIN classification_rules r ON s.rule_id = r.id
		WHERE s.id = $1
	`

	var suggestion classification.ClassificationSuggestion
	var rule classification.ClassificationRule
	var affectedDevicesJSON, basedOnDevicesJSON []byte
	var conditionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, suggestionID).Scan(
		&suggestion.ID, &suggestion.RuleID, &suggestion.Confidence, &suggestion.Status,
		&affectedDevicesJSON, &basedOnDevicesJSON, &suggestion.CreatedAt, &suggestion.UpdatedAt,
		&rule.ID, &rule.Name, &rule.Description, &rule.LogicOperator, &conditionsJSON,
		&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get classification suggestion: %w", err)
	}

	// JSON配列をスライスに変換
	if err := json.Unmarshal(affectedDevicesJSON, &suggestion.AffectedDevices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal affected devices: %w", err)
	}
	if err := json.Unmarshal(basedOnDevicesJSON, &suggestion.BasedOnDevices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal based on devices: %w", err)
	}
	if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rule conditions: %w", err)
	}

	suggestion.Rule = rule

	return &suggestion, nil
}

func (r *postgresRepository) ListPendingClassificationSuggestions(ctx context.Context) ([]classification.ClassificationSuggestion, error) {
	query := `
		SELECT s.id, s.rule_id, s.confidence, s.status, s.affected_devices, s.based_on_devices, s.created_at, s.updated_at,
		       r.id, r.name, r.description, r.logic_operator, r.conditions, r.layer, r.device_type, r.priority, r.is_active, r.created_by, r.created_at, r.updated_at
		FROM classification_suggestions s
		JOIN classification_rules r ON s.rule_id = r.id
		WHERE s.status = 'pending'
		ORDER BY s.confidence DESC, s.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending classification suggestions: %w", err)
	}
	defer rows.Close()

	var suggestions []classification.ClassificationSuggestion
	for rows.Next() {
		var suggestion classification.ClassificationSuggestion
		var rule classification.ClassificationRule
		var affectedDevices, basedOnDevices pq.StringArray
		var conditionsJSON []byte

		err := rows.Scan(
			&suggestion.ID, &suggestion.RuleID, &suggestion.Confidence, &suggestion.Status,
			&affectedDevices, &basedOnDevices, &suggestion.CreatedAt, &suggestion.UpdatedAt,
			&rule.ID, &rule.Name, &rule.Description, &rule.LogicOperator, &conditionsJSON,
			&rule.Layer, &rule.DeviceType, &rule.Priority, &rule.IsActive, &rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classification suggestion: %w", err)
		}

		// JSONからConditionsをデシリアライズ
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule conditions: %w", err)
		}

		suggestion.AffectedDevices = []string(affectedDevices)
		suggestion.BasedOnDevices = []string(basedOnDevices)
		suggestion.Rule = rule

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

func (r *postgresRepository) SaveClassificationSuggestion(ctx context.Context, suggestion classification.ClassificationSuggestion) error {
	// UUIDが設定されていない場合は生成
	if suggestion.ID == "" {
		suggestion.ID = uuid.New().String()
	}

	// 作成日時が設定されていない場合は現在時刻を設定
	if suggestion.CreatedAt.IsZero() {
		suggestion.CreatedAt = time.Now()
	}
	suggestion.UpdatedAt = time.Now()

	query := `
		INSERT INTO classification_suggestions (id, rule_id, confidence, status, affected_devices, based_on_devices, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		suggestion.ID, suggestion.RuleID, suggestion.Confidence, string(suggestion.Status),
		pq.Array(suggestion.AffectedDevices), pq.Array(suggestion.BasedOnDevices),
		suggestion.CreatedAt, suggestion.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save classification suggestion: %w", err)
	}

	return nil
}

func (r *postgresRepository) UpdateClassificationSuggestionStatus(ctx context.Context, suggestionID string, status classification.SuggestionStatus) error {
	query := `UPDATE classification_suggestions SET status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, suggestionID, string(status), time.Now())
	if err != nil {
		return fmt.Errorf("failed to update suggestion status: %w", err)
	}

	return nil
}

func (r *postgresRepository) DeleteClassificationSuggestion(ctx context.Context, suggestionID string) error {
	query := `DELETE FROM classification_suggestions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, suggestionID)
	if err != nil {
		return fmt.Errorf("failed to delete classification suggestion: %w", err)
	}

	return nil
}

// Hierarchy Layers
func (r *postgresRepository) GetHierarchyLayer(ctx context.Context, layerID int) (*classification.HierarchyLayer, error) {
	query := `
		SELECT id, name, description, order_index, color, created_at, updated_at
		FROM hierarchy_layers 
		WHERE id = $1
	`

	var layer classification.HierarchyLayer
	err := r.db.QueryRowContext(ctx, query, layerID).Scan(
		&layer.ID, &layer.Name, &layer.Description, &layer.Order, &layer.Color, &layer.CreatedAt, &layer.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy layer: %w", err)
	}

	return &layer, nil
}

func (r *postgresRepository) ListHierarchyLayers(ctx context.Context) ([]classification.HierarchyLayer, error) {
	query := `
		SELECT id, name, description, order_index, color, created_at, updated_at
		FROM hierarchy_layers 
		ORDER BY order_index
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list hierarchy layers: %w", err)
	}
	defer rows.Close()

	var layers []classification.HierarchyLayer
	for rows.Next() {
		var layer classification.HierarchyLayer
		err := rows.Scan(
			&layer.ID, &layer.Name, &layer.Description, &layer.Order, &layer.Color, &layer.CreatedAt, &layer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hierarchy layer: %w", err)
		}
		layers = append(layers, layer)
	}

	return layers, nil
}

func (r *postgresRepository) SaveHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	// 作成日時が設定されていない場合は現在時刻を設定
	if layer.CreatedAt.IsZero() {
		layer.CreatedAt = time.Now()
	}
	layer.UpdatedAt = time.Now()

	query := `
		INSERT INTO hierarchy_layers (id, name, description, order_index, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			order_index = EXCLUDED.order_index,
			color = EXCLUDED.color,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		layer.ID, layer.Name, layer.Description, layer.Order, layer.Color, layer.CreatedAt, layer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save hierarchy layer: %w", err)
	}

	return nil
}

func (r *postgresRepository) UpdateHierarchyLayer(ctx context.Context, layer classification.HierarchyLayer) error {
	layer.UpdatedAt = time.Now()

	query := `
		UPDATE hierarchy_layers 
		SET name = $2, description = $3, order_index = $4, color = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		layer.ID, layer.Name, layer.Description, layer.Order, layer.Color, layer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update hierarchy layer: %w", err)
	}

	return nil
}

func (r *postgresRepository) DeleteHierarchyLayer(ctx context.Context, layerID int) error {
	query := `DELETE FROM hierarchy_layers WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, layerID)
	if err != nil {
		return fmt.Errorf("failed to delete hierarchy layer: %w", err)
	}

	return nil
}

