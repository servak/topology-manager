package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/servak/topology-manager/internal/domain/topology"
)

// Link-related repository methods

func (r *sqliteRepository) AddLink(ctx context.Context, link topology.Link) error {
	query := `
		INSERT OR REPLACE INTO links (id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metadataJSON, err := json.Marshal(link.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		link.ID, link.SourceID, link.TargetID, link.SourcePort, link.TargetPort,
		link.Weight, string(metadataJSON), link.LastSeen, link.CreatedAt, link.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add link: %w", err)
	}

	return nil
}

func (r *sqliteRepository) UpdateLink(ctx context.Context, link topology.Link) error {
	return r.AddLink(ctx, link) // Use upsert logic
}

func (r *sqliteRepository) GetLink(ctx context.Context, linkID string) (*topology.Link, error) {
	query := `
		SELECT id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at
		FROM links 
		WHERE id = ?
	`

	var link topology.Link
	var metadataJSON string

	err := r.db.QueryRowxContext(ctx, query, linkID).Scan(
		&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
		&link.Weight, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	// Parse metadata JSON
	if err := json.Unmarshal([]byte(metadataJSON), &link.Metadata); err != nil {
		link.Metadata = make(map[string]string)
	}

	return &link, nil
}

func (r *sqliteRepository) RemoveLink(ctx context.Context, linkID string) error {
	query := `DELETE FROM links WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("failed to remove link: %w", err)
	}
	return nil
}

func (r *sqliteRepository) GetDeviceLinks(ctx context.Context, deviceID string) ([]topology.Link, error) {
	query := `
		SELECT id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at
		FROM links 
		WHERE source_id = ? OR target_id = ?
		ORDER BY id
	`

	rows, err := r.db.QueryxContext(ctx, query, deviceID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device links: %w", err)
	}
	defer rows.Close()

	var links []topology.Link
	for rows.Next() {
		var link topology.Link
		var metadataJSON string

		err := rows.Scan(
			&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
			&link.Weight, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}

		// Parse metadata JSON
		if err := json.Unmarshal([]byte(metadataJSON), &link.Metadata); err != nil {
			link.Metadata = make(map[string]string)
		}
		links = append(links, link)
	}

	return links, nil
}

func (r *sqliteRepository) FindLinksByPort(ctx context.Context, deviceID, port string) ([]topology.Link, error) {
	query := `
		SELECT id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at
		FROM links 
		WHERE (source_id = ? AND source_port = ?) OR (target_id = ? AND target_port = ?)
		ORDER BY id
	`

	rows, err := r.db.QueryxContext(ctx, query, deviceID, port, deviceID, port)
	if err != nil {
		return nil, fmt.Errorf("failed to find links by port: %w", err)
	}
	defer rows.Close()

	var links []topology.Link
	for rows.Next() {
		var link topology.Link
		var metadataJSON string

		err := rows.Scan(
			&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
			&link.Weight, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}

		// Parse metadata JSON
		if err := json.Unmarshal([]byte(metadataJSON), &link.Metadata); err != nil {
			link.Metadata = make(map[string]string)
		}
		links = append(links, link)
	}

	return links, nil
}

func (r *sqliteRepository) BulkAddLinks(ctx context.Context, links []topology.Link) error {
	if len(links) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT OR REPLACE INTO links (id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
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
			link.Weight, string(metadataJSON), link.LastSeen, link.CreatedAt, link.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert link %s: %w", link.ID, err)
		}
	}

	return tx.Commit()
}