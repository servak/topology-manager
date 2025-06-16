package service

import (
	"context"
	"testing"

	"github.com/servak/topology-manager/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassificationService_ListHierarchyLayers(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	ctx := context.Background()

	// Create classification service
	classificationService := NewClassificationService(setup.Repo, setup.Logger)

	// Test listing hierarchy layers (default ones should be present from migration)
	layers, err := classificationService.ListHierarchyLayers(ctx)
	require.NoError(t, err)
	
	// Should have default layers from migration
	assert.GreaterOrEqual(t, len(layers), 1)
	
	// Find the "Core" layer that should exist
	var coreLayer *string
	for _, layer := range layers {
		if layer.Name == "Core" {
			name := layer.Name
			coreLayer = &name
			break
		}
	}
	assert.NotNil(t, coreLayer, "Should have a Core layer from default data")
}

func TestClassificationService_SaveAndGetHierarchyLayer(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	ctx := context.Background()

	// Create a test layer
	testLayer := testutil.CreateTestHierarchyLayer(100, "Test Layer")
	
	// Save the layer
	err := setup.ClassificationService.SaveHierarchyLayer(ctx, testLayer)
	require.NoError(t, err)

	// Retrieve the layer
	retrievedLayer, err := setup.ClassificationService.GetHierarchyLayer(ctx, testLayer.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedLayer)

	// Verify the retrieved layer
	assert.Equal(t, testLayer.ID, retrievedLayer.ID)
	assert.Equal(t, testLayer.Name, retrievedLayer.Name)
	assert.Equal(t, testLayer.Description, retrievedLayer.Description)
	assert.Equal(t, testLayer.Order, retrievedLayer.Order)
}

func TestClassificationService_ListClassificationRules(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	ctx := context.Background()

	// Create test rules
	rule1 := testutil.CreateTestClassificationRule("rule-001", "Switch Rule")
	rule2 := testutil.CreateTestClassificationRule("rule-002", "Router Rule")

	// Save rules
	err := setup.ClassificationService.SaveClassificationRule(ctx, rule1)
	require.NoError(t, err)
	
	err = setup.ClassificationService.SaveClassificationRule(ctx, rule2)
	require.NoError(t, err)

	// List all rules
	rules, err := setup.ClassificationService.ListClassificationRules(ctx)
	require.NoError(t, err)
	
	// Should have at least our test rules (plus any default ones)
	assert.GreaterOrEqual(t, len(rules), 2)
	
	// Find our test rules
	foundRule1 := false
	foundRule2 := false
	for _, rule := range rules {
		if rule.ID == "rule-001" {
			foundRule1 = true
			assert.Equal(t, "Switch Rule", rule.Name)
		}
		if rule.ID == "rule-002" {
			foundRule2 = true
			assert.Equal(t, "Router Rule", rule.Name)
		}
	}
	assert.True(t, foundRule1, "Should find rule-001")
	assert.True(t, foundRule2, "Should find rule-002")
}

func TestClassificationService_ListActiveClassificationRules(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	ctx := context.Background()

	// Create active and inactive rules
	activeRule := testutil.CreateTestClassificationRule("active-rule", "Active Rule")
	activeRule.IsActive = true
	
	inactiveRule := testutil.CreateTestClassificationRule("inactive-rule", "Inactive Rule")
	inactiveRule.IsActive = false

	// Save rules
	err := setup.ClassificationService.SaveClassificationRule(ctx, activeRule)
	require.NoError(t, err)
	
	err = setup.ClassificationService.SaveClassificationRule(ctx, inactiveRule)
	require.NoError(t, err)

	// List only active rules
	activeRules, err := setup.ClassificationService.ListActiveClassificationRules(ctx)
	require.NoError(t, err)

	// Find our active rule (should be present)
	foundActive := false
	foundInactive := false
	for _, rule := range activeRules {
		if rule.ID == "active-rule" {
			foundActive = true
			assert.True(t, rule.IsActive)
		}
		if rule.ID == "inactive-rule" {
			foundInactive = true
		}
	}
	assert.True(t, foundActive, "Should find active rule")
	assert.False(t, foundInactive, "Should not find inactive rule in active list")
}

func TestClassificationService_ListUnclassifiedDevices(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()
	setup.SeedTestData(t)

	ctx := context.Background()

	// All devices should be unclassified initially
	devices, err := setup.ClassificationService.ListUnclassifiedDevices(ctx)
	require.NoError(t, err)
	
	// Should have our test devices
	assert.GreaterOrEqual(t, len(devices), 3)
	
	// Check that we have our test device IDs
	deviceIDs := make(map[string]bool)
	for _, device := range devices {
		deviceIDs[device.ID] = true
	}
	assert.True(t, deviceIDs["device-001"], "Should find device-001")
	assert.True(t, deviceIDs["device-002"], "Should find device-002")
	assert.True(t, deviceIDs["device-003"], "Should find device-003")
}

func TestClassificationService_ListUnclassifiedDevicesWithPagination(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()
	setup.SeedTestData(t)

	ctx := context.Background()

	// Test pagination with limit 2, offset 0
	devices, err := setup.ClassificationService.ListUnclassifiedDevicesWithPagination(ctx, 2, 0)
	require.NoError(t, err)
	
	// Should get exactly 2 devices
	assert.Equal(t, 2, len(devices))

	// Test pagination with limit 2, offset 1
	devicesOffset, err := setup.ClassificationService.ListUnclassifiedDevicesWithPagination(ctx, 2, 1)
	require.NoError(t, err)
	
	// Should get at least 1 device (remaining ones)
	assert.GreaterOrEqual(t, len(devicesOffset), 1)
	
	// Devices should be different from first page
	if len(devices) > 0 && len(devicesOffset) > 0 {
		assert.NotEqual(t, devices[0].ID, devicesOffset[0].ID)
	}
}

func TestClassificationService_CountUnclassifiedDevices(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()
	setup.SeedTestData(t)

	ctx := context.Background()

	// Count unclassified devices
	count, err := setup.ClassificationService.CountUnclassifiedDevices(ctx)
	require.NoError(t, err)
	
	// Should have at least our 3 test devices
	assert.GreaterOrEqual(t, count, 3)
}