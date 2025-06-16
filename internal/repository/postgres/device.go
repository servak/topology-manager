package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/servak/topology-manager/internal/domain/topology"
)

// Device-related repository methods

func (r *postgresRepository) AddDevice(ctx context.Context, device topology.Device) error {
	query := `
		INSERT INTO devices (id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			hardware = EXCLUDED.hardware,
			layer_id = EXCLUDED.layer_id,
			device_type = EXCLUDED.device_type,
			classified_by = EXCLUDED.classified_by,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = EXCLUDED.updated_at
	`

	metadataJSON := "{}"
	if len(device.Metadata) > 0 {
		// TODO: Proper JSON serialization for metadata
		metadataJSON = "{}"
	}

	_, err := r.db.ExecContext(ctx, query,
		device.ID, device.Type, device.Hardware, device.LayerID,
		device.DeviceType, device.ClassifiedBy, metadataJSON, device.LastSeen,
		device.CreatedAt, device.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add device: %w", err)
	}

	return nil
}

func (r *postgresRepository) UpdateDevice(ctx context.Context, device topology.Device) error {
	return r.AddDevice(ctx, device) // Use upsert logic
}

func (r *postgresRepository) GetDevice(ctx context.Context, deviceID string) (*topology.Device, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE id = $1
	`

	var device topology.Device
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&device.ID, &device.Type, &device.Hardware, &device.LayerID,
		&device.DeviceType, &device.ClassifiedBy, &metadataJSON, &device.LastSeen,
		&device.CreatedAt, &device.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// Initialize metadata map
	device.Metadata = make(map[string]string)
	// TODO: Parse JSON metadata

	return &device, nil
}

func (r *postgresRepository) GetDevices(ctx context.Context, opts topology.PaginationOptions) ([]topology.Device, *topology.PaginationResult, error) {
	// Count total devices
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM devices"
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count devices: %w", err)
	}

	// Calculate pagination
	offset := (opts.Page - 1) * opts.PageSize
	totalPages := (totalCount + opts.PageSize - 1) / opts.PageSize

	// Get devices with pagination
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, opts.PageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get devices: %w", err)
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON string

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.LayerID,
			&device.DeviceType, &device.ClassifiedBy, &metadataJSON, &device.LastSeen,
			&device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan device: %w", err)
		}

		device.Metadata = make(map[string]string)
		devices = append(devices, device)
	}

	result := &topology.PaginationResult{
		TotalCount: totalCount,
		TotalPages: totalPages,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		HasNext:    opts.Page < totalPages,
		HasPrev:    opts.Page > 1,
	}

	return devices, result, nil
}

func (r *postgresRepository) SearchDevices(ctx context.Context, query string, limit int) ([]topology.Device, error) {
	searchQuery := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE id ILIKE $1 OR type ILIKE $1 OR hardware ILIKE $1 OR device_type ILIKE $1
		ORDER BY id
		LIMIT $2
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search devices: %w", err)
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON string

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.LayerID,
			&device.DeviceType, &device.ClassifiedBy, &metadataJSON, &device.LastSeen,
			&device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		device.Metadata = make(map[string]string)
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *postgresRepository) RemoveDevice(ctx context.Context, deviceID string) error {
	query := `DELETE FROM devices WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to remove device: %w", err)
	}
	return nil
}

func (r *postgresRepository) FindDevicesByType(ctx context.Context, deviceType string) ([]topology.Device, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE device_type = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to find devices by type: %w", err)
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON string

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.LayerID,
			&device.DeviceType, &device.ClassifiedBy, &metadataJSON, &device.LastSeen,
			&device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		device.Metadata = make(map[string]string)
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *postgresRepository) FindDevicesByHardware(ctx context.Context, hardware string) ([]topology.Device, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE hardware = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, hardware)
	if err != nil {
		return nil, fmt.Errorf("failed to find devices by hardware: %w", err)
	}
	defer rows.Close()

	var devices []topology.Device
	for rows.Next() {
		var device topology.Device
		var metadataJSON string

		err := rows.Scan(
			&device.ID, &device.Type, &device.Hardware, &device.LayerID,
			&device.DeviceType, &device.ClassifiedBy, &metadataJSON, &device.LastSeen,
			&device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		device.Metadata = make(map[string]string)
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *postgresRepository) BulkAddDevices(ctx context.Context, devices []topology.Device) error {
	if len(devices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO devices (id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			hardware = EXCLUDED.hardware,
			layer_id = EXCLUDED.layer_id,
			device_type = EXCLUDED.device_type,
			classified_by = EXCLUDED.classified_by,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, device := range devices {
		metadataJSON := "{}"
		_, err = stmt.ExecContext(ctx,
			device.ID, device.Type, device.Hardware, device.LayerID,
			device.DeviceType, device.ClassifiedBy, metadataJSON, device.LastSeen,
			device.CreatedAt, device.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert device %s: %w", device.ID, err)
		}
	}

	return tx.Commit()
}