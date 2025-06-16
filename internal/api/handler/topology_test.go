package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/servak/topology-manager/internal/service"
	"github.com/servak/topology-manager/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTopologyHandler(t *testing.T) (*TopologyHandler, *testutil.TestSetup, huma.API) {
	setup := testutil.NewTestSetup(t)
	setup.SeedTestData(t)

	// Create services
	topologyService := service.NewTopologyService(setup.Repo, setup.Logger)

	// Create handler
	handler := NewTopologyHandler(topologyService, setup.Logger)

	// Create test API
	router := chi.NewRouter()
	config := huma.DefaultConfig("Test API", "1.0.0")
	api := humachi.New(router, config)

	// Register routes
	handler.Register(api)

	return handler, setup, api
}

func TestTopologyHandler_SearchDevices(t *testing.T) {
	handler, setup, api := setupTopologyHandler(t)
	defer setup.Cleanup()

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/devices/search?q=device&limit=10", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse response body
	var response struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
		Query   string                   `json:"query"`
		Limit   int                      `json:"limit"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find our test devices
	assert.GreaterOrEqual(t, response.Count, 3)
	assert.GreaterOrEqual(t, len(response.Devices), 3)
	assert.Equal(t, "device", response.Query)
	assert.Equal(t, 10, response.Limit)

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
	assert.True(t, deviceIDs["device-001"], "Should find device-001")
	assert.True(t, deviceIDs["device-002"], "Should find device-002")
	assert.True(t, deviceIDs["device-003"], "Should find device-003")
}

func TestTopologyHandler_SearchDevicesWithEmptyQuery(t *testing.T) {
	handler, setup, api := setupTopologyHandler(t)
	defer setup.Cleanup()

	// Create test request with empty query
	req := httptest.NewRequest("GET", "/api/v1/devices/search?q=&limit=5", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Should handle empty query gracefully
	assert.Equal(t, http.StatusOK, resp.Code)

	var response struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Empty query might return no results or all results depending on implementation
	assert.GreaterOrEqual(t, response.Count, 0)
}

func TestTopologyHandler_SearchDevicesByHardware(t *testing.T) {
	handler, setup, api := setupTopologyHandler(t)
	defer setup.Cleanup()

	// Search for devices by hardware pattern
	req := httptest.NewRequest("GET", "/api/v1/devices/search?q=Test%20Hardware&limit=10", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var response struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
		Query   string                   `json:"query"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find devices with "Test Hardware" in their hardware field
	assert.GreaterOrEqual(t, response.Count, 3)
	assert.Equal(t, "Test Hardware", response.Query)

	// Verify that found devices actually contain the search term
	for _, device := range response.Devices {
		if hardware, ok := device["hardware"].(string); ok {
			assert.Contains(t, hardware, "Test Hardware")
		}
	}
}

func TestTopologyHandler_SearchDevicesWithLimit(t *testing.T) {
	handler, setup, api := setupTopologyHandler(t)
	defer setup.Cleanup()

	// Search with small limit
	req := httptest.NewRequest("GET", "/api/v1/devices/search?q=device&limit=2", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var response struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
		Limit   int                      `json:"limit"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should respect the limit
	assert.LessOrEqual(t, len(response.Devices), 2)
	assert.LessOrEqual(t, response.Count, 2)
	assert.Equal(t, 2, response.Limit)
}

func TestTopologyHandler_SearchDevicesNonExistent(t *testing.T) {
	handler, setup, api := setupTopologyHandler(t)
	defer setup.Cleanup()

	// Search for non-existent devices
	req := httptest.NewRequest("GET", "/api/v1/devices/search?q=nonexistent&limit=10", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var response struct {
		Devices []map[string]interface{} `json:"devices"`
		Count   int                      `json:"count"`
		Query   string                   `json:"query"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return empty results
	assert.Equal(t, 0, response.Count)
	assert.Equal(t, 0, len(response.Devices))
	assert.Equal(t, "nonexistent", response.Query)
}

func TestTopologyHandler_SearchDevicesInvalidParameters(t *testing.T) {
	handler, setup, api := setupTopologyHandler(t)
	defer setup.Cleanup()

	t.Run("Invalid limit parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/devices/search?q=device&limit=invalid", nil)
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		// Should handle invalid limit gracefully (might use default or return error)
		// The exact behavior depends on the validation implementation
		assert.True(t, resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
	})

	t.Run("Negative limit parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/devices/search?q=device&limit=-1", nil)
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		// Should handle negative limit gracefully
		assert.True(t, resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
	})

	t.Run("Very large limit parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/devices/search?q=device&limit=10000", nil)
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		// Should handle large limit gracefully
		assert.Equal(t, http.StatusOK, resp.Code)

		var response struct {
			Devices []map[string]interface{} `json:"devices"`
			Count   int                      `json:"count"`
			Limit   int                      `json:"limit"`
		}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should either respect the limit or cap it at a reasonable maximum
		assert.GreaterOrEqual(t, response.Limit, 1)
	})
}

func TestTopologyHandler_SearchDevicesResponseFormat(t *testing.T) {
	handler, setup, api := setupTopologyHandler(t)
	defer setup.Cleanup()

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/devices/search?q=device-001&limit=5", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))

	// Parse and validate response structure
	var response struct {
		Devices []struct {
			ID           string                 `json:"id"`
			Type         string                 `json:"type"`
			Hardware     string                 `json:"hardware"`
			LayerID      *int                   `json:"layer_id"`
			DeviceType   string                 `json:"device_type"`
			ClassifiedBy string                 `json:"classified_by"`
			Metadata     map[string]interface{} `json:"metadata"`
			LastSeen     string                 `json:"last_seen"`
			CreatedAt    string                 `json:"created_at"`
			UpdatedAt    string                 `json:"updated_at"`
		} `json:"devices"`
		Count int    `json:"count"`
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find device-001
	assert.GreaterOrEqual(t, response.Count, 1)
	
	if len(response.Devices) > 0 {
		device := response.Devices[0]
		assert.NotEmpty(t, device.ID)
		assert.NotEmpty(t, device.Type)
		assert.NotEmpty(t, device.Hardware)
		assert.NotEmpty(t, device.DeviceType)
		assert.NotEmpty(t, device.LastSeen)
		assert.NotEmpty(t, device.CreatedAt)
		assert.NotEmpty(t, device.UpdatedAt)
	}
}