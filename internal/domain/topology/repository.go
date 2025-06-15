package topology

import (
	"context"
)

type Repository interface {
	// 単体取得（可視化API使用中）
	GetDevice(ctx context.Context, deviceID string) (*Device, error)

	// デバイス検索（フロントエンド検索機能使用中）
	SearchDevices(ctx context.Context, query string, limit int) ([]Device, error)
	
	// デバイス一覧取得（分類サービス使用中）
	GetDevices(ctx context.Context, opts PaginationOptions) ([]Device, *PaginationResult, error)

	// 更新操作（Worker使用中）
	UpdateDevice(ctx context.Context, device Device) error

	// トポロジー検索（API使用中）
	FindReachableDevices(ctx context.Context, deviceID string, opts ReachabilityOptions) ([]Device, error)
	FindShortestPath(ctx context.Context, fromID, toID string, opts PathOptions) (*Path, error)
	ExtractSubTopology(ctx context.Context, deviceID string, opts SubTopologyOptions) ([]Device, []Link, error)

	// リンク検索（可視化API使用中）
	GetDeviceLinks(ctx context.Context, deviceID string) ([]Link, error)

	// バルク操作（seedDataコマンド使用中）
	BulkAddDevices(ctx context.Context, devices []Device) error
	BulkAddLinks(ctx context.Context, links []Link) error

	// 管理操作
	Close() error
	Health(ctx context.Context) error
}
