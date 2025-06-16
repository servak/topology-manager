package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/repository/sqlite"
	"github.com/servak/topology-manager/pkg/logger"
	"github.com/stretchr/testify/require"
)

// TestSetup contains all necessary components for testing
type TestSetup struct {
	Repo   repository.Repository
	Logger *logger.Logger
}

// NewTestSetup creates a new test setup with in-memory SQLite
func NewTestSetup(t *testing.T) *TestSetup {
	// Create in-memory SQLite repository
	config := sqlite.Config{
		Path: ":memory:",
	}
	
	repo, err := sqlite.NewSQliteRepository(config)
	require.NoError(t, err)
	
	// Run migrations
	err = repo.Migrate()
	require.NoError(t, err)
	
	// Create logger
	appLogger := logger.New("debug")
	
	return &TestSetup{
		Repo:   repo,
		Logger: appLogger,
	}
}

// Cleanup cleans up test resources
func (ts *TestSetup) Cleanup() {
	if ts.Repo != nil {
		ts.Repo.Close()
	}
}

// CreateTestDevice creates a test device with default values
func CreateTestDevice(id string) topology.Device {
	now := time.Now()
	layerID := 1
	return topology.Device{
		ID:           id,
		Type:         "switch",
		Hardware:     "Test Hardware " + id,
		LayerID:      &layerID,
		DeviceType:   "network-switch",
		ClassifiedBy: "test",
		Metadata:     map[string]string{"test": "value", "id": id},
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
		Metadata:   map[string]string{"test": "link", "id": id},
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
				Value:    "Test",
			},
		},
		Layer:      1,
		DeviceType: "network-switch",
		Priority:   100,
		IsActive:   true,
		CreatedBy:  "test-user",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// SeedTestData seeds the test database with sample data
func (ts *TestSetup) SeedTestData(t *testing.T) {
	ctx := context.Background()
	
	// Create test devices
	devices := []topology.Device{
		CreateTestDevice("device-001"),
		CreateTestDevice("device-002"),
		CreateTestDevice("device-003"),
	}
	
	for _, device := range devices {
		err := ts.Repo.AddDevice(ctx, device)
		require.NoError(t, err)
	}
	
	// Create test links
	links := []topology.Link{
		CreateTestLink("link-001", "device-001", "device-002"),
		CreateTestLink("link-002", "device-002", "device-003"),
	}
	
	for _, link := range links {
		err := ts.Repo.AddLink(ctx, link)
		require.NoError(t, err)
	}
	
	// Create test classification rule
	rule := CreateTestClassificationRule("rule-001", "Test Switch Rule")
	err := ts.Repo.SaveClassificationRule(ctx, rule)
	require.NoError(t, err)
}