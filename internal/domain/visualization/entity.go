package visualization

import (
	"time"
)

type VisualTopology struct {
	RootDevice string      `json:"root_device"`
	Depth      int         `json:"depth"`
	Timestamp  int64       `json:"timestamp"`
	Nodes      []VisualNode `json:"nodes"`
	Edges      []VisualEdge `json:"edges"`
	Layout     Layout      `json:"layout"`
	Stats      TopologyStats `json:"stats"`
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
	TotalNodes int            `json:"total_nodes"`
	TotalEdges int            `json:"total_edges"`
	Layers     map[string]int `json:"layers"`
	Generated  time.Time      `json:"generated"`
}