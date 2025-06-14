package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/servak/topology-manager/internal/domain/topology"
)

type PostgresRepository struct {
	db *sqlx.DB
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// DSN returns the PostgreSQL connection string
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// Validate checks if the configuration is valid
func (c *PostgresConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("postgres host is required")
	}
	if c.User == "" {
		return fmt.Errorf("postgres user is required")
	}
	if c.Password == "" {
		return fmt.Errorf("postgres password is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("postgres database name is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("postgres port must be between 1 and 65535, got %d", c.Port)
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable" // Default value
	}
	return nil
}

func (r *PostgresRepository) DB() *sqlx.DB {
	return r.db
}

func (r *PostgresRepository) GetDB() *sql.DB {
	return r.db.DB
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) AddDevice(ctx context.Context, device topology.Device) error {
	metadataJSON, err := json.Marshal(device.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO devices (id, type, hardware, instance, location, status, layer, metadata, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			hardware = EXCLUDED.hardware,
			instance = EXCLUDED.instance,
			location = EXCLUDED.location,
			status = EXCLUDED.status,
			layer = EXCLUDED.layer,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP`

	_, err = r.db.ExecContext(ctx, query,
		device.ID, device.Type, device.Hardware, device.Instance,
		device.Location, device.Status, device.Layer,
		metadataJSON, device.LastSeen)

	return err
}

func (r *PostgresRepository) AddLink(ctx context.Context, link topology.Link) error {
	metadataJSON, err := json.Marshal(link.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO links (id, source_id, target_id, source_port, target_port, weight, status, metadata, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (source_id, source_port, target_id, target_port) DO UPDATE SET
			weight = EXCLUDED.weight,
			status = EXCLUDED.status,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP`

	_, err = r.db.ExecContext(ctx, query,
		link.ID, link.SourceID, link.TargetID, link.SourcePort, link.TargetPort,
		link.Weight, link.Status, metadataJSON, link.LastSeen)

	return err
}

func (r *PostgresRepository) UpdateDevice(ctx context.Context, device topology.Device) error {
	metadataJSON, err := json.Marshal(device.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE devices SET
			type = $2, hardware = $3, instance = $4,
			location = $5, status = $6, layer = $7, metadata = $8, last_seen = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		device.ID, device.Type, device.Hardware, device.Instance,
		device.Location, device.Status, device.Layer,
		metadataJSON, device.LastSeen)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("device with ID %s not found", device.ID)
	}

	return nil
}

func (r *PostgresRepository) UpdateLink(ctx context.Context, link topology.Link) error {
	metadataJSON, err := json.Marshal(link.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE links SET
			source_id = $2, target_id = $3, source_port = $4, target_port = $5,
			weight = $6, status = $7, metadata = $8, last_seen = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		link.ID, link.SourceID, link.TargetID, link.SourcePort, link.TargetPort,
		link.Weight, link.Status, metadataJSON, link.LastSeen)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("link with ID %s not found", link.ID)
	}

	return nil
}

func (r *PostgresRepository) RemoveDevice(ctx context.Context, deviceID string) error {
	query := `DELETE FROM devices WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("device with ID %s not found", deviceID)
	}

	return nil
}

func (r *PostgresRepository) RemoveLink(ctx context.Context, linkID string) error {
	query := `DELETE FROM links WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("link with ID %s not found", linkID)
	}

	return nil
}

func (r *PostgresRepository) GetDevice(ctx context.Context, deviceID string) (*topology.Device, error) {
	query := `SELECT id, type, hardware, instance, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE id = $1`

	var device topology.Device
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&device.ID, &device.Type, &device.Hardware, &device.Instance,
		&device.Location, &device.Status, &device.Layer,
		&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &device, nil
}

// SearchDevices searches for devices by ID or IP address with fuzzy matching
func (r *PostgresRepository) SearchDevices(ctx context.Context, query string, limit int) ([]topology.Device, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}

	sqlQuery := `
		SELECT id, type, hardware, instance, location, status, layer, metadata, last_seen, created_at, updated_at 
		FROM devices 
		WHERE 
			id ILIKE $1 OR 
			hardware ILIKE $1
		ORDER BY 
			CASE 
				WHEN id = $2 THEN 1
				WHEN id ILIKE $3 THEN 2
				ELSE 3
			END,
			id
		LIMIT $4`

	searchPattern := "%" + query + "%"
	exactQuery := query
	prefixPattern := query + "%"

	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, exactQuery, prefixPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON []byte

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.Instance,
			&device.Location, &device.Status, &device.Layer,
			&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		devices = append(devices, device)
	}

	return devices, rows.Err()
}

func (r *PostgresRepository) GetLink(ctx context.Context, linkID string) (*topology.Link, error) {
	query := `SELECT id, source_id, target_id, source_port, target_port, weight, status, metadata, last_seen, created_at, updated_at FROM links WHERE id = $1`

	var link topology.Link
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, linkID).Scan(
		&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
		&link.Weight, &link.Status, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(metadataJSON, &link.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &link, nil
}

func (r *PostgresRepository) FindDevicesByType(ctx context.Context, deviceType string) ([]topology.Device, error) {
	query := `SELECT id, type, hardware, instance, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE type = $1`

	rows, err := r.db.QueryContext(ctx, query, deviceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON []byte

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.Instance,
			&device.Location, &device.Status, &device.Layer,
			&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		devices = append(devices, device)
	}

	return devices, rows.Err()
}

func (r *PostgresRepository) FindDevicesByHardware(ctx context.Context, hardware string) ([]topology.Device, error) {
	query := `SELECT id, type, hardware, instance, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE hardware ILIKE $1`

	rows, err := r.db.QueryContext(ctx, query, "%"+hardware+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON []byte

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.Instance,
			&device.Location, &device.Status, &device.Layer,
			&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		devices = append(devices, device)
	}

	return devices, rows.Err()
}

func (r *PostgresRepository) FindDevicesByInstance(ctx context.Context, instance string) ([]topology.Device, error) {
	query := `SELECT id, type, hardware, instance, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE instance = $1`

	rows, err := r.db.QueryContext(ctx, query, instance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON []byte

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.Instance,
			&device.Location, &device.Status, &device.Layer,
			&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		devices = append(devices, device)
	}

	return devices, rows.Err()
}

func (r *PostgresRepository) GetDeviceLinks(ctx context.Context, deviceID string) ([]topology.Link, error) {
	query := `SELECT id, source_id, target_id, source_port, target_port, weight, status, metadata, last_seen, created_at, updated_at FROM links WHERE source_id = $1 OR target_id = $1`

	rows, err := r.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []topology.Link
	for rows.Next() {
		var link topology.Link
		var metadataJSON []byte

		err := rows.Scan(
			&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
			&link.Weight, &link.Status, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &link.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

func (r *PostgresRepository) FindLinksByPort(ctx context.Context, deviceID, port string) ([]topology.Link, error) {
	query := `
		SELECT id, source_id, target_id, source_port, target_port, weight, status, metadata, last_seen, created_at, updated_at 
		FROM links 
		WHERE (source_id = $1 AND source_port = $2) OR (target_id = $1 AND target_port = $2)`

	rows, err := r.db.QueryContext(ctx, query, deviceID, port)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []topology.Link
	for rows.Next() {
		var link topology.Link
		var metadataJSON []byte

		err := rows.Scan(
			&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
			&link.Weight, &link.Status, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &link.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

func (r *PostgresRepository) BulkAddDevices(ctx context.Context, devices []topology.Device) error {
	if len(devices) == 0 {
		return nil
	}

	const batchSize = 1000
	for i := 0; i < len(devices); i += batchSize {
		end := i + batchSize
		if end > len(devices) {
			end = len(devices)
		}

		if err := r.bulkAddDevicesBatch(ctx, devices[i:end]); err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepository) bulkAddDevicesBatch(ctx context.Context, devices []topology.Device) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// COPY形式でバルクインサート
	valueStrings := make([]string, 0, len(devices))
	valueArgs := make([]interface{}, 0, len(devices)*9)

	for i, device := range devices {
		metadataJSON, err := json.Marshal(device.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for device %s: %w", device.ID, err)
		}

		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))

		valueArgs = append(valueArgs,
			device.ID, device.Type, device.Hardware, device.Instance,
			device.Location, device.Status, device.Layer,
			metadataJSON, device.LastSeen)
	}

	query := fmt.Sprintf(`
		INSERT INTO devices (id, type, hardware, instance, location, status, layer, metadata, last_seen)
		VALUES %s
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			hardware = EXCLUDED.hardware,
			instance = EXCLUDED.instance,
			location = EXCLUDED.location,
			status = EXCLUDED.status,
			layer = EXCLUDED.layer,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP`,
		strings.Join(valueStrings, ","))

	_, err = tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to bulk insert devices: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresRepository) BulkAddLinks(ctx context.Context, links []topology.Link) error {
	if len(links) == 0 {
		return nil
	}

	const batchSize = 1000
	for i := 0; i < len(links); i += batchSize {
		end := i + batchSize
		if end > len(links) {
			end = len(links)
		}

		if err := r.bulkAddLinksBatch(ctx, links[i:end]); err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepository) bulkAddLinksBatch(ctx context.Context, links []topology.Link) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// COPY形式でバルクインサート
	valueStrings := make([]string, 0, len(links))
	valueArgs := make([]interface{}, 0, len(links)*9)

	for i, link := range links {
		metadataJSON, err := json.Marshal(link.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for link %s: %w", link.ID, err)
		}

		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))

		valueArgs = append(valueArgs,
			link.ID, link.SourceID, link.TargetID, link.SourcePort, link.TargetPort,
			link.Weight, link.Status, metadataJSON, link.LastSeen)
	}

	query := fmt.Sprintf(`
		INSERT INTO links (id, source_id, target_id, source_port, target_port, weight, status, metadata, last_seen)
		VALUES %s
		ON CONFLICT (source_id, source_port, target_id, target_port) DO UPDATE SET
			weight = EXCLUDED.weight,
			status = EXCLUDED.status,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP`,
		strings.Join(valueStrings, ","))

	_, err = tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to bulk insert links: %w", err)
	}

	return tx.Commit()
}

// 複雑な検索メソッドのスタブ実装（後で拡張）
func (r *PostgresRepository) FindReachableDevices(ctx context.Context, deviceID string, opts topology.ReachabilityOptions) ([]topology.Device, error) {
	if opts.MaxHops <= 0 {
		opts.MaxHops = 5 // デフォルト最大5ホップ
	}
	if opts.Algorithm == "" {
		opts.Algorithm = topology.AlgorithmBFS // デフォルトはBFS
	}

	var query string
	if opts.Algorithm == topology.AlgorithmBFS {
		// BFS: 幅優先探索（レベル順に探索）
		query = `
		WITH RECURSIVE reachability AS (
			-- 起点デバイス
			SELECT 
				d.id, d.type, d.hardware, d.instance,
				d.location, d.status, d.layer,
				d.metadata, d.last_seen, d.created_at, d.updated_at,
				0 as hop_count,
				d.id::text as path
			FROM devices d 
			WHERE d.id = $1
			
			UNION ALL
			
			-- 隣接デバイス（BFS: レベル順）
			SELECT 
				d.id, d.type, d.hardware, d.instance,
				d.location, d.status, d.layer,
				d.metadata, d.last_seen, d.created_at, d.updated_at,
				r.hop_count + 1,
				r.path || ',' || d.id::text
			FROM devices d
			INNER JOIN links l ON (d.id = l.source_id OR d.id = l.target_id)
			INNER JOIN reachability r ON (
				(l.source_id = r.id AND d.id = l.target_id) OR
				(l.target_id = r.id AND d.id = l.source_id)
			)
			WHERE r.hop_count < $2
			  AND position(',' || d.id::text || ',' in ',' || r.path || ',') = 0
		)
		SELECT DISTINCT 
			id, type, hardware, instance, location,
			status, layer, metadata, last_seen, created_at, updated_at, hop_count
		FROM reachability
		WHERE hop_count > 0  -- 起点デバイス自身は除外
		ORDER BY hop_count, id`
	} else {
		// DFS: 深度優先探索（深い経路を優先）
		query = `
		WITH RECURSIVE reachability AS (
			-- 起点デバイス
			SELECT 
				d.id, d.type, d.hardware, d.instance,
				d.location, d.status, d.layer,
				d.metadata, d.last_seen, d.created_at, d.updated_at,
				0 as hop_count,
				d.id::text as path,
				1 as search_order
			FROM devices d 
			WHERE d.id = $1
			
			UNION ALL
			
			-- 隣接デバイス（DFS: 深度優先）
			SELECT 
				d.id, d.type, d.hardware, d.instance,
				d.location, d.status, d.layer,
				d.metadata, d.last_seen, d.created_at, d.updated_at,
				r.hop_count + 1,
				r.path || ',' || d.id::text,
				r.search_order + 1
			FROM devices d
			INNER JOIN links l ON (d.id = l.source_id OR d.id = l.target_id)
			INNER JOIN reachability r ON (
				(l.source_id = r.id AND d.id = l.target_id) OR
				(l.target_id = r.id AND d.id = l.source_id)
			)
			WHERE r.hop_count < $2
			  AND position(',' || d.id::text || ',' in ',' || r.path || ',') = 0
		)
		SELECT DISTINCT 
			id, type, hardware, instance, location,
			status, layer, metadata, last_seen, created_at, updated_at, search_order
		FROM reachability
		WHERE hop_count > 0  -- 起点デバイス自身は除外
		ORDER BY search_order DESC, id` // DFS順序
	}

	rows, err := r.db.QueryContext(ctx, query, deviceID, opts.MaxHops)
	if err != nil {
		return nil, fmt.Errorf("failed to query reachable devices: %w", err)
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON []byte
		var orderField interface{} // hop_count or search_order

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.Instance,
			&device.Location, &device.Status, &device.Layer,
			&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt, &orderField)

		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		devices = append(devices, device)
	}

	return devices, rows.Err()
}

func (r *PostgresRepository) ExtractSubTopology(ctx context.Context, deviceID string, opts topology.SubTopologyOptions) ([]topology.Device, []topology.Link, error) {
	if opts.Radius <= 0 {
		opts.Radius = 3
	}

	// より効率的な非再帰アプローチ - レベル毎に実行
	// 現実的なSeedデータなので制限を撤廃

	// CTEを使った効率的な階層検索クエリ（改良版）
	query := `
	WITH RECURSIVE topology_traversal AS (
		-- 起点デバイス
		SELECT 
			d.id, d.type, d.hardware, d.instance, 
			d.location, d.status, d.layer, 
			d.metadata, d.last_seen, d.created_at, d.updated_at,
			0 as level,
			d.id::text as path -- 循環検出用（文字列として保存）
		FROM devices d 
		WHERE d.id = $1
		
		UNION ALL
		
		-- 隣接デバイス（再帰部分）
		SELECT 
			d.id, d.type, d.hardware, d.instance,
			d.location, d.status, d.layer,
			d.metadata, d.last_seen, d.created_at, d.updated_at,
			tt.level + 1,
			tt.path || ',' || d.id::text
		FROM devices d
		INNER JOIN links l ON (d.id = l.source_id OR d.id = l.target_id)
		INNER JOIN topology_traversal tt ON (
			(l.source_id = tt.id AND d.id = l.target_id) OR
			(l.target_id = tt.id AND d.id = l.source_id)
		)
		WHERE tt.level < $2 
		  AND position(',' || d.id::text || ',' in ',' || tt.path || ',') = 0 -- 循環回避
	)
	SELECT DISTINCT 
		id, type, hardware, instance, location, 
		status, layer, metadata, last_seen, created_at, updated_at
	FROM topology_traversal
	ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, deviceID, opts.Radius)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query sub-topology devices: %w", err)
	}
	defer rows.Close()

	var devices []topology.Device
	deviceNames := make([]string, 0)

	for rows.Next() {
		var device topology.Device
		var metadataJSON []byte

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.Instance,
			&device.Location, &device.Status, &device.Layer,
			&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		devices = append(devices, device)
		deviceNames = append(deviceNames, device.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error during device scan: %w", err)
	}

	// デバイス間のリンクを一括取得
	if len(deviceNames) == 0 {
		return devices, []topology.Link{}, nil
	}

	// プレースホルダーを動的に生成
	placeholders1 := make([]string, len(deviceNames))
	placeholders2 := make([]string, len(deviceNames))
	linkArgs := make([]interface{}, len(deviceNames)*2)

	for i, name := range deviceNames {
		placeholders1[i] = fmt.Sprintf("$%d", i+1)
		placeholders2[i] = fmt.Sprintf("$%d", i+1+len(deviceNames))
		linkArgs[i] = name
		linkArgs[i+len(deviceNames)] = name
	}

	linkQuery := fmt.Sprintf(`
		SELECT id, source_id, target_id, source_port, target_port, weight, status, metadata, last_seen, created_at, updated_at
		FROM links 
		WHERE source_id IN (%s) AND target_id IN (%s)`,
		strings.Join(placeholders1, ","),
		strings.Join(placeholders2, ","))

	linkRows, err := r.db.QueryContext(ctx, linkQuery, linkArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query sub-topology links: %w", err)
	}
	defer linkRows.Close()

	var links []topology.Link
	for linkRows.Next() {
		var link topology.Link
		var metadataJSON []byte

		err := linkRows.Scan(
			&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
			&link.Weight, &link.Status, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan link: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &link.Metadata); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal link metadata: %w", err)
		}

		links = append(links, link)
	}

	return devices, links, linkRows.Err()
}

func (r *PostgresRepository) FindShortestPath(ctx context.Context, fromID, toID string, opts topology.PathOptions) (*topology.Path, error) {
	if opts.Algorithm == "" {
		opts.Algorithm = topology.PathAlgorithmDijkstra
	}

	// Dijkstraアルゴリズム（重み付き最短パス）
	query := `
	WITH RECURSIVE dijkstra AS (
		-- 起点デバイス
		SELECT 
			d.id, d.type, d.hardware, d.instance,
			d.location, d.status, d.layer,
			d.metadata, d.last_seen, d.created_at, d.updated_at,
			0.0 as total_cost,
			0 as hop_count,
			d.id::text as path_devices,
			''::text as path_links
		FROM devices d 
		WHERE d.id = $1
		
		UNION ALL
		
		-- 隣接デバイス（Dijkstra）
		SELECT 
			d.id, d.type, d.hardware, d.instance,
			d.location, d.status, d.layer,
			d.metadata, d.last_seen, d.created_at, d.updated_at,
			dij.total_cost + l.weight as total_cost,
			dij.hop_count + 1,
			dij.path_devices || ',' || d.id::text,
			CASE 
				WHEN dij.path_links = '' THEN l.id
				ELSE dij.path_links || ',' || l.id
			END
		FROM devices d
		INNER JOIN links l ON (d.id = l.source_id OR d.id = l.target_id)
		INNER JOIN dijkstra dij ON (
			(l.source_id = dij.id AND d.id = l.target_id) OR
			(l.target_id = dij.id AND d.id = l.source_id)
		)
		WHERE position(',' || d.id::text || ',' in ',' || dij.path_devices || ',') = 0
		  AND dij.hop_count < 6  -- 最大6ホップ制限（現実的な範囲）
	)
	SELECT 
		path_devices,
		path_links,
		total_cost,
		hop_count
	FROM dijkstra
	WHERE id = $2
	ORDER BY total_cost ASC, hop_count ASC
	LIMIT 1`

	var pathDevices, pathLinks string
	var totalCost float64
	var hopCount int

	err := r.db.QueryRowContext(ctx, query, fromID, toID).Scan(
		&pathDevices, &pathLinks, &totalCost, &hopCount)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no path found from %s to %s", fromID, toID)
		}
		return nil, fmt.Errorf("failed to find shortest path: %w", err)
	}

	// パスの詳細情報を取得
	deviceIDs := strings.Split(pathDevices, ",")
	linkIDs := []string{}
	if pathLinks != "" {
		linkIDs = strings.Split(pathLinks, ",")
	}

	// デバイス情報を取得
	devices := make([]topology.Device, 0, len(deviceIDs))
	for _, deviceID := range deviceIDs {
		device, err := r.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get device %s: %w", deviceID, err)
		}
		if device != nil {
			devices = append(devices, *device)
		}
	}

	// リンク情報を取得
	links := make([]topology.Link, 0, len(linkIDs))
	for _, linkID := range linkIDs {
		link, err := r.GetLink(ctx, linkID)
		if err != nil {
			return nil, fmt.Errorf("failed to get link %s: %w", linkID, err)
		}
		if link != nil {
			links = append(links, *link)
		}
	}

	return &topology.Path{
		Devices:   devices,
		Links:     links,
		TotalCost: totalCost,
		HopCount:  hopCount,
	}, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) GetDevices(ctx context.Context, opts topology.PaginationOptions) ([]topology.Device, *topology.PaginationResult, error) {
	// デフォルト値設定
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100 // 最大100件
	}
	if opts.OrderBy == "" {
		opts.OrderBy = "name"
	}
	if opts.SortDir == "" {
		opts.SortDir = "ASC"
	}

	// WHERE句とパラメータの構築
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if opts.Type != "" {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, opts.Type)
		argIndex++
	}
	if opts.Hardware != "" {
		whereClause += fmt.Sprintf(" AND hardware ILIKE $%d", argIndex)
		args = append(args, "%"+opts.Hardware+"%")
		argIndex++
	}
	if opts.Instance != "" {
		whereClause += fmt.Sprintf(" AND instance = $%d", argIndex)
		args = append(args, opts.Instance)
		argIndex++
	}

	// 総件数取得
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM devices %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count devices: %w", err)
	}

	// ページ計算
	totalPages := (totalCount + opts.PageSize - 1) / opts.PageSize
	offset := (opts.Page - 1) * opts.PageSize

	// ソート方向の検証
	sortDir := "ASC"
	if opts.SortDir == "DESC" || opts.SortDir == "desc" {
		sortDir = "DESC"
	}

	// カラム名の検証（SQLインジェクション対策）
	allowedColumns := map[string]bool{
		"id": true, "type": true, "hardware": true, "instance": true,
		"layer": true, "status": true, "created_at": true, "updated_at": true,
	}
	orderBy := "id"
	if allowedColumns[opts.OrderBy] {
		orderBy = opts.OrderBy
	}

	// データ取得クエリ
	query := fmt.Sprintf(`
		SELECT id, type, hardware, instance, location, status, layer, metadata, last_seen, created_at, updated_at 
		FROM devices %s 
		ORDER BY %s %s 
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, sortDir, argIndex, argIndex+1)

	args = append(args, opts.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON []byte

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.Instance,
			&device.Location, &device.Status, &device.Layer,
			&metadataJSON, &device.LastSeen, &device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	// ページング結果
	result := &topology.PaginationResult{
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    opts.Page < totalPages,
		HasPrev:    opts.Page > 1,
	}

	return devices, result, nil
}

func (r *PostgresRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
