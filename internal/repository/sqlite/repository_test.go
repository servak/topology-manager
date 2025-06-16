package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteRepository(t *testing.T) {
	// Create in-memory SQLite repository for testing
	config := Config{Path: ":memory:"}
	repo, err := NewSQliteRepository(config)
	require.NoError(t, err)
	defer repo.Close()

	// Run migrations
	err = repo.Migrate()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Health Check", func(t *testing.T) {
		err := repo.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("Add and Get Device", func(t *testing.T) {
		device := topology.Device{
			ID:           "test-device-01",
			Type:         "switch",
			Hardware:     "Arista 7280",
			DeviceType:   "core",
			ClassifiedBy: "user:admin",
			Metadata:     map[string]string{"location": "datacenter-1"},
			LastSeen:     time.Now(),
		}

		// Add device
		err := repo.AddDevice(ctx, device)
		assert.NoError(t, err)

		// Get device
		retrieved, err := repo.GetDevice(ctx, device.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, device.ID, retrieved.ID)
		assert.Equal(t, device.Type, retrieved.Type)
		assert.Equal(t, device.Hardware, retrieved.Hardware)
		assert.Equal(t, device.DeviceType, retrieved.DeviceType)
		assert.Equal(t, device.ClassifiedBy, retrieved.ClassifiedBy)
		assert.Equal(t, device.Metadata["location"], retrieved.Metadata["location"])
	})

	t.Run("Update Device", func(t *testing.T) {
		device := topology.Device{
			ID:           "test-device-02",
			Type:         "router",
			Hardware:     "Cisco ASR",
			DeviceType:   "distribution",
			ClassifiedBy: "rule:auto-router",
			Metadata:     map[string]string{"location": "datacenter-2"},
			LastSeen:     time.Now(),
		}

		// Add device
		err := repo.AddDevice(ctx, device)
		require.NoError(t, err)

		// Update device
		device.Hardware = "Cisco ASR 9000"
		device.Metadata["updated"] = "true"
		err = repo.UpdateDevice(ctx, device)
		assert.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetDevice(ctx, device.ID)
		require.NoError(t, err)
		assert.Equal(t, "Cisco ASR 9000", retrieved.Hardware)
		assert.Equal(t, "true", retrieved.Metadata["updated"])
	})

	t.Run("Search Devices", func(t *testing.T) {
		// Add test devices
		devices := []topology.Device{
			{ID: "search-test-01", Type: "switch", Hardware: "Arista 7050", LastSeen: time.Now()},
			{ID: "search-test-02", Type: "router", Hardware: "Cisco ISR", LastSeen: time.Now()},
			{ID: "search-test-03", Type: "switch", Hardware: "Arista 7280", LastSeen: time.Now()},
		}

		for _, device := range devices {
			err := repo.AddDevice(ctx, device)
			require.NoError(t, err)
		}

		// Search for Arista devices
		results, err := repo.SearchDevices(ctx, "Arista", 10)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		// Search for specific device ID
		results, err = repo.SearchDevices(ctx, "search-test-02", 10)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "search-test-02", results[0].ID)
	})

	t.Run("Add and Get Link", func(t *testing.T) {
		// First add devices for the link
		sourceDevice := topology.Device{
			ID:       "link-source",
			Type:     "switch",
			Hardware: "Test Switch",
			LastSeen: time.Now(),
		}
		targetDevice := topology.Device{
			ID:       "link-target",
			Type:     "switch",
			Hardware: "Test Switch",
			LastSeen: time.Now(),
		}

		err := repo.AddDevice(ctx, sourceDevice)
		require.NoError(t, err)
		err = repo.AddDevice(ctx, targetDevice)
		require.NoError(t, err)

		// Add link
		link := topology.Link{
			ID:         "test-link-01",
			SourceID:   "link-source",
			TargetID:   "link-target",
			SourcePort: "eth0",
			TargetPort: "eth1",
			Weight:     1.0,
			Metadata:   map[string]string{"type": "ethernet"},
			LastSeen:   time.Now(),
		}

		err = repo.AddLink(ctx, link)
		assert.NoError(t, err)

		// Get link
		retrieved, err := repo.GetLink(ctx, link.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, link.ID, retrieved.ID)
		assert.Equal(t, link.SourceID, retrieved.SourceID)
		assert.Equal(t, link.TargetID, retrieved.TargetID)
		assert.Equal(t, link.SourcePort, retrieved.SourcePort)
		assert.Equal(t, link.TargetPort, retrieved.TargetPort)
		assert.Equal(t, link.Weight, retrieved.Weight)
	})

	t.Run("Get Device Links", func(t *testing.T) {
		deviceID := "link-test-device"

		// Add test device
		device := topology.Device{
			ID:       deviceID,
			Type:     "switch",
			Hardware: "Test Switch",
			LastSeen: time.Now(),
		}
		err := repo.AddDevice(ctx, device)
		require.NoError(t, err)

		// Add another device for links
		otherDevice := topology.Device{
			ID:       "other-device",
			Type:     "switch",
			Hardware: "Other Switch",
			LastSeen: time.Now(),
		}
		err = repo.AddDevice(ctx, otherDevice)
		require.NoError(t, err)

		// Add links where device is source and target
		links := []topology.Link{
			{
				ID:         "outbound-link",
				SourceID:   deviceID,
				TargetID:   "other-device",
				SourcePort: "eth0",
				TargetPort: "eth0",
				Weight:     1.0,
				LastSeen:   time.Now(),
			},
			{
				ID:         "inbound-link",
				SourceID:   "other-device",
				TargetID:   deviceID,
				SourcePort: "eth1",
				TargetPort: "eth1",
				Weight:     1.0,
				LastSeen:   time.Now(),
			},
		}

		for _, link := range links {
			err := repo.AddLink(ctx, link)
			require.NoError(t, err)
		}

		// Get device links
		deviceLinks, err := repo.GetDeviceLinks(ctx, deviceID)
		require.NoError(t, err)
		assert.Len(t, deviceLinks, 2)
	})

	t.Run("Bulk Operations", func(t *testing.T) {
		// Bulk add devices
		devices := []topology.Device{
			{ID: "bulk-01", Type: "switch", Hardware: "Bulk Switch 1", LastSeen: time.Now()},
			{ID: "bulk-02", Type: "switch", Hardware: "Bulk Switch 2", LastSeen: time.Now()},
			{ID: "bulk-03", Type: "router", Hardware: "Bulk Router 1", LastSeen: time.Now()},
		}

		err := repo.BulkAddDevices(ctx, devices)
		assert.NoError(t, err)

		// Verify devices were added
		for _, device := range devices {
			retrieved, err := repo.GetDevice(ctx, device.ID)
			require.NoError(t, err)
			assert.Equal(t, device.ID, retrieved.ID)
		}

		// Bulk add links
		links := []topology.Link{
			{
				ID:         "bulk-link-01",
				SourceID:   "bulk-01",
				TargetID:   "bulk-02",
				SourcePort: "eth0",
				TargetPort: "eth0",
				Weight:     1.0,
				LastSeen:   time.Now(),
			},
			{
				ID:         "bulk-link-02",
				SourceID:   "bulk-02",
				TargetID:   "bulk-03",
				SourcePort: "eth1",
				TargetPort: "eth0",
				Weight:     1.0,
				LastSeen:   time.Now(),
			},
		}

		err = repo.BulkAddLinks(ctx, links)
		assert.NoError(t, err)

		// Verify links were added
		for _, link := range links {
			retrieved, err := repo.GetLink(ctx, link.ID)
			require.NoError(t, err)
			assert.Equal(t, link.ID, retrieved.ID)
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		// Add multiple devices for pagination test
		for i := 0; i < 15; i++ {
			device := topology.Device{
				ID:       fmt.Sprintf("page-test-%02d", i),
				Type:     "switch",
				Hardware: "Page Test Switch",
				LastSeen: time.Now(),
			}
			err := repo.AddDevice(ctx, device)
			require.NoError(t, err)
		}

		// Test pagination
		opts := topology.PaginationOptions{
			Page:     1,
			PageSize: 5,
			OrderBy:  "id",
			SortDir:  "ASC",
		}

		devices, pagination, err := repo.GetDevices(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, devices, 5)
		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 5, pagination.PageSize)
		assert.True(t, pagination.TotalCount >= 15) // At least 15 from this test
		assert.True(t, pagination.HasNext)
		assert.False(t, pagination.HasPrev)

		// Test second page
		opts.Page = 2
		devices, pagination, err = repo.GetDevices(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, devices, 5)
		assert.Equal(t, 2, pagination.Page)
		assert.True(t, pagination.HasPrev)
	})
}

func TestSQLiteConfig(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		config := Config{Path: "/tmp/test.db"}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("In-Memory Config", func(t *testing.T) {
		config := Config{Path: ":memory:"}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, ":memory:", config.DSN())
	})

	t.Run("Invalid Config", func(t *testing.T) {
		config := Config{Path: ""}
		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("DSN Generation", func(t *testing.T) {
		config := Config{Path: "/tmp/test.db"}
		assert.Equal(t, "/tmp/test.db", config.DSN())
	})
}
