package service

import (
	"context"
	"fmt"
	"time"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/domain/visualization"
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

	// BFS探索でトポロジーを構築
	devices, links, err := s.exploreTopology(ctx, rootDeviceID, rootDevice.Layer, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to explore topology: %w", err)
	}

	// 可視化用のノードとエッジに変換
	visualNodes := make([]visualization.VisualNode, 0, len(devices))
	nodeMap := make(map[string]*visualization.VisualNode)

	for _, device := range devices {
		visualNode := visualization.VisualNode{
			ID:       device.ID,
			Name:     device.Name,
			Type:     device.Type,
			Hardware: device.Hardware,
			Status:   device.Status,
			Layer:    device.Layer,
			IsRoot:   device.ID == rootDeviceID,
			Position: visualization.Position{X: 0, Y: 0}, // レイアウト計算で後から設定
			Style:    s.getNodeStyle(device.Type, device.Status, device.ID == rootDeviceID),
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
				Status:     link.Status,
				Weight:     link.Weight,
				Style:      s.getEdgeStyle(link.Status, link.Weight),
			}
			visualEdges = append(visualEdges, visualEdge)
		}
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
		TotalNodes: len(visualNodes),
		TotalEdges: len(visualEdges),
		Layers:     layerStats,
		Generated:  time.Now(),
	}

	return &visualization.VisualTopology{
		RootDevice: rootDeviceID,
		Depth:      depth,
		Timestamp:  time.Now().Unix(),
		Nodes:      visualNodes,
		Edges:      visualEdges,
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
			shouldInclude := s.shouldIncludeNeighbor(rootLayer, neighbor.Layer, current.level, depth)
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