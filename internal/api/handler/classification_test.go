package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/servak/topology-manager/internal/service"
	"github.com/servak/topology-manager/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClassificationHandler(t *testing.T) (*ClassificationHandler, *testutil.TestSetup, huma.API) {
	setup := testutil.NewTestSetup(t)
	setup.SeedTestData(t)

	// Create services
	classificationService := service.NewClassificationService(setup.Repo, setup.Logger)

	// Create handler
	handler := NewClassificationHandler(classificationService, setup.Logger)

	// Create test API
	router := chi.NewRouter()
	config := huma.DefaultConfig("Test API", "1.0.0")
	api := humachi.New(router, config)

	// Register routes
	handler.RegisterRoutes(api)

	return handler, setup, api
}

func TestClassificationHandler_ListHierarchyLayers(t *testing.T) {
	handler, setup, api := setupClassificationHandler(t)
	defer setup.Cleanup()

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/classification/layers", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse response body
	var response struct {
		Layers []map[string]interface{} `json:"layers"`
		Count  int                      `json:"count"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have default layers from migration
	assert.GreaterOrEqual(t, response.Count, 1)
	assert.GreaterOrEqual(t, len(response.Layers), 1)

	// Check that we have the expected Core layer
	foundCore := false
	for _, layer := range response.Layers {
		if name, ok := layer["name"].(string); ok && name == "Core" {
			foundCore = true
			assert.Equal(t, "Core network layer - backbone switches and routers", layer["description"])
			break
		}
	}
	assert.True(t, foundCore, "Should find Core layer in response")
}

func TestClassificationHandler_ListClassificationRules(t *testing.T) {
	handler, setup, api := setupClassificationHandler(t)
	defer setup.Cleanup()

	// Add a test rule first
	ctx := context.Background()
	testRule := testutil.CreateTestClassificationRule("test-rule-001", "Test API Rule")
	// Create classification service to add test rule
	classificationService := service.NewClassificationService(setup.Repo, setup.Logger)
	err := classificationService.SaveClassificationRule(ctx, testRule)
	require.NoError(t, err)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/classification/rules", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse response body
	var response struct {
		Rules []map[string]interface{} `json:"rules"`
		Count int                      `json:"count"`
	}
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have at least our test rule
	assert.GreaterOrEqual(t, response.Count, 1)
	assert.GreaterOrEqual(t, len(response.Rules), 1)

	// Find our test rule
	foundTestRule := false
	for _, rule := range response.Rules {
		if id, ok := rule["id"].(string); ok && id == "test-rule-001" {
			foundTestRule = true
			assert.Equal(t, "Test API Rule", rule["name"])
			assert.Equal(t, true, rule["is_active"])
			break
		}
	}
	assert.True(t, foundTestRule, "Should find test rule in response")
}

func TestClassificationHandler_ListUnclassifiedDevices(t *testing.T) {
	handler, setup, api := setupClassificationHandler(t)
	defer setup.Cleanup()

	// Create test request with query parameters
	req := httptest.NewRequest("GET", "/api/v1/classification/devices/unclassified?limit=10&offset=0", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse response body
	var response struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
		Total   int                      `json:"total"`
		Limit   int                      `json:"limit"`
		Offset  int                      `json:"offset"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have our test devices (they're unclassified by default)
	assert.GreaterOrEqual(t, response.Count, 3)
	assert.GreaterOrEqual(t, len(response.Devices), 3)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Offset)

	// Check that devices have expected properties
	deviceIDs := make(map[string]bool)
	for _, device := range response.Devices {
		if id, ok := device["id"].(string); ok {
			deviceIDs[id] = true
			assert.NotEmpty(t, device["type"], "Device should have type")
			assert.NotEmpty(t, device["hardware"], "Device should have hardware")
		}
	}

	// Should find our seeded test devices
	assert.True(t, deviceIDs["device-001"] || deviceIDs["device-002"] || deviceIDs["device-003"], 
		"Should find at least one of our test devices")
}

func TestClassificationHandler_ListClassifiedDevices(t *testing.T) {
	handler, setup, api := setupClassificationHandler(t)
	defer setup.Cleanup()

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/classification/devices/classified", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse response body
	var response struct {
		Classifications []map[string]interface{} `json:"classifications"`
		Count           int                      `json:"count"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Initially should have 0 classified devices (all test devices are unclassified)
	assert.Equal(t, 0, response.Count)
	assert.Equal(t, 0, len(response.Classifications))
}

func TestClassificationHandler_CreateClassificationRule(t *testing.T) {
	handler, setup, api := setupClassificationHandler(t)
	defer setup.Cleanup()

	// Create test rule data
	ruleData := map[string]interface{}{
		"name":        "API Test Rule",
		"description": "Rule created via API test",
		"conditions": []map[string]interface{}{
			{
				"field":    "hardware",
				"operator": "contains",
				"value":    "Cisco",
			},
		},
		"logic_operator": "AND",
		"layer":          2,
		"device_type":    "cisco-switch",
		"priority":       150,
		"is_active":      true,
		"created_by":     "api-test",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(ruleData)
	require.NoError(t, err)

	// Create test request
	req := httptest.NewRequest("POST", "/api/v1/classification/rules", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have created rule with generated ID
	assert.NotEmpty(t, response["id"])
	assert.Equal(t, "API Test Rule", response["name"])
	assert.Equal(t, true, response["is_active"])
}

func TestClassificationHandler_CreateHierarchyLayer(t *testing.T) {
	handler, setup, api := setupClassificationHandler(t)
	defer setup.Cleanup()

	// Create test layer data
	layerData := map[string]interface{}{
		"name":        "Test API Layer",
		"description": "Layer created via API test",
		"order":       10,
		"color":       "#00FF00",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(layerData)
	require.NoError(t, err)

	// Create test request
	req := httptest.NewRequest("POST", "/api/v1/classification/layers", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have created layer
	assert.NotEmpty(t, response["id"])
	assert.Equal(t, "Test API Layer", response["name"])
	assert.Equal(t, "Layer created via API test", response["description"])
	assert.Equal(t, float64(10), response["order"]) // JSON numbers are float64
	assert.Equal(t, "#00FF00", response["color"])
}

func TestClassificationHandler_ErrorHandling(t *testing.T) {
	handler, setup, api := setupClassificationHandler(t)
	defer setup.Cleanup()

	t.Run("Invalid JSON for rule creation", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/classification/rules", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Missing required fields for rule", func(t *testing.T) {
		incompleteRule := map[string]interface{}{
			"name": "Incomplete Rule",
			// Missing required fields like conditions, layer, device_type
		}

		jsonData, _ := json.Marshal(incompleteRule)
		req := httptest.NewRequest("POST", "/api/v1/classification/rules", strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}