package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/servak/topology-manager/internal/domain/topology"
)

// Device-related repository methods

func (r *sqliteRepository) AddDevice(ctx context.Context, device topology.Device) error {
	query := `
		INSERT OR REPLACE INTO devices (id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metadataJSON, err := json.Marshal(device.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		device.ID, device.Type, device.Hardware, device.LayerID,
		device.DeviceType, device.ClassifiedBy, string(metadataJSON), device.LastSeen,
		device.CreatedAt, device.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add device: %w", err)
	}

	return nil
}

func (r *sqliteRepository) UpdateDevice(ctx context.Context, device topology.Device) error {
	return r.AddDevice(ctx, device) // Use upsert logic
}

func (r *sqliteRepository) GetDevice(ctx context.Context, deviceID string) (*topology.Device, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE id = ?
	`

	var device topology.Device
	var metadataJSON string

	err := r.db.QueryRowxContext(ctx, query, deviceID).Scan(
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

	// Parse metadata JSON
	if err := json.Unmarshal([]byte(metadataJSON), &device.Metadata); err != nil {
		device.Metadata = make(map[string]string)
	}

	return &device, nil
}

func (r *sqliteRepository) GetDevices(ctx context.Context, opts topology.PaginationOptions) ([]topology.Device, *topology.PaginationResult, error) {
	// Count total devices
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM devices"
	err := r.db.QueryRowxContext(ctx, countQuery).Scan(&totalCount)
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
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryxContext(ctx, query, opts.PageSize, offset)
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

		// Parse metadata JSON
		if err := json.Unmarshal([]byte(metadataJSON), &device.Metadata); err != nil {
			device.Metadata = make(map[string]string)
		}
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

func (r *sqliteRepository) SearchDevices(ctx context.Context, query string, limit int) ([]topology.Device, error) {
	searchQuery := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE id LIKE ? OR type LIKE ? OR hardware LIKE ? OR device_type LIKE ?
		ORDER BY id
		LIMIT ?
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryxContext(ctx, searchQuery, searchPattern, searchPattern, searchPattern, searchPattern, limit)
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

		// Parse metadata JSON
		if err := json.Unmarshal([]byte(metadataJSON), &device.Metadata); err != nil {
			device.Metadata = make(map[string]string)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *sqliteRepository) RemoveDevice(ctx context.Context, deviceID string) error {
	query := `DELETE FROM devices WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to remove device: %w", err)
	}
	return nil
}

func (r *sqliteRepository) FindDevicesByType(ctx context.Context, deviceType string) ([]topology.Device, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE device_type = ?
		ORDER BY id
	`

	rows, err := r.db.QueryxContext(ctx, query, deviceType)
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

		// Parse metadata JSON
		if err := json.Unmarshal([]byte(metadataJSON), &device.Metadata); err != nil {
			device.Metadata = make(map[string]string)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *sqliteRepository) FindDevicesByHardware(ctx context.Context, hardware string) ([]topology.Device, error) {
	query := `
		SELECT id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at
		FROM devices 
		WHERE hardware = ?
		ORDER BY id
	`

	rows, err := r.db.QueryxContext(ctx, query, hardware)
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

		// Parse metadata JSON
		if err := json.Unmarshal([]byte(metadataJSON), &device.Metadata); err != nil {
			device.Metadata = make(map[string]string)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *sqliteRepository) BulkAddDevices(ctx context.Context, devices []topology.Device) error {
	if len(devices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT OR REPLACE INTO devices (id, type, hardware, layer_id, device_type, classified_by, metadata, last_seen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
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
			device.ID, device.Type, device.Hardware, device.LayerID,
			device.DeviceType, device.ClassifiedBy, string(metadataJSON), device.LastSeen,
			device.CreatedAt, device.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert device %s: %w", device.ID, err)
		}
	}

	return tx.Commit()
}