package service

import (
	"context"
	"testing"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/domain/visualization"
)

func TestDebugGroupCreation(t *testing.T) {
	repo := createTestTopology()
	service := NewVisualizationService(repo)

	ctx := context.Background()
	
	// ExtractSubTopologyの結果を確認
	devices, links, err := repo.ExtractSubTopology(ctx, "core-001", topology.SubTopologyOptions{Radius: 3})
	if err != nil {
		t.Fatalf("ExtractSubTopology failed: %v", err)
	}

	t.Logf("=== ExtractSubTopology Results ===")
	t.Logf("Devices count: %d", len(devices))
	for _, device := range devices {
		t.Logf("  Device: %s (type: %s, layer: %d)", device.ID, device.Type, device.Layer)
	}

	t.Logf("Links count: %d", len(links))
	for _, link := range links {
		t.Logf("  Link: %s -> %s", link.SourceID, link.TargetID)
	}

	// VisualNodeへの変換を確認
	visualNodes := make([]visualization.VisualNode, 0, len(devices))
	deviceDepthMap := service.calculateDeviceDepths(devices, links, "core-001")

	t.Logf("\n=== Device Depths ===")
	for deviceID, depth := range deviceDepthMap {
		t.Logf("  %s: depth %d", deviceID, depth)
	}

	for _, device := range devices {
		visualNode := visualization.VisualNode{
			ID:       device.ID,
			Name:     device.ID,
			Type:     device.Type,
			Hardware: device.Hardware,
			Status:   device.Status,
			Layer:    device.Layer,
			IsRoot:   device.ID == "core-001",
			Position: visualization.Position{X: 0, Y: 0},
		}
		visualNodes = append(visualNodes, visualNode)
	}

	t.Logf("\n=== Visual Nodes ===")
	for _, node := range visualNodes {
		t.Logf("  Node: %s (type: %s, isRoot: %v, layer: %d)", node.ID, node.Type, node.IsRoot, node.Layer)
	}

	// グループ作成の条件を確認
	groupingOpts := visualization.GroupingOptions{
		Enabled:       true,
		MinGroupSize:  3,
		MaxDepth:      2, // 深度2以上でグループ化
		GroupByPrefix: true,
		GroupByType:   false,
		PrefixMinLen:  3,
	}

	t.Logf("\n=== Grouping Options ===")
	t.Logf("  Enabled: %v", groupingOpts.Enabled)
	t.Logf("  MinGroupSize: %d", groupingOpts.MinGroupSize)
	t.Logf("  MaxDepth: %d", groupingOpts.MaxDepth)
	t.Logf("  GroupByPrefix: %v", groupingOpts.GroupByPrefix)
	t.Logf("  PrefixMinLen: %d", groupingOpts.PrefixMinLen)

	// 候補ノードを手動で計算
	candidateNodes := make([]visualization.VisualNode, 0)
	for _, node := range visualNodes {
		if !node.IsRoot && deviceDepthMap[node.ID] >= groupingOpts.MaxDepth {
			candidateNodes = append(candidateNodes, node)
			t.Logf("  Candidate node: %s (depth: %d)", node.ID, deviceDepthMap[node.ID])
		}
	}

	t.Logf("\n=== Candidate Nodes for Grouping ===")
	t.Logf("  Total candidates: %d", len(candidateNodes))
	t.Logf("  Minimum required: %d", groupingOpts.MinGroupSize)

	if len(candidateNodes) < groupingOpts.MinGroupSize {
		t.Logf("  -> Not enough candidates for grouping")
		return
	}

	// プレフィックスグループの手動確認
	deviceNames := make([]string, len(candidateNodes))
	for i, node := range candidateNodes {
		deviceNames[i] = node.Name
		t.Logf("  Device name for grouping: %s", node.Name)
	}

	// groupingパッケージを直接テスト
	t.Logf("\n=== Testing grouping package directly ===")
	// まず、groupingパッケージが正しく動作するかテスト
	groups := service.createGroups(candidateNodes, []visualization.VisualEdge{}, deviceDepthMap, groupingOpts)
	
	t.Logf("  Created groups: %d", len(groups))
	for i, group := range groups {
		t.Logf("  Group %d: %s (prefix: %s, count: %d, devices: %v)", 
			i, group.ID, group.Prefix, group.Count, group.DeviceIDs)
	}
}