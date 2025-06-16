package handler

import (
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

func setupVisualizationHandler(t *testing.T) (*VisualizationHandler, *testutil.TestSetup, huma.API) {
	setup := testutil.NewTestSetup(t)
	setup.SeedTestData(t)

	// Create services
	visualizationService := service.NewVisualizationService(setup.Repo, setup.Logger)

	// Create handler
	handler := NewVisualizationHandler(visualizationService, setup.Logger)

	// Create test API
	router := chi.NewRouter()
	config := huma.DefaultConfig("Test API", "1.0.0")
	api := humachi.New(router, config)

	// Register routes
	handler.Register(api)

	return handler, setup, api
}

func TestVisualizationHandler_GetTopology(t *testing.T) {
	handler, setup, api := setupVisualizationHandler(t)
	defer setup.Cleanup()

	// Create test request for device-001 with default depth
	req := httptest.NewRequest("GET", "/api/v1/topology/device-001", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))

	// Parse response body
	var response struct {
		RootDevice string `json:"root_device"`
		Depth      int    `json:"depth"`
		Stats      struct {
			TotalNodes int `json:"total_nodes"`
			TotalEdges int `json:"total_edges"`
			MaxDepth   int `json:"max_depth"`
		} `json:"stats"`
		Nodes []struct {
			ID       string `json:"id"`
			Label    string `json:"label"`
			Position struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"position"`
			Style struct {
				BackgroundColor string `json:"backgroundColor"`
				Shape          string `json:"shape"`
			} `json:"style"`
			Data struct {
				Type     string `json:"type"`
				Hardware string `json:"hardware"`
			} `json:"data"`
		} `json:"nodes"`
		Edges []struct {
			ID     string `json:"id"`
			Source string `json:"source"`
			Target string `json:"target"`
			Style  struct {
				LineColor string `json:"lineColor"`
			} `json:"style"`
			Data struct {
				Weight float64 `json:"weight"`
			} `json:"data"`
		} `json:"edges"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify basic topology structure
	assert.Equal(t, "device-001", response.RootDevice)
	assert.GreaterOrEqual(t, response.Depth, 0)
	assert.GreaterOrEqual(t, response.Stats.TotalNodes, 1)
	assert.GreaterOrEqual(t, len(response.Nodes), 1)

	// Check that root device is in nodes
	foundRootDevice := false
	for _, node := range response.Nodes {
		if node.ID == "device-001" {
			foundRootDevice = true
			assert.NotEmpty(t, node.Label)
			assert.NotEmpty(t, node.Style.BackgroundColor)
			assert.NotEmpty(t, node.Data.Type)
			break
		}
	}
	assert.True(t, foundRootDevice, "Should find root device in nodes")

	// Verify stats consistency
	assert.Equal(t, len(response.Nodes), response.Stats.TotalNodes)
	assert.Equal(t, len(response.Edges), response.Stats.TotalEdges)
}

func TestVisualizationHandler_GetTopologyWithDepth(t *testing.T) {
	handler, setup, api := setupVisualizationHandler(t)
	defer setup.Cleanup()

	// Create test request with specific depth
	req := httptest.NewRequest("GET", "/api/v1/topology/device-002?depth=1", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var response struct {
		RootDevice string `json:"root_device"`
		Depth      int    `json:"depth"`
		Stats      struct {
			TotalNodes int `json:"total_nodes"`
			TotalEdges int `json:"total_edges"`
			MaxDepth   int `json:"max_depth"`
		} `json:"stats"`
		Nodes []map[string]interface{} `json:"nodes"`
		Edges []map[string]interface{} `json:"edges"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have device-002 as root with depth 1
	assert.Equal(t, "device-002", response.RootDevice)
	assert.Equal(t, 1, response.Depth)
	assert.Equal(t, 1, response.Stats.MaxDepth)

	// Should have multiple nodes (device-002 plus connected devices)
	assert.GreaterOrEqual(t, response.Stats.TotalNodes, 2)
	assert.GreaterOrEqual(t, len(response.Nodes), 2)

	// Should have edges representing connections
	assert.GreaterOrEqual(t, response.Stats.TotalEdges, 1)
	assert.GreaterOrEqual(t, len(response.Edges), 1)
}

func TestVisualizationHandler_GetTopologyWithDepthZero(t *testing.T) {
	handler, setup, api := setupVisualizationHandler(t)
	defer setup.Cleanup()

	// Create test request with depth 0
	req := httptest.NewRequest("GET", "/api/v1/topology/device-001?depth=0", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var response struct {
		RootDevice string `json:"root_device"`
		Depth      int    `json:"depth"`
		Stats      struct {
			TotalNodes int `json:"total_nodes"`
			TotalEdges int `json:"total_edges"`
		} `json:"stats"`
		Nodes []map[string]interface{} `json:"nodes"`
		Edges []map[string]interface{} `json:"edges"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have only the root device
	assert.Equal(t, "device-001", response.RootDevice)
	assert.Equal(t, 0, response.Depth)
	assert.Equal(t, 1, response.Stats.TotalNodes)
	assert.Equal(t, 1, len(response.Nodes))
	assert.Equal(t, 0, response.Stats.TotalEdges)
	assert.Equal(t, 0, len(response.Edges))
}

func TestVisualizationHandler_GetTopologyNonExistentDevice(t *testing.T) {
	handler, setup, api := setupVisualizationHandler(t)
	defer setup.Cleanup()

	// Create test request for non-existent device
	req := httptest.NewRequest("GET", "/api/v1/topology/non-existent-device", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Should return error or empty topology
	if resp.Code == http.StatusNotFound || resp.Code == http.StatusBadRequest {
		// Error response is acceptable
		var errorResponse struct {
			Title  string `json:"title"`
			Status int    `json:"status"`
			Detail string `json:"detail"`
		}
		err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse.Detail, "not found")
	} else if resp.Code == http.StatusOK {
		// Empty topology is also acceptable
		var response struct {
			Stats struct {
				TotalNodes int `json:"total_nodes"`
			} `json:"stats"`
		}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 0, response.Stats.TotalNodes)
	} else {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}

func TestVisualizationHandler_GetTopologyInvalidDepth(t *testing.T) {
	handler, setup, api := setupVisualizationHandler(t)
	defer setup.Cleanup()

	t.Run("Negative depth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/topology/device-001?depth=-1", nil)
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		// Should handle gracefully (either use default or return error)
		assert.True(t, resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
	})

	t.Run("Invalid depth format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/topology/device-001?depth=invalid", nil)
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		// Should handle gracefully
		assert.True(t, resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
	})

	t.Run("Very large depth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/topology/device-001?depth=1000", nil)
		resp := httptest.NewRecorder()

		api.ServeHTTP(resp, req)

		// Should handle gracefully (might cap the depth)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestVisualizationHandler_GetTopologyResponseStructure(t *testing.T) {
	handler, setup, api := setupVisualizationHandler(t)
	defer setup.Cleanup()

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/topology/device-002?depth=1", nil)
	resp := httptest.NewRecorder()

	// Execute request
	api.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse and validate detailed response structure
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check required top-level fields
	assert.Contains(t, response, "root_device")
	assert.Contains(t, response, "depth")
	assert.Contains(t, response, "stats")
	assert.Contains(t, response, "nodes")
	assert.Contains(t, response, "edges")

	// Check stats structure
	stats, ok := response["stats"].(map[string]interface{})
	require.True(t, ok, "Stats should be an object")
	assert.Contains(t, stats, "total_nodes")
	assert.Contains(t, stats, "total_edges")
	assert.Contains(t, stats, "max_depth")

	// Check nodes structure
	nodes, ok := response["nodes"].([]interface{})
	require.True(t, ok, "Nodes should be an array")
	
	if len(nodes) > 0 {
		node := nodes[0].(map[string]interface{})
		assert.Contains(t, node, "id")
		assert.Contains(t, node, "label")
		assert.Contains(t, node, "position")
		assert.Contains(t, node, "style")
		assert.Contains(t, node, "data")

		// Check position structure
		position := node["position"].(map[string]interface{})
		assert.Contains(t, position, "x")
		assert.Contains(t, position, "y")

		// Check style structure
		style := node["style"].(map[string]interface{})
		assert.Contains(t, style, "backgroundColor")

		// Check data structure
		data := node["data"].(map[string]interface{})
		assert.Contains(t, data, "type")
	}

	// Check edges structure if any edges exist
	edges, ok := response["edges"].([]interface{})
	require.True(t, ok, "Edges should be an array")

	if len(edges) > 0 {
		edge := edges[0].(map[string]interface{})
		assert.Contains(t, edge, "id")
		assert.Contains(t, edge, "source")
		assert.Contains(t, edge, "target")
		assert.Contains(t, edge, "style")
		assert.Contains(t, edge, "data")
	}
}