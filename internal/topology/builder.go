package topology

import (
	"context"
	"fmt"
	"time"

	"github.com/servak/topology-manager/internal/storage"
)

type TopologyBuilder struct {
	redis *storage.RedisClient
}

type Topology struct {
	RootDevice string            `json:"root_device"`
	Depth      int               `json:"depth"`
	Timestamp  int64             `json:"timestamp"`
	Nodes      []Node            `json:"nodes"`
	Edges      []Edge            `json:"edges"`
	Stats      TopologyStats     `json:"stats"`
}

type Node struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Hardware string `json:"hardware"`
	Status string `json:"status"`
	Layer  int    `json:"layer"`
	IsRoot bool   `json:"is_root"`
}

type Edge struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	LocalPort  string `json:"local_port"`
	RemotePort string `json:"remote_port"`
	Status     string `json:"status"`
}

type TopologyStats struct {
	TotalNodes int            `json:"total_nodes"`
	TotalEdges int            `json:"total_edges"`
	Layers     map[string]int `json:"layers"`
}

func NewTopologyBuilder(redis *storage.RedisClient) *TopologyBuilder {
	return &TopologyBuilder{redis: redis}
}

func (tb *TopologyBuilder) BuildTopology(ctx context.Context, rootDevice string, depth int) (*Topology, error) {
	if depth <= 0 {
		depth = 3
	}

	rootDeviceInfo, err := tb.redis.GetDevice(ctx, rootDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to get root device: %w", err)
	}
	if rootDeviceInfo == nil {
		return nil, fmt.Errorf("root device %s not found", rootDevice)
	}

	visited := make(map[string]bool)
	nodeMap := make(map[string]*Node)
	var edges []Edge

	if err := tb.exploreFromRoot(ctx, rootDevice, rootDeviceInfo.Layer, depth, visited, nodeMap, &edges); err != nil {
		return nil, fmt.Errorf("failed to explore topology: %w", err)
	}

	nodes := make([]Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}

	layerStats := make(map[string]int)
	for _, node := range nodes {
		layerKey := fmt.Sprintf("%d", node.Layer)
		layerStats[layerKey]++
	}

	topology := &Topology{
		RootDevice: rootDevice,
		Depth:      depth,
		Timestamp:  time.Now().Unix(),
		Nodes:      nodes,
		Edges:      edges,
		Stats: TopologyStats{
			TotalNodes: len(nodes),
			TotalEdges: len(edges),
			Layers:     layerStats,
		},
	}

	return topology, nil
}

func (tb *TopologyBuilder) exploreFromRoot(ctx context.Context, rootDevice string, rootLayer, depth int, visited map[string]bool, nodeMap map[string]*Node, edges *[]Edge) error {
	queue := []struct {
		device string
		level  int
	}{{rootDevice, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.device] || current.level > depth {
			continue
		}

		visited[current.device] = true

		deviceInfo, err := tb.redis.GetDevice(ctx, current.device)
		if err != nil {
			return fmt.Errorf("failed to get device info for %s: %w", current.device, err)
		}
		if deviceInfo == nil {
			continue
		}

		node := &Node{
			Name:     deviceInfo.Name,
			Type:     deviceInfo.Type,
			Hardware: deviceInfo.Hardware,
			Status:   deviceInfo.Status,
			Layer:    deviceInfo.Layer,
			IsRoot:   current.device == rootDevice,
		}
		nodeMap[current.device] = node

		neighbors, err := tb.redis.GetNeighbors(ctx, current.device)
		if err != nil {
			return fmt.Errorf("failed to get neighbors for %s: %w", current.device, err)
		}

		for _, neighbor := range neighbors {
			neighborInfo, err := tb.redis.GetDevice(ctx, neighbor)
			if err != nil {
				continue
			}
			if neighborInfo == nil {
				continue
			}

			shouldInclude := tb.shouldIncludeNeighbor(rootLayer, neighborInfo.Layer, current.level, depth)
			if shouldInclude {
				if !visited[neighbor] {
					queue = append(queue, struct {
						device string
						level  int
					}{neighbor, current.level + 1})
				}

				link, err := tb.redis.GetLink(ctx, current.device, neighbor)
				if err == nil && link != nil {
					edge := Edge{
						Source:     link.Source,
						Target:     link.Target,
						LocalPort:  link.LocalPort,
						RemotePort: link.RemotePort,
						Status:     link.Status,
					}
					*edges = append(*edges, edge)
				}

				reverseLink, err := tb.redis.GetLink(ctx, neighbor, current.device)
				if err == nil && reverseLink != nil {
					found := false
					for _, existingEdge := range *edges {
						if (existingEdge.Source == reverseLink.Source && existingEdge.Target == reverseLink.Target) ||
							(existingEdge.Source == reverseLink.Target && existingEdge.Target == reverseLink.Source) {
							found = true
							break
						}
					}
					if !found {
						edge := Edge{
							Source:     reverseLink.Source,
							Target:     reverseLink.Target,
							LocalPort:  reverseLink.LocalPort,
							RemotePort: reverseLink.RemotePort,
							Status:     reverseLink.Status,
						}
						*edges = append(*edges, edge)
					}
				}
			}
		}
	}

	return nil
}

func (tb *TopologyBuilder) shouldIncludeNeighbor(rootLayer, neighborLayer, currentLevel, maxDepth int) bool {
	if currentLevel >= maxDepth {
		return false
	}

	if neighborLayer < rootLayer && currentLevel <= 2 {
		return true
	}

	if neighborLayer >= rootLayer {
		return true
	}

	return false
}

func (tb *TopologyBuilder) GetDeviceInfo(ctx context.Context, deviceName string) (*DeviceInfo, error) {
	device, err := tb.redis.GetDevice(ctx, deviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return nil, fmt.Errorf("device %s not found", deviceName)
	}

	neighbors, err := tb.redis.GetNeighbors(ctx, deviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get neighbors: %w", err)
	}

	var neighborDetails []NeighborInfo
	for _, neighbor := range neighbors {
		link, err := tb.redis.GetLink(ctx, deviceName, neighbor)
		if err != nil {
			continue
		}

		neighborDevice, err := tb.redis.GetDevice(ctx, neighbor)
		if err != nil {
			continue
		}

		neighborInfo := NeighborInfo{
			Name:       neighbor,
			Type:       neighborDevice.Type,
			Layer:      neighborDevice.Layer,
			LocalPort:  link.LocalPort,
			RemotePort: link.RemotePort,
			Status:     link.Status,
		}
		neighborDetails = append(neighborDetails, neighborInfo)
	}

	deviceInfo := &DeviceInfo{
		Device:    *device,
		Neighbors: neighborDetails,
	}

	return deviceInfo, nil
}

type DeviceInfo struct {
	Device    storage.Device `json:"device"`
	Neighbors []NeighborInfo `json:"neighbors"`
}

type NeighborInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Layer      int    `json:"layer"`
	LocalPort  string `json:"local_port"`
	RemotePort string `json:"remote_port"`
	Status     string `json:"status"`
}