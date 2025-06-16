package classification

import "time"

// DeviceClassification represents the manual or automatic classification of a device
type DeviceClassification struct {
	ID         string    `json:"id" db:"id"`
	DeviceID   string    `json:"device_id" db:"device_id"`
	Layer      int       `json:"layer" db:"layer"`
	DeviceType string    `json:"device_type" db:"device_type"`
	IsManual   bool      `json:"is_manual" db:"is_manual"`
	CreatedBy  string    `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// RuleCondition represents a single condition in a classification rule
type RuleCondition struct {
	Field    string `json:"field"`    // "name", "hardware", "ip_address", "type"
	Operator string `json:"operator"` // "contains", "starts_with", "ends_with", "equals", "regex"
	Value    string `json:"value"`
}

// ClassificationRule represents a rule for automatic device classification
type ClassificationRule struct {
	ID            string          `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description" db:"description"`
	LogicOperator string          `json:"logic" db:"logic_operator"`
	Conditions    []RuleCondition `json:"conditions" db:"conditions"`
	Layer         int             `json:"layer" db:"layer"`
	DeviceType    string          `json:"device_type" db:"device_type"`
	Priority      int             `json:"priority" db:"priority"` // Higher priority rules are applied first
	IsActive      bool            `json:"is_active" db:"is_active"`
	Confidence    float64         `json:"confidence" db:"confidence"` // 0.0 - 1.0
	CreatedBy     string          `json:"created_by" db:"created_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// ClassificationSuggestion represents a suggested rule based on manual classifications
type ClassificationSuggestion struct {
	ID              string             `json:"id"`
	RuleID          string             `json:"rule_id"`
	Rule            ClassificationRule `json:"rule"`
	AffectedDevices []string           `json:"affected_devices"`
	BasedOnDevices  []string           `json:"based_on_devices"`
	Confidence      float64            `json:"confidence"`
	Status          SuggestionStatus   `json:"status"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// SuggestionStatus represents the status of a classification suggestion
type SuggestionStatus string

const (
	SuggestionStatusPending  SuggestionStatus = "pending"
	SuggestionStatusAccepted SuggestionStatus = "accepted"
	SuggestionStatusRejected SuggestionStatus = "rejected"
	SuggestionStatusModified SuggestionStatus = "modified"
)

// HierarchyLayer represents a network layer definition
type HierarchyLayer struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Order       int       `json:"order" db:"order_index"` // Display order (0 = top)
	Color       string    `json:"color" db:"color"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DefaultHierarchyLayers returns the default network hierarchy layers
func DefaultHierarchyLayers() []HierarchyLayer {
	return []HierarchyLayer{
		{ID: 0, Name: "Internet Gateway", Description: "External internet connection point", Order: 0, Color: "#e74c3c"},
		{ID: 1, Name: "Firewall", Description: "Security appliances", Order: 1, Color: "#e67e22"},
		{ID: 2, Name: "Core Router", Description: "Core network routing", Order: 2, Color: "#f39c12"},
		{ID: 3, Name: "Distribution", Description: "Distribution layer switches", Order: 3, Color: "#3498db"},
		{ID: 4, Name: "Access", Description: "Access layer switches", Order: 4, Color: "#2ecc71"},
		{ID: 5, Name: "Server", Description: "End devices and servers", Order: 5, Color: "#95a5a6"},
	}
}
