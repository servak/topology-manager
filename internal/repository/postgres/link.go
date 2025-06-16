package postgres

import (
	"context"
	"fmt"

	"github.com/servak/topology-manager/internal/domain/topology"
)

// Link-related repository methods

func (r *postgresRepository) AddLink(ctx context.Context, link topology.Link) error {
	query := `
		INSERT INTO links (id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			source_id = EXCLUDED.source_id,
			target_id = EXCLUDED.target_id,
			source_port = EXCLUDED.source_port,
			target_port = EXCLUDED.target_port,
			weight = EXCLUDED.weight,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = EXCLUDED.updated_at
	`

	metadataJSON := "{}"
	if len(link.Metadata) > 0 {
		// TODO: Proper JSON serialization for metadata
		metadataJSON = "{}"
	}

	_, err := r.db.ExecContext(ctx, query,
		link.ID, link.SourceID, link.TargetID, link.SourcePort, link.TargetPort,
		link.Weight, metadataJSON, link.LastSeen, link.CreatedAt, link.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add link: %w", err)
	}

	return nil
}

func (r *postgresRepository) UpdateLink(ctx context.Context, link topology.Link) error {
	return r.AddLink(ctx, link) // Use upsert logic
}

func (r *postgresRepository) GetLink(ctx context.Context, linkID string) (*topology.Link, error) {
	query := `
		SELECT id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at
		FROM links 
		WHERE id = $1
	`

	var link topology.Link
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, query, linkID).Scan(
		&link.ID, &link.SourceID, &link.TargetID, &link.SourcePort, &link.TargetPort,
		&link.Weight, &metadataJSON, &link.LastSeen, &link.CreatedAt, &link.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	// Initialize metadata map
	link.Metadata = make(map[string]string)
	// TODO: Parse JSON metadata

	return &link, nil
}

func (r *postgresRepository) RemoveLink(ctx context.Context, linkID string) error {
	query := `DELETE FROM links WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("failed to remove link: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetDeviceLinks(ctx context.Context, deviceID string) ([]topology.Link, error) {
	query := `
		SELECT id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at
		FROM links 
		WHERE source_id = $1 OR target_id = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, deviceID)
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

		link.Metadata = make(map[string]string)
		links = append(links, link)
	}

	return links, nil
}

func (r *postgresRepository) FindLinksByPort(ctx context.Context, deviceID, port string) ([]topology.Link, error) {
	query := `
		SELECT id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at
		FROM links 
		WHERE (source_id = $1 AND source_port = $2) OR (target_id = $1 AND target_port = $2)
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, deviceID, port)
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

		link.Metadata = make(map[string]string)
		links = append(links, link)
	}

	return links, nil
}

func (r *postgresRepository) BulkAddLinks(ctx context.Context, links []topology.Link) error {
	if len(links) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO links (id, source_id, target_id, source_port, target_port, weight, metadata, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			source_id = EXCLUDED.source_id,
			target_id = EXCLUDED.target_id,
			source_port = EXCLUDED.source_port,
			target_port = EXCLUDED.target_port,
			weight = EXCLUDED.weight,
			metadata = EXCLUDED.metadata,
			last_seen = EXCLUDED.last_seen,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, link := range links {
		metadataJSON := "{}"
		_, err = stmt.ExecContext(ctx,
			link.ID, link.SourceID, link.TargetID, link.SourcePort, link.TargetPort,
			link.Weight, metadataJSON, link.LastSeen, link.CreatedAt, link.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert link %s: %w", link.ID, err)
		}
	}

	return tx.Commit()
}