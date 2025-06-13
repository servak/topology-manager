package visualization

import (
	"time"
)

type VisualTopology struct {
	RootDevice string             `json:"root_device"`
	Depth      int                `json:"depth"`
	Timestamp  int64              `json:"timestamp"`
	Nodes      []VisualNode       `json:"nodes"`
	Edges      []VisualEdge       `json:"edges"`
	Groups     []GroupedVisualNode `json:"groups,omitempty"`
	Layout     Layout             `json:"layout"`
	Stats      TopologyStats      `json:"stats"`
}

type VisualNode struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Hardware string    `json:"hardware"`
	Status   string    `json:"status"`
	Layer    int       `json:"layer"`
	IsRoot   bool      `json:"is_root"`
	Position Position  `json:"position"`
	Style    NodeStyle `json:"style"`
}

type VisualEdge struct {
	ID         string    `json:"id"`
	Source     string    `json:"source"`
	Target     string    `json:"target"`
	LocalPort  string    `json:"local_port"`
	RemotePort string    `json:"remote_port"`
	Status     string    `json:"status"`
	Weight     float64   `json:"weight"`
	Style      EdgeStyle `json:"style"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type NodeStyle struct {
	Color       string  `json:"color"`
	Shape       string  `json:"shape"`
	Size        float64 `json:"size"`
	BorderColor string  `json:"border_color"`
	BorderWidth float64 `json:"border_width"`
}

type EdgeStyle struct {
	Color     string  `json:"color"`
	Width     float64 `json:"width"`
	LineStyle string  `json:"line_style"`
}

type Layout struct {
	Type      string                 `json:"type"`
	Options   map[string]interface{} `json:"options"`
	Positions map[string]Position    `json:"positions"`
}

type TopologyStats struct {
	TotalNodes   int            `json:"total_nodes"`
	TotalEdges   int            `json:"total_edges"`
	TotalGroups  int            `json:"total_groups"`
	Layers       map[string]int `json:"layers"`
	Generated    time.Time      `json:"generated"`
}

// GroupedVisualNode represents a group of nodes that are visually collapsed
type GroupedVisualNode struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`          // "group"
	GroupType   string           `json:"group_type"`    // "prefix", "depth", "type"
	Prefix      string           `json:"prefix"`
	Count       int              `json:"count"`
	DeviceIDs   []string         `json:"device_ids"`
	Depth       int              `json:"depth"`
	IsExpanded  bool             `json:"is_expanded"`
	Position    Position         `json:"position"`
	Style       GroupedNodeStyle `json:"style"`
	// グループ内のエッジ情報
	InternalEdgeCount int `json:"internal_edge_count"`
	ExternalEdges     []string `json:"external_edges"` // このグループと接続するエッジのID
}

// GroupedNodeStyle represents the visual style for a grouped node
type GroupedNodeStyle struct {
	Color       string  `json:"color"`
	Shape       string  `json:"shape"`      // "round-rectangle" for groups
	Size        float64 `json:"size"`
	BorderColor string  `json:"border_color"`
	BorderWidth float64 `json:"border_width"`
	Label       string  `json:"label"`
}

// GroupingOptions specifies how nodes should be grouped
type GroupingOptions struct {
	Enabled        bool   `json:"enabled"`
	MinGroupSize   int    `json:"min_group_size"`   // 最小グループサイズ
	MaxDepth       int    `json:"max_depth"`        // この深度より深いノードをグループ化
	GroupByPrefix  bool   `json:"group_by_prefix"`  // 共通プレフィックスでグループ化
	GroupByType    bool   `json:"group_by_type"`    // デバイスタイプでグループ化
	GroupByDepth   bool   `json:"group_by_depth"`   // 深度でグループ化
	PrefixMinLen   int    `json:"prefix_min_len"`   // 最小プレフィックス長
}