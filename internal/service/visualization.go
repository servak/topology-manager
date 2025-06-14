package service

import (
	"context"
	"fmt"
	"time"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/domain/visualization"
	"github.com/servak/topology-manager/pkg/grouping"
)

type VisualizationService struct {
	topologyRepo topology.Repository
}

func NewVisualizationService(topologyRepo topology.Repository) *VisualizationService {
	return &VisualizationService{
		topologyRepo: topologyRepo,
	}
}

func (s *VisualizationService) GetVisualTopology(ctx context.Context, rootDeviceID string, depth int) (*visualization.VisualTopology, error) {
	return s.GetVisualTopologyWithGrouping(ctx, rootDeviceID, depth, visualization.GroupingOptions{
		Enabled: false,
	})
}

// GetSimpleVisualTopology returns a simplified visual topology without grouping for hierarchical display
func (s *VisualizationService) GetSimpleVisualTopology(ctx context.Context, rootDeviceID string, depth int) (*visualization.VisualTopology, error) {
	if depth <= 0 {
		depth = 3
	}

	// ルートデバイスの存在確認
	rootDevice, err := s.topologyRepo.GetDevice(ctx, rootDeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get root device: %w", err)
	}
	if rootDevice == nil {
		return nil, fmt.Errorf("root device %s not found", rootDeviceID)
	}

	// サブトポロジー抽出
	devices, links, err := s.topologyRepo.ExtractSubTopology(ctx, rootDeviceID, topology.SubTopologyOptions{
		Radius: depth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract sub-topology: %w", err)
	}

	// デバイスマップ作成（レイヤー情報の参照用）
	deviceMap := make(map[string]topology.Device)
	for _, device := range devices {
		deviceMap[device.ID] = device
	}

	// シンプルなビジュアルノード作成（グループ化なし）
	visualNodes := make([]visualization.VisualNode, 0, len(devices))
	nodeMap := make(map[string]*visualization.VisualNode, len(devices))

	for _, device := range devices {
		// 接続分類を追加
		connections := s.classifyConnections(ctx, device.ID, deviceMap, links)
		
		visualNode := visualization.VisualNode{
			ID:          device.ID,
			Name:        device.ID,
			Type:        device.Type,
			Hardware:    device.Hardware,
			Status:      "active", // default status since status field removed
			Layer:       s.getDeviceLayer(device.LayerID),
			IsRoot:      device.ID == rootDeviceID,
			Position:    visualization.Position{X: 0, Y: 0}, // レイアウト計算で後から設定
			Style:       s.getNodeStyle(device.Type, "active", device.ID == rootDeviceID),
			Connections: connections, // 新しい接続分類情報
		}
		visualNodes = append(visualNodes, visualNode)
		nodeMap[device.ID] = &visualNode
	}

	// シンプルなビジュアルエッジ作成
	visualEdges := make([]visualization.VisualEdge, 0, len(links))
	for _, link := range links {
		// 両方のノードが存在することを確認
		if nodeMap[link.SourceID] != nil && nodeMap[link.TargetID] != nil {
			// 接続タイプを決定
			connectionType := s.determineConnectionType(link, deviceMap)
			
			visualEdge := visualization.VisualEdge{
				ID:             link.ID,
				Source:         link.SourceID,
				Target:         link.TargetID,
				LocalPort:      link.SourcePort,
				RemotePort:     link.TargetPort,
				Status:         "active", // default status since status field removed
				Weight:         link.Weight,
				Style:          s.getEdgeStyle("active", link.Weight),
				ConnectionType: connectionType, // 新しい接続タイプ情報
			}
			visualEdges = append(visualEdges, visualEdge)
		}
	}

	// シンプルなレイアウト計算（階層ベース）
	s.calculateHierarchicalLayout(visualNodes, visualEdges, rootDeviceID)

	visualTopology := &visualization.VisualTopology{
		Nodes:      visualNodes,
		Edges:      visualEdges,
		RootDevice: rootDeviceID,
		Timestamp:  time.Now().Unix(),
	}

	return visualTopology, nil
}

func (s *VisualizationService) GetVisualTopologyWithGrouping(ctx context.Context, rootDeviceID string, depth int, groupingOpts visualization.GroupingOptions) (*visualization.VisualTopology, error) {
	if depth <= 0 {
		depth = 3
	}

	// ルートデバイスの存在確認
	rootDevice, err := s.topologyRepo.GetDevice(ctx, rootDeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get root device: %w", err)
	}
	if rootDevice == nil {
		return nil, fmt.Errorf("root device %s not found", rootDeviceID)
	}

	// 最適化されたサブトポロジー抽出を使用
	devices, links, err := s.topologyRepo.ExtractSubTopology(ctx, rootDeviceID, topology.SubTopologyOptions{
		Radius: depth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract sub-topology: %w", err)
	}

	// 可視化用のノードとエッジに変換
	visualNodes := make([]visualization.VisualNode, 0, len(devices))
	nodeMap := make(map[string]*visualization.VisualNode)
	deviceDepthMap := make(map[string]int)

	// ルートからの距離を計算
	deviceDepthMap = s.calculateDeviceDepths(devices, links, rootDeviceID)

	for _, device := range devices {
		visualNode := visualization.VisualNode{
			ID:       device.ID,
			Name:     device.ID, // IDをNameとして使用
			Type:     device.Type,
			Hardware: device.Hardware,
			Status:   "active", // default status since status field removed
			Layer:    s.getDeviceLayer(device.LayerID),
			IsRoot:   device.ID == rootDeviceID,
			Position: visualization.Position{X: 0, Y: 0}, // レイアウト計算で後から設定
			Style:    s.getNodeStyle(device.Type, "active", device.ID == rootDeviceID),
		}
		visualNodes = append(visualNodes, visualNode)
		nodeMap[device.ID] = &visualNode
	}

	visualEdges := make([]visualization.VisualEdge, 0, len(links))
	for _, link := range links {
		// 両方のノードが存在することを確認
		if nodeMap[link.SourceID] != nil && nodeMap[link.TargetID] != nil {
			visualEdge := visualization.VisualEdge{
				ID:         link.ID,
				Source:     link.SourceID,
				Target:     link.TargetID,
				LocalPort:  link.SourcePort,
				RemotePort: link.TargetPort,
				Status:     "active", // default status since status field removed
				Weight:     link.Weight,
				Style:      s.getEdgeStyle("active", link.Weight),
			}
			visualEdges = append(visualEdges, visualEdge)
		}
	}

	// グルーピング処理
	var groups []visualization.GroupedVisualNode
	if groupingOpts.Enabled {
		groups = s.createGroups(visualNodes, visualEdges, deviceDepthMap, groupingOpts)
		// グループ化されたノードを除外し、グループノードを追加
		visualNodes, visualEdges = s.applyGrouping(visualNodes, visualEdges, groups, rootDeviceID)
	}

	// レイアウト計算
	layout := s.calculateLayout(visualNodes, visualEdges, rootDeviceID)

	// 統計情報の計算
	layerStats := make(map[string]int)
	for _, node := range visualNodes {
		layerKey := fmt.Sprintf("%d", node.Layer)
		layerStats[layerKey]++
	}

	stats := visualization.TopologyStats{
		TotalNodes:  len(visualNodes),
		TotalEdges:  len(visualEdges),
		TotalGroups: len(groups),
		Layers:      layerStats,
		Generated:   time.Now(),
	}

	return &visualization.VisualTopology{
		RootDevice: rootDeviceID,
		Depth:      depth,
		Timestamp:  time.Now().Unix(),
		Nodes:      visualNodes,
		Edges:      visualEdges,
		Groups:     groups,
		Layout:     layout,
		Stats:      stats,
	}, nil
}

func (s *VisualizationService) exploreTopology(ctx context.Context, rootDeviceID string, rootLayer, depth int) ([]topology.Device, []topology.Link, error) {
	visited := make(map[string]bool)
	deviceMap := make(map[string]topology.Device)
	linkMap := make(map[string]topology.Link)

	queue := []struct {
		deviceID string
		level    int
	}{{rootDeviceID, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.deviceID] || current.level > depth {
			continue
		}

		visited[current.deviceID] = true

		// デバイス情報を取得
		device, err := s.topologyRepo.GetDevice(ctx, current.deviceID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get device %s: %w", current.deviceID, err)
		}
		if device == nil {
			continue
		}

		deviceMap[current.deviceID] = *device

		// デバイスのリンクを取得
		links, err := s.topologyRepo.GetDeviceLinks(ctx, current.deviceID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get links for device %s: %w", current.deviceID, err)
		}

		for _, link := range links {
			var neighborID string
			if link.SourceID == current.deviceID {
				neighborID = link.TargetID
			} else {
				neighborID = link.SourceID
			}

			// 隣接デバイス情報を取得
			neighbor, err := s.topologyRepo.GetDevice(ctx, neighborID)
			if err != nil {
				continue
			}
			if neighbor == nil {
				continue
			}

			// 階層に基づく包含判定
			shouldInclude := s.shouldIncludeNeighbor(rootLayer, s.getDeviceLayer(neighbor.LayerID), current.level, depth)
			if shouldInclude {
				// リンクを追加（重複チェック）
				linkKey := fmt.Sprintf("%s-%s", link.SourceID, link.TargetID)
				reverseLinkKey := fmt.Sprintf("%s-%s", link.TargetID, link.SourceID)

				if _, exists := linkMap[linkKey]; !exists {
					if _, exists := linkMap[reverseLinkKey]; !exists {
						linkMap[linkKey] = link
					}
				}

				// 未訪問の隣接デバイスをキューに追加
				if !visited[neighborID] {
					queue = append(queue, struct {
						deviceID string
						level    int
					}{neighborID, current.level + 1})
				}
			}
		}
	}

	// マップからスライスに変換
	devices := make([]topology.Device, 0, len(deviceMap))
	for _, device := range deviceMap {
		devices = append(devices, device)
	}

	links := make([]topology.Link, 0, len(linkMap))
	for _, link := range linkMap {
		links = append(links, link)
	}

	return devices, links, nil
}

func (s *VisualizationService) shouldIncludeNeighbor(rootLayer, neighborLayer, currentLevel, maxDepth int) bool {
	if currentLevel >= maxDepth {
		return false
	}

	// 上位階層（レイヤー値が小さい）の場合は2レベルまで
	if neighborLayer < rootLayer && currentLevel <= 2 {
		return true
	}

	// 同じ階層以下の場合は含める
	if neighborLayer >= rootLayer {
		return true
	}

	return false
}

func (s *VisualizationService) getNodeStyle(deviceType, status string, isRoot bool) visualization.NodeStyle {
	style := visualization.NodeStyle{
		Shape:       "ellipse",
		Size:        30,
		BorderWidth: 2,
	}

	// ルートノードは特別なスタイル
	if isRoot {
		style.Color = "#ff6b6b"
		style.BorderColor = "#d63447"
		style.Size = 40
		return style
	}

	// デバイスタイプ別の色分け
	switch deviceType {
	case "switch":
		style.Color = "#4ecdc4"
		style.BorderColor = "#26d0ce"
	case "router":
		style.Color = "#45b7d1"
		style.BorderColor = "#2980b9"
	case "server":
		style.Color = "#f9ca24"
		style.BorderColor = "#f0932b"
	default:
		style.Color = "#95a5a6"
		style.BorderColor = "#7f8c8d"
	}

	// ステータス別の調整
	if status == "down" || status == "error" {
		style.Color = "#e74c3c"
		style.BorderColor = "#c0392b"
	}

	return style
}

func (s *VisualizationService) getEdgeStyle(status string, weight float64) visualization.EdgeStyle {
	style := visualization.EdgeStyle{
		Width:     2,
		LineStyle: "solid",
	}

	// ステータス別の色分け
	switch status {
	case "up", "active":
		style.Color = "#2ecc71"
	case "down", "inactive":
		style.Color = "#e74c3c"
	default:
		style.Color = "#95a5a6"
	}

	// 重みに基づく線の太さ調整
	if weight > 10 {
		style.Width = 4
	} else if weight > 5 {
		style.Width = 3
	}

	return style
}

func (s *VisualizationService) calculateLayout(nodes []visualization.VisualNode, edges []visualization.VisualEdge, rootDeviceID string) visualization.Layout {
	// 基本的な階層レイアウトを実装
	positions := make(map[string]visualization.Position)

	// 階層ごとにノードを分類
	layerNodes := make(map[int][]visualization.VisualNode)
	for _, node := range nodes {
		layerNodes[node.Layer] = append(layerNodes[node.Layer], node)
	}

	// Y座標は階層に基づいて設定
	layerY := 0.0
	layerSpacing := 150.0

	for layer := 0; layer <= 10; layer++ { // 最大10階層まで
		if nodesInLayer, exists := layerNodes[layer]; exists {
			nodeSpacing := 200.0
			totalWidth := float64(len(nodesInLayer)-1) * nodeSpacing
			startX := -totalWidth / 2

			for i, node := range nodesInLayer {
				x := startX + float64(i)*nodeSpacing
				positions[node.ID] = visualization.Position{
					X: x,
					Y: layerY,
				}
			}
			layerY += layerSpacing
		}
	}

	return visualization.Layout{
		Type: "hierarchical",
		Options: map[string]interface{}{
			"direction": "top-to-bottom",
			"spacing":   layerSpacing,
		},
		Positions: positions,
	}
}

// calculateDeviceDepths calculates the depth of each device from the root
func (s *VisualizationService) calculateDeviceDepths(devices []topology.Device, links []topology.Link, rootDeviceID string) map[string]int {
	depthMap := make(map[string]int)
	visited := make(map[string]bool)

	// Build adjacency list
	adjList := make(map[string][]string)
	for _, link := range links {
		adjList[link.SourceID] = append(adjList[link.SourceID], link.TargetID)
		adjList[link.TargetID] = append(adjList[link.TargetID], link.SourceID)
	}

	// BFS to calculate depths
	queue := []struct {
		deviceID string
		depth    int
	}{{rootDeviceID, 0}}

	depthMap[rootDeviceID] = 0
	visited[rootDeviceID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighborID := range adjList[current.deviceID] {
			if !visited[neighborID] {
				visited[neighborID] = true
				depthMap[neighborID] = current.depth + 1
				queue = append(queue, struct {
					deviceID string
					depth    int
				}{neighborID, current.depth + 1})
			}
		}
	}

	return depthMap
}

// createGroups creates groups based on grouping options
func (s *VisualizationService) createGroups(nodes []visualization.VisualNode, edges []visualization.VisualEdge, deviceDepthMap map[string]int, opts visualization.GroupingOptions) []visualization.GroupedVisualNode {
	var groups []visualization.GroupedVisualNode

	if opts.MinGroupSize <= 1 {
		opts.MinGroupSize = 3 // デフォルト最小グループサイズ
	}

	// 深度によるフィルタリング
	candidateNodes := make([]visualization.VisualNode, 0)
	for _, node := range nodes {
		if !node.IsRoot && deviceDepthMap[node.ID] >= opts.MaxDepth {
			candidateNodes = append(candidateNodes, node)
		}
	}

	if len(candidateNodes) < opts.MinGroupSize {
		return groups
	}

	// プレフィックスによるグルーピング
	if opts.GroupByPrefix {
		deviceNames := make([]string, len(candidateNodes))
		deviceNodeMap := make(map[string]visualization.VisualNode)
		for i, node := range candidateNodes {
			deviceNames[i] = node.Name
			deviceNodeMap[node.Name] = node
		}

		prefixGroups := grouping.GroupByLongestCommonPrefix(deviceNames, opts.MinGroupSize)
		for i, group := range prefixGroups {
			if len(group.Prefix) >= opts.PrefixMinLen {
				groupID := fmt.Sprintf("group-prefix-%d", i)
				groupNode := visualization.GroupedVisualNode{
					ID:         groupID,
					Name:       fmt.Sprintf("%s* (%d)", group.Prefix, group.Count),
					Type:       "group",
					GroupType:  "prefix",
					Prefix:     group.Prefix,
					Count:      group.Count,
					DeviceIDs:  group.DeviceIDs,
					Depth:      opts.MaxDepth,
					IsExpanded: false,
					Position:   visualization.Position{X: 0, Y: 0},
					Style: visualization.GroupedNodeStyle{
						Color:       "#95a5a6",
						Shape:       "round-rectangle",
						Size:        50,
						BorderColor: "#7f8c8d",
						BorderWidth: 3,
						Label:       fmt.Sprintf("%s* (%d)", group.Prefix, group.Count),
					},
					InternalEdgeCount: s.countInternalEdges(group.DeviceIDs, edges),
					ExternalEdges:     s.findExternalEdges(group.DeviceIDs, edges),
				}
				groups = append(groups, groupNode)
			}
		}
	}

	// タイプによるグルーピング
	if opts.GroupByType {
		deviceTypes := make(map[string]string)
		for _, node := range candidateNodes {
			deviceTypes[node.ID] = node.Type
		}

		typeGroups := grouping.GroupByType(deviceTypes)
		for i, group := range typeGroups {
			if group.Count >= opts.MinGroupSize {
				groupID := fmt.Sprintf("group-type-%d", i)
				groupNode := visualization.GroupedVisualNode{
					ID:         groupID,
					Name:       fmt.Sprintf("%s (%d)", group.Prefix, group.Count),
					Type:       "group",
					GroupType:  "type",
					Prefix:     group.Prefix,
					Count:      group.Count,
					DeviceIDs:  group.DeviceIDs,
					Depth:      opts.MaxDepth,
					IsExpanded: false,
					Position:   visualization.Position{X: 0, Y: 0},
					Style: visualization.GroupedNodeStyle{
						Color:       "#3498db",
						Shape:       "round-rectangle",
						Size:        50,
						BorderColor: "#2980b9",
						BorderWidth: 3,
						Label:       fmt.Sprintf("%s (%d)", group.Prefix, group.Count),
					},
					InternalEdgeCount: s.countInternalEdges(group.DeviceIDs, edges),
					ExternalEdges:     s.findExternalEdges(group.DeviceIDs, edges),
				}
				groups = append(groups, groupNode)
			}
		}
	}

	return groups
}

// applyGrouping applies grouping by removing grouped nodes and adding group nodes
func (s *VisualizationService) applyGrouping(nodes []visualization.VisualNode, edges []visualization.VisualEdge, groups []visualization.GroupedVisualNode, rootDeviceID string) ([]visualization.VisualNode, []visualization.VisualEdge) {
	if len(groups) == 0 {
		return nodes, edges
	}

	// グループ化されるデバイスIDのセットを作成
	groupedDeviceIDs := make(map[string]bool)
	for _, group := range groups {
		for _, deviceID := range group.DeviceIDs {
			groupedDeviceIDs[deviceID] = true
		}
	}

	// グループ化されたノードの先のノードも特定
	nodesAfterGroups := make(map[string]bool)
	s.findNodesAfterGroups(nodes, edges, groupedDeviceIDs, nodesAfterGroups, rootDeviceID)

	// グループ化されないノードを保持（グループの先のノードも除外、ただしルートノードは除外しない）
	filteredNodes := make([]visualization.VisualNode, 0)
	for _, node := range nodes {
		if !groupedDeviceIDs[node.ID] && (!nodesAfterGroups[node.ID] || node.IsRoot) {
			filteredNodes = append(filteredNodes, node)
		}
	}

	// グループノードを追加（VisualNodeとして）
	for _, group := range groups {
		groupVisualNode := visualization.VisualNode{
			ID:       group.ID,
			Name:     group.Name,
			Type:     "group",
			Hardware: fmt.Sprintf("Group of %d devices", group.Count),
			Status:   "active",
			Layer:    0, // グループノードの階層
			IsRoot:   false,
			Position: group.Position,
			Style: visualization.NodeStyle{
				Color:       group.Style.Color,
				Shape:       group.Style.Shape,
				Size:        group.Style.Size,
				BorderColor: group.Style.BorderColor,
				BorderWidth: group.Style.BorderWidth,
			},
		}
		filteredNodes = append(filteredNodes, groupVisualNode)
	}

	// シンプルなエッジ変換アプローチ
	filteredEdges := make([]visualization.VisualEdge, 0)
	edgeIDMap := make(map[string]bool) // 重複エッジを防ぐ

	fmt.Printf("Processing %d edges for grouping\n", len(edges))

	for _, edge := range edges {
		sourceGrouped := groupedDeviceIDs[edge.Source]
		targetGrouped := groupedDeviceIDs[edge.Target]

		fmt.Printf("Edge %s: %s->%s, sourceGrouped=%v, targetGrouped=%v\n",
			edge.ID, edge.Source, edge.Target, sourceGrouped, targetGrouped)

		// Case 1: 両方ともグループ化されていない → そのまま保持
		if !sourceGrouped && !targetGrouped {
			filteredEdges = append(filteredEdges, edge)
			fmt.Printf("  Kept original edge\n")
			continue
		}

		// Case 2: ソースのみグループ化 → グループ->ターゲットのエッジ作成
		if sourceGrouped && !targetGrouped {
			groupID := s.findGroupIDForDevice(edge.Source, groups)
			if groupID != "" {
				newEdgeID := fmt.Sprintf("%s-%s", groupID, edge.Target)
				if !edgeIDMap[newEdgeID] {
					newEdge := visualization.VisualEdge{
						ID:         newEdgeID,
						Source:     groupID,
						Target:     edge.Target,
						LocalPort:  "group",
						RemotePort: edge.RemotePort,
						Status:     edge.Status,
						Weight:     edge.Weight,
						Style:      edge.Style,
					}
					filteredEdges = append(filteredEdges, newEdge)
					edgeIDMap[newEdgeID] = true
					fmt.Printf("  Created group edge: %s->%s\n", groupID, edge.Target)
				}
			}
			continue
		}

		// Case 3: ターゲットのみグループ化 → ソース->グループのエッジ作成
		if !sourceGrouped && targetGrouped {
			groupID := s.findGroupIDForDevice(edge.Target, groups)
			if groupID != "" {
				newEdgeID := fmt.Sprintf("%s-%s", edge.Source, groupID)
				if !edgeIDMap[newEdgeID] {
					newEdge := visualization.VisualEdge{
						ID:         newEdgeID,
						Source:     edge.Source,
						Target:     groupID,
						LocalPort:  edge.LocalPort,
						RemotePort: "group",
						Status:     edge.Status,
						Weight:     edge.Weight,
						Style:      edge.Style,
					}
					filteredEdges = append(filteredEdges, newEdge)
					edgeIDMap[newEdgeID] = true
					fmt.Printf("  Created group edge: %s->%s\n", edge.Source, groupID)
				}
			}
			continue
		}

		// Case 4: 両方がグループ化 → 内部エッジなので除外
		fmt.Printf("  Skipped internal edge\n")
	}

	fmt.Printf("Final filtered edges: %d\n", len(filteredEdges))

	return filteredNodes, filteredEdges
}

// countInternalEdges counts edges within a group
func (s *VisualizationService) countInternalEdges(deviceIDs []string, edges []visualization.VisualEdge) int {
	deviceSet := make(map[string]bool)
	for _, id := range deviceIDs {
		deviceSet[id] = true
	}

	count := 0
	for _, edge := range edges {
		if deviceSet[edge.Source] && deviceSet[edge.Target] {
			count++
		}
	}
	return count
}

// findExternalEdges finds edges connecting to devices outside the group
func (s *VisualizationService) findExternalEdges(deviceIDs []string, edges []visualization.VisualEdge) []string {
	deviceSet := make(map[string]bool)
	for _, id := range deviceIDs {
		deviceSet[id] = true
	}

	var externalEdges []string
	for _, edge := range edges {
		sourceInGroup := deviceSet[edge.Source]
		targetInGroup := deviceSet[edge.Target]

		// 片方だけがグループ内にある場合は外部エッジ
		if (sourceInGroup && !targetInGroup) || (!sourceInGroup && targetInGroup) {
			externalEdges = append(externalEdges, edge.ID)
		}
	}
	return externalEdges
}

// findGroupIDForDevice finds the group ID that contains the given device
func (s *VisualizationService) findGroupIDForDevice(deviceID string, groups []visualization.GroupedVisualNode) string {
	for _, group := range groups {
		for _, id := range group.DeviceIDs {
			if id == deviceID {
				return group.ID
			}
		}
	}
	return ""
}

// findNodesAfterGroups identifies nodes that are only reachable through grouped nodes
func (s *VisualizationService) findNodesAfterGroups(nodes []visualization.VisualNode, edges []visualization.VisualEdge, groupedDeviceIDs map[string]bool, nodesAfterGroups map[string]bool, rootDeviceID string) {
	// Convert VisualEdge to Link for calculateDeviceDepths
	links := make([]topology.Link, len(edges))
	for i, edge := range edges {
		links[i] = topology.Link{
			SourceID: edge.Source,
			TargetID: edge.Target,
		}
	}

	// Convert VisualNode to Device for calculateDeviceDepths
	devices := make([]topology.Device, len(nodes))
	for i, node := range nodes {
		devices[i] = topology.Device{
			ID: node.ID,
		}
	}

	// Create a map of node depths from root
	deviceDepthMap := s.calculateDeviceDepths(devices, links, rootDeviceID)

	// Find the maximum depth of grouped devices
	maxGroupDepth := 0
	for deviceID := range groupedDeviceIDs {
		if depth, exists := deviceDepthMap[deviceID]; exists && depth > maxGroupDepth {
			maxGroupDepth = depth
		}
	}

	// Any non-grouped node that is deeper than the max group depth
	// and only reachable through grouped nodes should be hidden
	for _, node := range nodes {
		if !groupedDeviceIDs[node.ID] && !node.IsRoot {
			if nodeDepth, exists := deviceDepthMap[node.ID]; exists && nodeDepth > maxGroupDepth {
				// Check if this node is only reachable through grouped nodes
				if s.isOnlyReachableThroughGroups(node.ID, edges, groupedDeviceIDs) {
					nodesAfterGroups[node.ID] = true
				}
			}
		}
	}
}

// isOnlyReachableThroughGroups checks if a node can only be reached through grouped nodes
func (s *VisualizationService) isOnlyReachableThroughGroups(nodeID string, edges []visualization.VisualEdge, groupedDeviceIDs map[string]bool) bool {
	// Find all direct neighbors of this node
	neighbors := make([]string, 0)
	for _, edge := range edges {
		if edge.Source == nodeID {
			neighbors = append(neighbors, edge.Target)
		} else if edge.Target == nodeID {
			neighbors = append(neighbors, edge.Source)
		}
	}

	// If all neighbors are grouped devices, then this node is only reachable through groups
	for _, neighbor := range neighbors {
		if !groupedDeviceIDs[neighbor] {
			return false // Found a non-grouped neighbor
		}
	}

	return len(neighbors) > 0 // Only return true if there are neighbors (avoid isolated nodes)
}

// ExpandGroupInTopology expands a group node by replacing it with its constituent devices and their neighbors
func (s *VisualizationService) ExpandGroupInTopology(
	ctx context.Context,
	groupID string,
	rootDeviceID string,
	groupDeviceIDs []string,
	currentTopology visualization.VisualTopology,
	groupingOpts visualization.GroupingOptions,
	expandDepth int,
) (*visualization.VisualTopology, []visualization.VisualNode, []visualization.VisualEdge, error) {
	if expandDepth <= 0 {
		expandDepth = 2
	}

	// グループ内のデバイスとその近傍を取得
	expandedDevices := make(map[string]topology.Device)
	expandedLinks := make(map[string]topology.Link)

	// グループ内のデバイスを出発点として探索
	for _, deviceID := range groupDeviceIDs {
		// デバイス自体を追加
		device, err := s.topologyRepo.GetDevice(ctx, deviceID)
		if err != nil || device == nil {
			continue
		}
		expandedDevices[deviceID] = *device

		// 指定された深度まで近傍を探索
		neighbors, links, err := s.exploreFromDevice(ctx, deviceID, expandDepth)
		if err != nil {
			continue
		}

		// 結果をマージ
		for _, neighbor := range neighbors {
			expandedDevices[neighbor.ID] = neighbor
		}
		for _, link := range links {
			linkKey := fmt.Sprintf("%s-%s", link.SourceID, link.TargetID)
			expandedLinks[linkKey] = link
		}
	}

	// 展開されたデバイスを可視化ノードに変換
	newVisualNodes := make([]visualization.VisualNode, 0)
	deviceDepthMap := make(map[string]int)

	// ルートからの距離を再計算
	allDevices := make([]topology.Device, 0, len(expandedDevices))
	allLinks := make([]topology.Link, 0, len(expandedLinks))
	for _, device := range expandedDevices {
		allDevices = append(allDevices, device)
	}
	for _, link := range expandedLinks {
		allLinks = append(allLinks, link)
	}
	deviceDepthMap = s.calculateDeviceDepths(allDevices, allLinks, rootDeviceID)

	// 新しいノードを作成
	fmt.Printf("ExpandGroupInTopology: Found %d expanded devices\n", len(expandedDevices))
	for _, device := range expandedDevices {
		exists := s.nodeExistsInTopology(device.ID, currentTopology)
		fmt.Printf("Device %s exists in topology: %v\n", device.ID, exists)
		// 既存のトポロジーに含まれていないノードのみ追加
		if !exists {
			visualNode := visualization.VisualNode{
				ID:       device.ID,
				Name:     device.ID,
				Type:     device.Type,
				Hardware: device.Hardware,
				Status:   "active", // default status since status field removed
				Layer:    s.getDeviceLayer(device.LayerID),
				IsRoot:   device.ID == rootDeviceID,
				Position: visualization.Position{X: 0, Y: 0},
				Style:    s.getNodeStyle(device.Type, "active", device.ID == rootDeviceID),
			}
			newVisualNodes = append(newVisualNodes, visualNode)
			fmt.Printf("Added new visual node: %s\n", device.ID)
		}
	}
	fmt.Printf("Total new visual nodes created: %d\n", len(newVisualNodes))

	// 新しいエッジを作成
	newVisualEdges := make([]visualization.VisualEdge, 0)
	for _, link := range expandedLinks {
		// 既存のトポロジーに含まれていないエッジのみ追加
		if !s.edgeExistsInTopology(link.ID, currentTopology) {
			visualEdge := visualization.VisualEdge{
				ID:         link.ID,
				Source:     link.SourceID,
				Target:     link.TargetID,
				LocalPort:  link.SourcePort,
				RemotePort: link.TargetPort,
				Status:     "active", // default status since status field removed
				Weight:     link.Weight,
				Style:      s.getEdgeStyle("active", link.Weight),
			}
			newVisualEdges = append(newVisualEdges, visualEdge)
		}
	}

	// 現在のトポロジーを更新
	updatedTopology := currentTopology

	// グループノードを削除
	filteredNodes := make([]visualization.VisualNode, 0)
	for _, node := range updatedTopology.Nodes {
		if node.ID != groupID {
			filteredNodes = append(filteredNodes, node)
		}
	}

	// グループノードに接続されていたエッジを削除
	filteredEdges := make([]visualization.VisualEdge, 0)
	for _, edge := range updatedTopology.Edges {
		if edge.Source != groupID && edge.Target != groupID {
			filteredEdges = append(filteredEdges, edge)
		}
	}

	// 新しいノードとエッジを追加
	fmt.Printf("Before adding: filteredNodes=%d, newVisualNodes=%d\n", len(filteredNodes), len(newVisualNodes))
	updatedTopology.Nodes = append(filteredNodes, newVisualNodes...)
	updatedTopology.Edges = append(filteredEdges, newVisualEdges...)
	fmt.Printf("After adding: updatedTopology.Nodes=%d\n", len(updatedTopology.Nodes))

	// グループ情報を更新（展開されたグループを削除）
	filteredGroups := make([]visualization.GroupedVisualNode, 0)
	for _, group := range updatedTopology.Groups {
		if group.ID != groupID {
			filteredGroups = append(filteredGroups, group)
		}
	}
	updatedTopology.Groups = filteredGroups

	// 新しく追加されたノードに対して再帰的なグルーピングを適用
	if groupingOpts.Enabled {
		fmt.Printf("Applying recursive grouping...\n")
		// 新しいノードの中で深度が条件を満たすものをグルーピング対象とする
		candidateNodes := make([]visualization.VisualNode, 0)
		for _, node := range newVisualNodes {
			depth := deviceDepthMap[node.ID]
			fmt.Printf("Node %s depth=%d, maxDepth=%d\n", node.ID, depth, groupingOpts.MaxDepth)
			if !node.IsRoot && depth >= groupingOpts.MaxDepth {
				candidateNodes = append(candidateNodes, node)
			}
		}
		fmt.Printf("Candidate nodes for grouping: %d (min required: %d)\n", len(candidateNodes), groupingOpts.MinGroupSize)

		if len(candidateNodes) >= groupingOpts.MinGroupSize {
			newGroups := s.createGroups(candidateNodes, newVisualEdges, deviceDepthMap, groupingOpts)
			fmt.Printf("Created %d new groups\n", len(newGroups))
			if len(newGroups) > 0 {
				// 新しいグループを適用
				fmt.Printf("Before recursive grouping: %d nodes\n", len(updatedTopology.Nodes))
				groupedNodes, groupedEdges := s.applyGrouping(updatedTopology.Nodes, updatedTopology.Edges, newGroups, rootDeviceID)
				updatedTopology.Nodes = groupedNodes
				updatedTopology.Edges = groupedEdges
				updatedTopology.Groups = append(updatedTopology.Groups, newGroups...)
				fmt.Printf("After recursive grouping: %d nodes\n", len(updatedTopology.Nodes))
			}
		}
	}

	// 統計情報を更新
	updatedTopology.Stats = visualization.TopologyStats{
		TotalNodes:  len(updatedTopology.Nodes),
		TotalEdges:  len(updatedTopology.Edges),
		TotalGroups: len(updatedTopology.Groups),
		Generated:   time.Now(),
	}

	// レイアウトを再計算
	updatedTopology.Layout = s.calculateLayout(updatedTopology.Nodes, updatedTopology.Edges, rootDeviceID)

	return &updatedTopology, newVisualNodes, newVisualEdges, nil
}

// exploreFromDevice explores topology from a specific device up to a given depth
func (s *VisualizationService) exploreFromDevice(ctx context.Context, deviceID string, depth int) ([]topology.Device, []topology.Link, error) {
	visited := make(map[string]bool)
	deviceMap := make(map[string]topology.Device)
	linkMap := make(map[string]topology.Link)

	queue := []struct {
		deviceID string
		level    int
	}{{deviceID, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.deviceID] || current.level > depth {
			continue
		}

		visited[current.deviceID] = true

		// デバイス情報を取得
		device, err := s.topologyRepo.GetDevice(ctx, current.deviceID)
		if err != nil || device == nil {
			continue
		}
		deviceMap[current.deviceID] = *device

		// 隣接するリンクを取得
		links, err := s.topologyRepo.GetDeviceLinks(ctx, current.deviceID)
		if err != nil {
			continue
		}

		for _, link := range links {
			var neighborID string
			if link.SourceID == current.deviceID {
				neighborID = link.TargetID
			} else {
				neighborID = link.SourceID
			}

			// リンクを追加
			linkKey := fmt.Sprintf("%s-%s", link.SourceID, link.TargetID)
			reverseLinkKey := fmt.Sprintf("%s-%s", link.TargetID, link.SourceID)
			if _, exists := linkMap[linkKey]; !exists {
				if _, exists := linkMap[reverseLinkKey]; !exists {
					linkMap[linkKey] = link
				}
			}

			// 隣接デバイスをキューに追加
			if !visited[neighborID] && current.level < depth {
				queue = append(queue, struct {
					deviceID string
					level    int
				}{neighborID, current.level + 1})
			}
		}
	}

	// マップからスライスに変換
	devices := make([]topology.Device, 0, len(deviceMap))
	for _, device := range deviceMap {
		devices = append(devices, device)
	}

	links := make([]topology.Link, 0, len(linkMap))
	for _, link := range linkMap {
		links = append(links, link)
	}

	return devices, links, nil
}

// nodeExistsInTopology checks if a node already exists in the current topology
func (s *VisualizationService) nodeExistsInTopology(nodeID string, topology visualization.VisualTopology) bool {
	for _, node := range topology.Nodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

// edgeExistsInTopology checks if an edge already exists in the current topology
func (s *VisualizationService) edgeExistsInTopology(edgeID string, topology visualization.VisualTopology) bool {
	for _, edge := range topology.Edges {
		if edge.ID == edgeID {
			return true
		}
	}
	return false
}

// getDeviceLayer converts layer_id to layer value for compatibility
func (s *VisualizationService) getDeviceLayer(layerID *int) int {
	if layerID == nil {
		return 5 // default to server layer if not specified
	}
	return *layerID
}

// classifyConnections classifies device connections into uplinks, downlinks, and peers
func (s *VisualizationService) classifyConnections(ctx context.Context, deviceID string, deviceMap map[string]topology.Device, links []topology.Link) *visualization.ConnectionClassification {
	device, exists := deviceMap[deviceID]
	if !exists {
		return &visualization.ConnectionClassification{}
	}

	deviceLayer := s.getDeviceLayer(device.LayerID)
	
	var uplinks []visualization.ConnectionInfo
	var downlinks []visualization.ConnectionInfo
	var peers []visualization.ConnectionInfo

	// デバイスに接続されているリンクを探す
	for _, link := range links {
		var connectedDeviceID string
		var localPort, remotePort string

		if link.SourceID == deviceID {
			connectedDeviceID = link.TargetID
			localPort = link.SourcePort
			remotePort = link.TargetPort
		} else if link.TargetID == deviceID {
			connectedDeviceID = link.SourceID
			localPort = link.TargetPort
			remotePort = link.SourcePort
		} else {
			continue
		}

		connectedDevice, exists := deviceMap[connectedDeviceID]
		if !exists {
			continue
		}

		connectedLayer := s.getDeviceLayer(connectedDevice.LayerID)
		
		// 接続情報の構築
		connInfo := visualization.ConnectionInfo{
			DeviceID:        connectedDeviceID,
			DeviceName:      connectedDevice.ID,
			DeviceType:      connectedDevice.Type,
			DeviceHardware:  connectedDevice.Hardware,
			Layer:           connectedLayer,
			LocalPort:       localPort,
			RemotePort:      remotePort,
			Status:          "active", // デフォルト
			LinkWeight:      link.Weight,
		}

		// 階層レベルに基づく分類
		if connectedLayer < deviceLayer {
			// 上位階層への接続 = uplink
			uplinks = append(uplinks, connInfo)
		} else if connectedLayer > deviceLayer {
			// 下位階層への接続 = downlink
			downlinks = append(downlinks, connInfo)
		} else {
			// 同一階層への接続 = peer
			// さらに同じグループ（同じdownlinkに接続）かどうかを判定
			connInfo.IsSameGroup = s.isInSameGroup(ctx, deviceID, connectedDeviceID, deviceMap, links)
			peers = append(peers, connInfo)
		}
	}

	return &visualization.ConnectionClassification{
		Uplinks:   uplinks,
		Downlinks: downlinks,
		Peers:     peers,
	}
}

// determineConnectionType determines the type of connection between two devices
func (s *VisualizationService) determineConnectionType(link topology.Link, deviceMap map[string]topology.Device) string {
	sourceDevice, sourceExists := deviceMap[link.SourceID]
	targetDevice, targetExists := deviceMap[link.TargetID]
	
	if !sourceExists || !targetExists {
		return "unknown"
	}

	sourceLayer := s.getDeviceLayer(sourceDevice.LayerID)
	targetLayer := s.getDeviceLayer(targetDevice.LayerID)

	if sourceLayer < targetLayer {
		return "uplink" // source が上位階層、target が下位階層
	} else if sourceLayer > targetLayer {
		return "downlink" // source が下位階層、target が上位階層
	} else {
		return "peer" // 同一階層
	}
}

// isInSameGroup checks if two devices in the same layer are in the same group
// (connected to the same uplink devices)
func (s *VisualizationService) isInSameGroup(ctx context.Context, device1ID, device2ID string, deviceMap map[string]topology.Device, links []topology.Link) bool {
	// device1のuplinkを取得
	device1Uplinks := s.getUplinkDevices(device1ID, deviceMap, links)
	// device2のuplinkを取得
	device2Uplinks := s.getUplinkDevices(device2ID, deviceMap, links)

	// 共通するuplinkがあるかチェック
	for _, uplink1 := range device1Uplinks {
		for _, uplink2 := range device2Uplinks {
			if uplink1 == uplink2 {
				return true
			}
		}
	}

	return false
}

// getUplinkDevices returns list of uplink device IDs for a given device
func (s *VisualizationService) getUplinkDevices(deviceID string, deviceMap map[string]topology.Device, links []topology.Link) []string {
	device, exists := deviceMap[deviceID]
	if !exists {
		return []string{}
	}

	deviceLayer := s.getDeviceLayer(device.LayerID)
	var uplinks []string

	for _, link := range links {
		var connectedDeviceID string

		if link.SourceID == deviceID {
			connectedDeviceID = link.TargetID
		} else if link.TargetID == deviceID {
			connectedDeviceID = link.SourceID
		} else {
			continue
		}

		connectedDevice, exists := deviceMap[connectedDeviceID]
		if !exists {
			continue
		}

		connectedLayer := s.getDeviceLayer(connectedDevice.LayerID)
		
		// 上位階層のデバイスのみ追加
		if connectedLayer < deviceLayer {
			uplinks = append(uplinks, connectedDeviceID)
		}
	}

	return uplinks
}

// calculateHierarchicalLayout calculates positions for hierarchical display
func (s *VisualizationService) calculateHierarchicalLayout(nodes []visualization.VisualNode, edges []visualization.VisualEdge, rootDeviceID string) {
	// 階層別にノードを分類
	layers := make(map[int][]int) // layer -> node indices

	for i, node := range nodes {
		layer := node.Layer
		if layers[layer] == nil {
			layers[layer] = make([]int, 0)
		}
		layers[layer] = append(layers[layer], i)
	}

	// Y座標は階層別に設定
	layerSpacing := 150.0
	nodeSpacing := 120.0

	for layer, nodeIndices := range layers {
		y := float64(layer) * layerSpacing

		// X座標は同一階層内で均等配置
		startX := -(float64(len(nodeIndices)-1) * nodeSpacing) / 2

		for i, nodeIndex := range nodeIndices {
			x := startX + float64(i)*nodeSpacing
			nodes[nodeIndex].Position.X = x
			nodes[nodeIndex].Position.Y = y
		}
	}
}
