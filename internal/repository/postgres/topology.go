package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/servak/topology-manager/internal/domain/topology"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func (r *PostgresRepository) DB() *sqlx.DB {
	return r.db
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
		INSERT INTO devices (id, name, type, hardware, instance, ip_address, location, status, layer, metadata, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (name) DO UPDATE SET
			type = EXCLUDED.type,
			hardware = EXCLUDED.hardware,
			instance = EXCLUDED.instance,
			ip_address = EXCLUDED.ip_address,
			location = EXCLUDED.location,
			status = EXCLUDED.status,
			layer = EXCLUDED.layer,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP`

	_, err = r.db.ExecContext(ctx, query,
		device.ID, device.Name, device.Type, device.Hardware, device.Instance,
		device.IPAddress, device.Location, device.Status, device.Layer,
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
			name = $2, type = $3, hardware = $4, instance = $5, ip_address = $6,
			location = $7, status = $8, layer = $9, metadata = $10, last_seen = $11,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		device.ID, device.Name, device.Type, device.Hardware, device.Instance,
		device.IPAddress, device.Location, device.Status, device.Layer,
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
	query := `SELECT id, name, type, hardware, instance, ip_address, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE id = $1`
	
	var device topology.Device
	var metadataJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&device.ID, &device.Name, &device.Type, &device.Hardware, &device.Instance,
		&device.IPAddress, &device.Location, &device.Status, &device.Layer,
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
	query := `SELECT id, name, type, hardware, instance, ip_address, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE type = $1`
	
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
			&device.ID, &device.Name, &device.Type, &device.Hardware, &device.Instance,
			&device.IPAddress, &device.Location, &device.Status, &device.Layer,
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
	query := `SELECT id, name, type, hardware, instance, ip_address, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE hardware ILIKE $1`
	
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
			&device.ID, &device.Name, &device.Type, &device.Hardware, &device.Instance,
			&device.IPAddress, &device.Location, &device.Status, &device.Layer,
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
	query := `SELECT id, name, type, hardware, instance, ip_address, location, status, layer, metadata, last_seen, created_at, updated_at FROM devices WHERE instance = $1`
	
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
			&device.ID, &device.Name, &device.Type, &device.Hardware, &device.Instance,
			&device.IPAddress, &device.Location, &device.Status, &device.Layer,
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

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO devices (id, name, type, hardware, instance, ip_address, location, status, layer, metadata, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (name) DO UPDATE SET
			type = EXCLUDED.type,
			hardware = EXCLUDED.hardware,
			instance = EXCLUDED.instance,
			ip_address = EXCLUDED.ip_address,
			location = EXCLUDED.location,
			status = EXCLUDED.status,
			layer = EXCLUDED.layer,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP`)
	
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, device := range devices {
		metadataJSON, err := json.Marshal(device.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for device %s: %w", device.ID, err)
		}

		_, err = stmt.ExecContext(ctx,
			device.ID, device.Name, device.Type, device.Hardware, device.Instance,
			device.IPAddress, device.Location, device.Status, device.Layer,
			metadataJSON, device.LastSeen)
		
		if err != nil {
			return fmt.Errorf("failed to insert device %s: %w", device.ID, err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) BulkAddLinks(ctx context.Context, links []topology.Link) error {
	if len(links) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO links (id, source_id, target_id, source_port, target_port, weight, status, metadata, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (source_id, source_port, target_id, target_port) DO UPDATE SET
			weight = EXCLUDED.weight,
			status = EXCLUDED.status,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP`)
	
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, link := range links {
		metadataJSON, err := json.Marshal(link.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for link %s: %w", link.ID, err)
		}

		_, err = stmt.ExecContext(ctx,
			link.ID, link.SourceID, link.TargetID, link.SourcePort, link.TargetPort,
			link.Weight, link.Status, metadataJSON, link.LastSeen)
		
		if err != nil {
			return fmt.Errorf("failed to insert link %s: %w", link.ID, err)
		}
	}

	return tx.Commit()
}

// 複雑な検索メソッドのスタブ実装（後で拡張）
func (r *PostgresRepository) FindReachableDevices(ctx context.Context, deviceID string, opts topology.ReachabilityOptions) ([]topology.Device, error) {
	// BFS/DFS実装は後で追加
	return nil, fmt.Errorf("not implemented yet")
}

func (r *PostgresRepository) ExtractSubTopology(ctx context.Context, deviceID string, opts topology.SubTopologyOptions) ([]topology.Device, []topology.Link, error) {
	// 実装は後で追加
	return nil, nil, fmt.Errorf("not implemented yet")
}

func (r *PostgresRepository) FindShortestPath(ctx context.Context, fromID, toID string, opts topology.PathOptions) (*topology.Path, error) {
	// Dijkstra実装は後で追加
	return nil, fmt.Errorf("not implemented yet")
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
		"name": true, "type": true, "hardware": true, "instance": true,
		"layer": true, "status": true, "created_at": true, "updated_at": true,
	}
	orderBy := "name"
	if allowedColumns[opts.OrderBy] {
		orderBy = opts.OrderBy
	}

	// データ取得クエリ
	query := fmt.Sprintf(`
		SELECT id, name, type, hardware, instance, ip_address, location, status, layer, metadata, last_seen, created_at, updated_at 
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
			&device.ID, &device.Name, &device.Type, &device.Hardware, &device.Instance,
			&device.IPAddress, &device.Location, &device.Status, &device.Layer,
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