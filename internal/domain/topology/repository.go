package topology

import (
	"context"
)

type Repository interface {
	// 基本操作
	AddDevice(ctx context.Context, device Device) error
	AddLink(ctx context.Context, link Link) error
	UpdateDevice(ctx context.Context, device Device) error
	UpdateLink(ctx context.Context, link Link) error
	RemoveDevice(ctx context.Context, deviceID string) error
	RemoveLink(ctx context.Context, linkID string) error

	// 単体取得
	GetDevice(ctx context.Context, deviceID string) (*Device, error)
	GetLink(ctx context.Context, linkID string) (*Link, error)

	// 検索操作
	FindReachableDevices(ctx context.Context, deviceID string, opts ReachabilityOptions) ([]Device, error)
	ExtractSubTopology(ctx context.Context, deviceID string, opts SubTopologyOptions) ([]Device, []Link, error)
	FindShortestPath(ctx context.Context, fromID, toID string, opts PathOptions) (*Path, error)

	// 一覧取得（ページング対応）
	GetDevices(ctx context.Context, opts PaginationOptions) ([]Device, *PaginationResult, error)
	
	// フィルタリング
	FindDevicesByType(ctx context.Context, deviceType string) ([]Device, error)
	FindDevicesByHardware(ctx context.Context, hardware string) ([]Device, error)
	FindDevicesByInstance(ctx context.Context, instance string) ([]Device, error)

	// リンク検索
	GetDeviceLinks(ctx context.Context, deviceID string) ([]Link, error)
	FindLinksByPort(ctx context.Context, deviceID, port string) ([]Link, error)

	// バルク操作
	BulkAddDevices(ctx context.Context, devices []Device) error
	BulkAddLinks(ctx context.Context, links []Link) error

	// 管理操作
	Close() error
	Health(ctx context.Context) error
}