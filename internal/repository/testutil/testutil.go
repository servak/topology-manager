package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/stretchr/testify/require"
)

// CreateTestDevice creates a test device with default values
func CreateTestDevice(id string) topology.Device {
	now := time.Now()
	layerID := 1
	return topology.Device{
		ID:           id,
		Type:         "switch",
		Hardware:     "Test Hardware",
		LayerID:      &layerID,
		DeviceType:   "network-switch",
		ClassifiedBy: "test",
		Metadata:     map[string]string{"test": "value"},
		LastSeen:     now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// CreateTestLink creates a test link with default values
func CreateTestLink(id, sourceID, targetID string) topology.Link {
	now := time.Now()
	return topology.Link{
		ID:         id,
		SourceID:   sourceID,
		TargetID:   targetID,
		SourcePort: "eth0",
		TargetPort: "eth1",
		Weight:     1.0,
		Metadata:   map[string]string{"test": "link"},
		LastSeen:   now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CreateTestHierarchyLayer creates a test hierarchy layer
func CreateTestHierarchyLayer(id int, name string) classification.HierarchyLayer {
	now := time.Now()
	return classification.HierarchyLayer{
		ID:          id,
		Name:        name,
		Description: "Test layer: " + name,
		Order:       id,
		Color:       "#FF0000",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateTestClassificationRule creates a test classification rule
func CreateTestClassificationRule(id, name string) classification.ClassificationRule {
	now := time.Now()
	return classification.ClassificationRule{
		ID:            id,
		Name:          name,
		Description:   "Test rule: " + name,
		LogicOperator: "AND",
		Conditions: []classification.Condition{
			{
				Field:    "hardware",
				Operator: "contains",
				Value:    "test",
			},
		},
		Layer:      1,
		DeviceType: "test-device",
		Priority:   100,
		IsActive:   true,
		CreatedBy:  "test-user",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CreateTestDeviceClassification creates a test device classification
func CreateTestDeviceClassification(deviceID string) classification.DeviceClassification {
	now := time.Now()
	return classification.DeviceClassification{
		ID:        "test-classification-" + deviceID,
		DeviceID:  deviceID,
		Layer:     1,
		DeviceType: "test-device",
		IsManual:  true,
		CreatedBy: "test-user",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AssertDeviceEqual compares two devices for testing
func AssertDeviceEqual(t *testing.T, expected, actual topology.Device) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Type, actual.Type)
	require.Equal(t, expected.Hardware, actual.Hardware)
	require.Equal(t, expected.LayerID, actual.LayerID)
	require.Equal(t, expected.DeviceType, actual.DeviceType)
	require.Equal(t, expected.ClassifiedBy, actual.ClassifiedBy)
	require.Equal(t, expected.Metadata, actual.Metadata)
	// Don't check exact timestamps, just ensure they're set
	require.False(t, actual.CreatedAt.IsZero())
	require.False(t, actual.UpdatedAt.IsZero())
}

// AssertLinkEqual compares two links for testing
func AssertLinkEqual(t *testing.T, expected, actual topology.Link) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.SourceID, actual.SourceID)
	require.Equal(t, expected.TargetID, actual.TargetID)
	require.Equal(t, expected.SourcePort, actual.SourcePort)
	require.Equal(t, expected.TargetPort, actual.TargetPort)
	require.Equal(t, expected.Weight, actual.Weight)
	require.Equal(t, expected.Metadata, actual.Metadata)
	require.False(t, actual.CreatedAt.IsZero())
	require.False(t, actual.UpdatedAt.IsZero())
}

// CleanupRepository clears all data from repository for testing
func CleanupRepository(t *testing.T, repo interface {
	Clear() error
}) {
	err := repo.Clear()
	require.NoError(t, err)
}

// TestPaginationOptions creates test pagination options
func TestPaginationOptions(page, pageSize int) topology.PaginationOptions {
	return topology.PaginationOptions{
		Page:     page,
		PageSize: pageSize,
	}
}