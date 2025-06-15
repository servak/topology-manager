package topology

import (
	"time"
)

type Device struct {
	ID           string            `json:"id" db:"id"`
	Type         string            `json:"type" db:"type"`
	Hardware     string            `json:"hardware" db:"hardware"`
	LayerID      *int              `json:"layer_id" db:"layer_id"`        // NULL許可
	DeviceType   string            `json:"device_type" db:"device_type"`
	ClassifiedBy string            `json:"classified_by" db:"classified_by"`
	Metadata     map[string]string `json:"metadata" db:"metadata"`
	LastSeen     time.Time         `json:"last_seen" db:"last_seen"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
}

type Link struct {
	ID         string            `json:"id" db:"id"`
	SourceID   string            `json:"source_id" db:"source_id"`
	TargetID   string            `json:"target_id" db:"target_id"`
	SourcePort string            `json:"source_port" db:"source_port"`
	TargetPort string            `json:"target_port" db:"target_port"`
	Weight     float64           `json:"weight" db:"weight"`
	Metadata   map[string]string `json:"metadata" db:"metadata"`
	LastSeen   time.Time         `json:"last_seen" db:"last_seen"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
}

type Path struct {
	Devices   []Device `json:"devices"`
	Links     []Link   `json:"links"`
	TotalCost float64  `json:"total_cost"`
	HopCount  int      `json:"hop_count"`
}

type SearchAlgorithm string

const (
	AlgorithmBFS SearchAlgorithm = "bfs"
	AlgorithmDFS SearchAlgorithm = "dfs"
)

type PathAlgorithm string

const (
	PathAlgorithmDijkstra  PathAlgorithm = "dijkstra"
	PathAlgorithmKShortest PathAlgorithm = "k_shortest"
)

type ReachabilityOptions struct {
	MaxHops   int             `json:"max_hops"`
	Algorithm SearchAlgorithm `json:"algorithm"`
}

type SubTopologyOptions struct {
	Radius int `json:"radius"`
}

type PathOptions struct {
	Algorithm PathAlgorithm `json:"algorithm"`
}

type PaginationOptions struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	OrderBy  string `json:"order_by"`
	SortDir  string `json:"sort_dir"`
	Type     string `json:"type,omitempty"`
	Hardware string `json:"hardware,omitempty"`
}

type PaginationResult struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}