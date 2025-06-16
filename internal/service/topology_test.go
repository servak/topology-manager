package service

import (
	"context"
	"testing"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopologyService_AddAndGetDevice(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	topologyService := NewTopologyService(setup.Repo)

	ctx := context.Background()

	// Create a test device
	testDevice := testutil.CreateTestDevice("test-device-001")

	// Add device through service
	err := topologyService.AddDevice(ctx, testDevice)
	require.NoError(t, err)

	// Retrieve device
	retrievedDevice, err := topologyService.GetDevice(ctx, testDevice.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedDevice)

	// Verify device properties
	assert.Equal(t, testDevice.ID, retrievedDevice.ID)
	assert.Equal(t, testDevice.Type, retrievedDevice.Type)
	assert.Equal(t, testDevice.Hardware, retrievedDevice.Hardware)
	assert.Equal(t, testDevice.DeviceType, retrievedDevice.DeviceType)
}

func TestTopologyService_GetDevices(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()
	setup.SeedTestData(t)

	topologyService := NewTopologyService(setup.Repo)

	ctx := context.Background()

	// Test getting devices with pagination
	paginationOpts := topology.PaginationOptions{
		Page:     1,
		PageSize: 2,
	}

	devices, paginationResult, err := topologyService.GetDevices(ctx, paginationOpts)
	require.NoError(t, err)
	require.NotNil(t, paginationResult)

	// Should get up to 2 devices
	assert.LessOrEqual(t, len(devices), 2)
	assert.GreaterOrEqual(t, len(devices), 1)

	// Check pagination result
	assert.Equal(t, 1, paginationResult.Page)
	assert.Equal(t, 2, paginationResult.PageSize)
	assert.GreaterOrEqual(t, paginationResult.TotalCount, 3) // At least our test data
}

func TestTopologyService_SearchDevices(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()
	setup.SeedTestData(t)

	topologyService := NewTopologyService(setup.Repo)

	ctx := context.Background()

	// Search for devices by ID pattern
	devices, err := topologyService.SearchDevices(ctx, "device-00", 10)
	require.NoError(t, err)

	// Should find our test devices
	assert.GreaterOrEqual(t, len(devices), 3)

	// Verify we found the right devices
	foundIDs := make(map[string]bool)
	for _, device := range devices {
		foundIDs[device.ID] = true
	}
	assert.True(t, foundIDs["device-001"])
	assert.True(t, foundIDs["device-002"])
	assert.True(t, foundIDs["device-003"])
}

func TestTopologyService_SearchDevicesByHardware(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()
	setup.SeedTestData(t)

	topologyService := NewTopologyService(setup.Repo)

	ctx := context.Background()

	// Search for devices by hardware (all test devices have "Test Hardware" in their hardware field)
	devices, err := topologyService.SearchDevices(ctx, "Test Hardware", 10)
	require.NoError(t, err)

	// Should find our test devices
	assert.GreaterOrEqual(t, len(devices), 3)

	// Verify all found devices have the expected hardware pattern
	for _, device := range devices {
		assert.Contains(t, device.Hardware, "Test Hardware")
	}
}

func TestTopologyService_AddAndGetLink(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	topologyService := NewTopologyService(setup.Repo)

	ctx := context.Background()

	// First add devices
	device1 := testutil.CreateTestDevice("device-link-001")
	device2 := testutil.CreateTestDevice("device-link-002")
	
	err := topologyService.AddDevice(ctx, device1)
	require.NoError(t, err)
	err = topologyService.AddDevice(ctx, device2)
	require.NoError(t, err)

	// Create and add link
	testLink := testutil.CreateTestLink("test-link-001", device1.ID, device2.ID)
	err = topologyService.AddLink(ctx, testLink)
	require.NoError(t, err)

	// Retrieve link
	retrievedLink, err := topologyService.GetLink(ctx, testLink.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedLink)

	// Verify link properties
	assert.Equal(t, testLink.ID, retrievedLink.ID)
	assert.Equal(t, testLink.SourceID, retrievedLink.SourceID)
	assert.Equal(t, testLink.TargetID, retrievedLink.TargetID)
	assert.Equal(t, testLink.SourcePort, retrievedLink.SourcePort)
	assert.Equal(t, testLink.TargetPort, retrievedLink.TargetPort)
}

func TestTopologyService_GetDeviceLinks(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()
	setup.SeedTestData(t)

	topologyService := NewTopologyService(setup.Repo)

	ctx := context.Background()

	// Get links for device-001 (should have link to device-002)
	links, err := topologyService.GetDeviceLinks(ctx, "device-001")
	require.NoError(t, err)

	// Should have at least one link
	assert.GreaterOrEqual(t, len(links), 1)

	// Find the link between device-001 and device-002
	foundLink := false
	for _, link := range links {
		if (link.SourceID == "device-001" && link.TargetID == "device-002") ||
			(link.SourceID == "device-002" && link.TargetID == "device-001") {
			foundLink = true
			break
		}
	}
	assert.True(t, foundLink, "Should find link between device-001 and device-002")
}

func TestTopologyService_RemoveDevice(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	ctx := context.Background()

	// Add a device
	testDevice := testutil.CreateTestDevice("device-to-remove")
	err := setup.TopologyService.AddDevice(ctx, testDevice)
	require.NoError(t, err)

	// Verify it exists
	retrievedDevice, err := setup.TopologyService.GetDevice(ctx, testDevice.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedDevice)

	// Remove the device
	err = setup.TopologyService.RemoveDevice(ctx, testDevice.ID)
	require.NoError(t, err)

	// Verify it's gone
	removedDevice, err := setup.TopologyService.GetDevice(ctx, testDevice.ID)
	require.NoError(t, err)
	assert.Nil(t, removedDevice)
}

func TestTopologyService_BulkAddDevices(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	ctx := context.Background()

	// Create multiple test devices
	devices := []topology.Device{
		testutil.CreateTestDevice("bulk-device-001"),
		testutil.CreateTestDevice("bulk-device-002"),
		testutil.CreateTestDevice("bulk-device-003"),
	}

	// Bulk add devices
	err := setup.TopologyService.BulkAddDevices(ctx, devices)
	require.NoError(t, err)

	// Verify all devices were added
	for _, device := range devices {
		retrievedDevice, err := setup.TopologyService.GetDevice(ctx, device.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedDevice)
		assert.Equal(t, device.ID, retrievedDevice.ID)
	}
}

func TestTopologyService_FindDevicesByType(t *testing.T) {
	setup := testutil.NewTestSetup(t)
	defer setup.Cleanup()

	ctx := context.Background()

	// Add devices with different types
	switchDevice := testutil.CreateTestDevice("switch-001")
	switchDevice.DeviceType = "network-switch"
	
	routerDevice := testutil.CreateTestDevice("router-001")
	routerDevice.DeviceType = "network-router"

	err := setup.TopologyService.AddDevice(ctx, switchDevice)
	require.NoError(t, err)
	err = setup.TopologyService.AddDevice(ctx, routerDevice)
	require.NoError(t, err)

	// Find devices by type
	switches, err := setup.TopologyService.FindDevicesByType(ctx, "network-switch")
	require.NoError(t, err)

	// Should find at least our switch device (plus any from seed data)
	foundSwitch := false
	for _, device := range switches {
		if device.ID == "switch-001" {
			foundSwitch = true
			assert.Equal(t, "network-switch", device.DeviceType)
		}
	}
	assert.True(t, foundSwitch, "Should find the switch device")
}