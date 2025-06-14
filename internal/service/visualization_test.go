package service

import (
	"context"
	"testing"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/domain/visualization"
)

// MockTopologyRepository はテスト用のモックリポジトリ
type MockTopologyRepository struct {
	devices map[string]*topology.Device
	links   []topology.Link
}

func NewMockTopologyRepository() *MockTopologyRepository {
	return &MockTopologyRepository{
		devices: make(map[string]*topology.Device),
		links:   make([]topology.Link, 0),
	}
}

func (m *MockTopologyRepository) AddDevice(device *topology.Device) {
	m.devices[device.ID] = device
}

func (m *MockTopologyRepository) AddLink(link topology.Link) {
	m.links = append(m.links, link)
}

func (m *MockTopologyRepository) GetDevice(ctx context.Context, deviceID string) (*topology.Device, error) {
	device, exists := m.devices[deviceID]
	if !exists {
		return nil, nil
	}
	return device, nil
}

func (m *MockTopologyRepository) ExtractSubTopology(ctx context.Context, rootDeviceID string, opts topology.SubTopologyOptions) ([]topology.Device, []topology.Link, error) {
	// シンプルに全デバイスと全リンクを返す（実際のテストでは適切にフィルタリング）
	devices := make([]topology.Device, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, *device)
	}
	return devices, m.links, nil
}

func (m *MockTopologyRepository) GetDeviceLinks(ctx context.Context, deviceID string) ([]topology.Link, error) {
	links := make([]topology.Link, 0)
	for _, link := range m.links {
		if link.SourceID == deviceID || link.TargetID == deviceID {
			links = append(links, link)
		}
	}
	return links, nil
}

// 他の必要なメソッドもモック実装（簡略化）
func (m *MockTopologyRepository) SearchDevices(ctx context.Context, query string, limit int) ([]topology.Device, error) { return nil, nil }
func (m *MockTopologyRepository) UpdateDevice(ctx context.Context, device topology.Device) error { return nil }
func (m *MockTopologyRepository) FindReachableDevices(ctx context.Context, deviceID string, opts topology.ReachabilityOptions) ([]topology.Device, error) { return nil, nil }
func (m *MockTopologyRepository) FindShortestPath(ctx context.Context, fromID, toID string, opts topology.PathOptions) (*topology.Path, error) { return nil, nil }
func (m *MockTopologyRepository) BulkAddDevices(ctx context.Context, devices []topology.Device) error { return nil }
func (m *MockTopologyRepository) BulkAddLinks(ctx context.Context, links []topology.Link) error { return nil }
func (m *MockTopologyRepository) Close() error { return nil }
func (m *MockTopologyRepository) Health(ctx context.Context) error { return nil }

// createTestTopology はテスト用のトポロジーを作成
func createTestTopology() *MockTopologyRepository {
	repo := NewMockTopologyRepository()

	// テストデバイスを追加
	repo.AddDevice(&topology.Device{
		ID:       "core-001",
		Type:     "core",
		Hardware: "Arista DCS-7280SR-48C6",
		Status:   "up",
		Layer:    3,
	})

	repo.AddDevice(&topology.Device{
		ID:       "core-002",
		Type:     "core",
		Hardware: "Arista DCS-7280SR-48C6",
		Status:   "up",
		Layer:    3,
	})

	// 同じプレフィックスを持つdist デバイス群（グループ化されるはず）
	repo.AddDevice(&topology.Device{
		ID:       "dist-100",
		Type:     "distribution",
		Hardware: "Juniper QFX5100",
		Status:   "up",
		Layer:    4,
	})

	repo.AddDevice(&topology.Device{
		ID:       "dist-101",
		Type:     "distribution", 
		Hardware: "Juniper QFX5100",
		Status:   "up",
		Layer:    4,
	})

	repo.AddDevice(&topology.Device{
		ID:       "dist-102",
		Type:     "distribution",
		Hardware: "Juniper QFX5100", 
		Status:   "up",
		Layer:    4,
	})

	// 同じプレフィックスを持つaccess デバイス群（グループ化されるはず）
	repo.AddDevice(&topology.Device{
		ID:       "access-001",
		Type:     "access",
		Hardware: "Cisco 2960X",
		Status:   "up",
		Layer:    5,
	})

	repo.AddDevice(&topology.Device{
		ID:       "access-002",
		Type:     "access",
		Hardware: "Cisco 2960X",
		Status:   "up", 
		Layer:    5,
	})

	repo.AddDevice(&topology.Device{
		ID:       "access-003",
		Type:     "access",
		Hardware: "Cisco 2960X",
		Status:   "up",
		Layer:    5,
	})

	// リンクを追加
	repo.AddLink(topology.Link{
		ID:         "core-001_core-002",
		SourceID:   "core-001",
		TargetID:   "core-002",
		SourcePort: "Ethernet1/1",
		TargetPort: "Ethernet1/1",
		Status:     "up",
		Weight:     1.0,
	})

	repo.AddLink(topology.Link{
		ID:         "core-001_dist-100",
		SourceID:   "core-001",
		TargetID:   "dist-100",
		SourcePort: "Ethernet1/2",
		TargetPort: "et-0/0/0",
		Status:     "up",
		Weight:     1.0,
	})

	repo.AddLink(topology.Link{
		ID:         "core-001_dist-101",
		SourceID:   "core-001", 
		TargetID:   "dist-101",
		SourcePort: "Ethernet1/3",
		TargetPort: "et-0/0/0",
		Status:     "up",
		Weight:     1.0,
	})

	repo.AddLink(topology.Link{
		ID:         "core-001_dist-102",
		SourceID:   "core-001",
		TargetID:   "dist-102", 
		SourcePort: "Ethernet1/4",
		TargetPort: "et-0/0/0",
		Status:     "up",
		Weight:     1.0,
	})

	repo.AddLink(topology.Link{
		ID:         "dist-100_access-001",
		SourceID:   "dist-100",
		TargetID:   "access-001",
		SourcePort: "et-0/0/1",
		TargetPort: "GigabitEthernet0/1",
		Status:     "up",
		Weight:     1.0,
	})

	repo.AddLink(topology.Link{
		ID:         "dist-101_access-002", 
		SourceID:   "dist-101",
		TargetID:   "access-002",
		SourcePort: "et-0/0/1",
		TargetPort: "GigabitEthernet0/1",
		Status:     "up",
		Weight:     1.0,
	})

	repo.AddLink(topology.Link{
		ID:         "dist-102_access-003",
		SourceID:   "dist-102",
		TargetID:   "access-003",
		SourcePort: "et-0/0/1", 
		TargetPort: "GigabitEthernet0/1",
		Status:     "up",
		Weight:     1.0,
	})

	return repo
}

func TestGetVisualTopologyWithGrouping_Basic(t *testing.T) {
	repo := createTestTopology()
	service := NewVisualizationService(repo)

	ctx := context.Background()
	groupingOpts := visualization.GroupingOptions{
		Enabled:       true,
		MinGroupSize:  3,
		MaxDepth:      2, // 深度2以上でグループ化
		GroupByPrefix: true,
		GroupByType:   false,
		PrefixMinLen:  3,
	}

	result, err := service.GetVisualTopologyWithGrouping(ctx, "core-001", 3, groupingOpts)
	if err != nil {
		t.Fatalf("GetVisualTopologyWithGrouping failed: %v", err)
	}

	t.Logf("Result nodes: %d", len(result.Nodes))
	t.Logf("Result edges: %d", len(result.Edges))
	t.Logf("Result groups: %d", len(result.Groups))

	// デバッグ: 各ノードの詳細を出力
	t.Logf("\n=== All Nodes ===")
	for _, node := range result.Nodes {
		t.Logf("Node: %s (type: %s, isRoot: %v)", node.ID, node.Type, node.IsRoot)
	}

	// グループが作成されているかチェック
	if len(result.Groups) == 0 {
		t.Error("Expected groups to be created, but got 0 groups")
	}

	// グループノードが含まれているかチェック
	groupNodeCount := 0
	for _, node := range result.Nodes {
		if node.Type == "group" {
			groupNodeCount++
			t.Logf("Group node found: %s (%s)", node.ID, node.Name)
		}
	}

	if groupNodeCount == 0 {
		t.Error("Expected group nodes in result, but found none")
	}

	// エッジが適切に処理されているかチェック
	if len(result.Edges) == 0 {
		t.Error("Expected edges in result, but found none")
	}

	// グループノードに接続されたエッジがあるかチェック
	groupEdgeCount := 0
	for _, edge := range result.Edges {
		// ソースまたはターゲットがグループノードの場合
		sourceIsGroup := false
		targetIsGroup := false
		
		for _, node := range result.Nodes {
			if node.Type == "group" {
				if edge.Source == node.ID {
					sourceIsGroup = true
				}
				if edge.Target == node.ID {
					targetIsGroup = true
				}
			}
		}
		
		if sourceIsGroup || targetIsGroup {
			groupEdgeCount++
			t.Logf("Group edge found: %s -> %s", edge.Source, edge.Target)
		}
	}

	if groupEdgeCount == 0 {
		t.Error("Expected edges connected to group nodes, but found none")
	}

	t.Logf("Group edge count: %d", groupEdgeCount)
}

func TestGetVisualTopologyWithGrouping_NoGrouping(t *testing.T) {
	repo := createTestTopology()
	service := NewVisualizationService(repo)

	ctx := context.Background()
	groupingOpts := visualization.GroupingOptions{
		Enabled: false,
	}

	result, err := service.GetVisualTopologyWithGrouping(ctx, "core-001", 3, groupingOpts)
	if err != nil {
		t.Fatalf("GetVisualTopologyWithGrouping failed: %v", err)
	}

	// グループが作成されていないことを確認
	if len(result.Groups) != 0 {
		t.Errorf("Expected 0 groups when grouping disabled, got %d", len(result.Groups))
	}

	// 全デバイスがノードとして表示されていることを確認
	expectedDevices := 8 // core-001, core-002, dist-100,101,102, access-001,002,003
	if len(result.Nodes) != expectedDevices {
		t.Errorf("Expected %d nodes when grouping disabled, got %d", expectedDevices, len(result.Nodes))
	}

	// 全リンクがエッジとして表示されていることを確認
	expectedLinks := 7 // 作成したリンク数
	if len(result.Edges) != expectedLinks {
		t.Errorf("Expected %d edges when grouping disabled, got %d", expectedLinks, len(result.Edges))
	}
}

func TestCalculateDeviceDepths(t *testing.T) {
	repo := createTestTopology()
	service := NewVisualizationService(repo)

	ctx := context.Background()
	devices, links, err := repo.ExtractSubTopology(ctx, "core-001", topology.SubTopologyOptions{Radius: 3})
	if err != nil {
		t.Fatalf("ExtractSubTopology failed: %v", err)
	}

	depthMap := service.calculateDeviceDepths(devices, links, "core-001")

	// 深度の確認
	expectedDepths := map[string]int{
		"core-001":   0, // ルート
		"core-002":   1, // 1ホップ
		"dist-100":   1, // 1ホップ 
		"dist-101":   1, // 1ホップ
		"dist-102":   1, // 1ホップ
		"access-001": 2, // 2ホップ
		"access-002": 2, // 2ホップ
		"access-003": 2, // 2ホップ
	}

	for deviceID, expectedDepth := range expectedDepths {
		actualDepth, exists := depthMap[deviceID]
		if !exists {
			t.Errorf("Expected device %s to have depth, but not found in depthMap", deviceID)
			continue
		}
		if actualDepth != expectedDepth {
			t.Errorf("Device %s: expected depth %d, got %d", deviceID, expectedDepth, actualDepth)
		}
	}

	t.Logf("Device depths: %+v", depthMap)
}

func TestGroupEdgeCreation(t *testing.T) {
	repo := createTestTopology()
	service := NewVisualizationService(repo)

	ctx := context.Background()
	groupingOpts := visualization.GroupingOptions{
		Enabled:       true,
		MinGroupSize:  3,
		MaxDepth:      1, // 深度1以上でグループ化（より多くグループ化）
		GroupByPrefix: true,
		GroupByType:   false,
		PrefixMinLen:  3,
	}

	result, err := service.GetVisualTopologyWithGrouping(ctx, "core-001", 3, groupingOpts)
	if err != nil {
		t.Fatalf("GetVisualTopologyWithGrouping failed: %v", err)
	}

	// 詳細ログ出力
	t.Logf("=== Test Result Analysis ===")
	t.Logf("Total nodes: %d", len(result.Nodes))
	t.Logf("Total edges: %d", len(result.Edges))
	t.Logf("Total groups: %d", len(result.Groups))

	t.Logf("\n=== Nodes ===")
	for _, node := range result.Nodes {
		t.Logf("Node: %s (type: %s, isRoot: %v)", node.ID, node.Type, node.IsRoot)
	}

	t.Logf("\n=== Groups ===")
	for _, group := range result.Groups {
		t.Logf("Group: %s (prefix: %s, count: %d, devices: %v)", 
			group.ID, group.Prefix, group.Count, group.DeviceIDs)
	}

	t.Logf("\n=== Edges ===")
	for _, edge := range result.Edges {
		t.Logf("Edge: %s -> %s (id: %s)", edge.Source, edge.Target, edge.ID)
	}

	// グループエッジの詳細検証
	groupNodes := make(map[string]bool)
	for _, node := range result.Nodes {
		if node.Type == "group" {
			groupNodes[node.ID] = true
		}
	}

	connectedGroupNodes := make(map[string]bool)
	for _, edge := range result.Edges {
		if groupNodes[edge.Source] {
			connectedGroupNodes[edge.Source] = true
		}
		if groupNodes[edge.Target] {
			connectedGroupNodes[edge.Target] = true
		}
	}

	t.Logf("\n=== Group Connection Analysis ===")
	t.Logf("Group nodes: %d", len(groupNodes))
	t.Logf("Connected group nodes: %d", len(connectedGroupNodes))

	for groupID := range groupNodes {
		if !connectedGroupNodes[groupID] {
			t.Errorf("Group node %s is not connected to any edges", groupID)
		}
	}

	// 最低限のエッジがあることを確認
	if len(result.Edges) == 0 {
		t.Error("Expected at least some edges in the result")
	}
}