package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/servak/topology-manager/internal/domain/topology"
)

// Neo4jRepository implements the topology.Repository interface using Neo4j
type Neo4jRepository struct {
	driver neo4j.DriverWithContext
	config *Neo4jConfig
}

// NewNeo4jRepository creates a new Neo4j repository instance
func NewNeo4jRepository(config *Neo4jConfig) (*Neo4jRepository, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Neo4j configuration: %w", err)
	}

	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		neo4j.BasicAuth(config.Username, config.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Test connection
	ctx := context.Background()
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	return &Neo4jRepository{
		driver: driver,
		config: config,
	}, nil
}

// Close closes the Neo4j driver connection
func (r *Neo4jRepository) Close() error {
	return r.driver.Close(context.Background())
}

// AddDevice adds a device to the Neo4j graph
func (r *Neo4jRepository) AddDevice(ctx context.Context, device topology.Device) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	query := `
		CREATE (d:Device {
			id: $id,
			type: $type,
			hardware: $hardware,
			instance: $instance,
			ip_address: $ip_address,
			location: $location,
			status: $status,
			layer: $layer,
			metadata: $metadata,
			last_seen: datetime($last_seen),
			created_at: datetime($created_at),
			updated_at: datetime($updated_at)
		})
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{
			"id":         device.ID,
			"type":       device.Type,
			"hardware":   device.Hardware,
			"instance":   device.Instance,
			"location":   device.Location,
			"status":     device.Status,
			"layer":      device.Layer,
			"metadata":   device.Metadata,
			"last_seen":  device.LastSeen.Format("2006-01-02T15:04:05Z"),
			"created_at": device.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at": device.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	})

	return err
}

// AddLink adds a link between devices in the Neo4j graph
func (r *Neo4jRepository) AddLink(ctx context.Context, link topology.Link) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (source:Device {id: $source_id})
		MATCH (target:Device {id: $target_id})
		CREATE (source)-[l:CONNECTED {
			id: $id,
			source_port: $source_port,
			target_port: $target_port,
			weight: $weight,
			status: $status,
			metadata: $metadata,
			last_seen: datetime($last_seen),
			created_at: datetime($created_at),
			updated_at: datetime($updated_at)
		}]->(target)
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{
			"source_id":   link.SourceID,
			"target_id":   link.TargetID,
			"id":          link.ID,
			"source_port": link.SourcePort,
			"target_port": link.TargetPort,
			"weight":      link.Weight,
			"status":      link.Status,
			"metadata":    link.Metadata,
			"last_seen":   link.LastSeen.Format("2006-01-02T15:04:05Z"),
			"created_at":  link.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at":  link.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	})

	return err
}

// GetDevice retrieves a device by ID from Neo4j
func (r *Neo4jRepository) GetDevice(ctx context.Context, deviceID string) (*topology.Device, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (d:Device {id: $id})
		RETURN d
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{
			"id": deviceID,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	records, ok := result.(neo4j.ResultWithContext)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	if !records.Next(ctx) {
		return nil, nil // Device not found
	}

	record := records.Record()
	deviceNode, ok := record.Get("d")
	if !ok {
		return nil, fmt.Errorf("device node not found in result")
	}

	device, err := r.nodeToDevice(deviceNode.(neo4j.Node))
	if err != nil {
		return nil, fmt.Errorf("failed to convert node to device: %w", err)
	}

	return device, nil
}

// SearchDevices searches for devices by ID, name, or IP address with fuzzy matching
func (r *Neo4jRepository) SearchDevices(ctx context.Context, query string, limit int) ([]topology.Device, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	cypherQuery := `
		MATCH (d:Device)
		WHERE 
			toLower(d.id) CONTAINS toLower($query) OR
			toLower(d.name) CONTAINS toLower($query) OR
			toLower(d.ip_address) CONTAINS toLower($query) OR
			toLower(d.hardware) CONTAINS toLower($query)
		RETURN d
		ORDER BY
			CASE 
				WHEN d.id = $exactQuery THEN 1
				WHEN d.id STARTS WITH $query THEN 2
				WHEN d.name = $exactQuery THEN 3
				WHEN d.name STARTS WITH $query THEN 4
				ELSE 5
			END,
			d.id
		LIMIT $limit
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, cypherQuery, map[string]interface{}{
			"query":      query,
			"exactQuery": query,
			"limit":      limit,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	records, ok := result.(neo4j.ResultWithContext)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	var devices []topology.Device
	for records.Next(ctx) {
		record := records.Record()
		deviceNode, ok := record.Get("d")
		if !ok {
			continue
		}

		device, err := r.nodeToDevice(deviceNode.(neo4j.Node))
		if err != nil {
			continue // Skip invalid nodes
		}

		devices = append(devices, *device)
	}

	return devices, nil
}

// BulkAddDevices adds multiple devices in a single transaction
func (r *Neo4jRepository) BulkAddDevices(ctx context.Context, devices []topology.Device) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, device := range devices {
			query := `
				CREATE (d:Device {
					id: $id,
					type: $type,
					hardware: $hardware,
					instance: $instance,
					ip_address: $ip_address,
					location: $location,
					status: $status,
					layer: $layer,
					metadata: $metadata,
					last_seen: datetime($last_seen),
					created_at: datetime($created_at),
					updated_at: datetime($updated_at)
				})
			`

			_, err := tx.Run(ctx, query, map[string]interface{}{
				"id":         device.ID,
				"type":       device.Type,
				"hardware":   device.Hardware,
				"instance":   device.Instance,
				"location":   device.Location,
				"status":     device.Status,
				"layer":      device.Layer,
				"metadata":   device.Metadata,
				"last_seen":  device.LastSeen.Format("2006-01-02T15:04:05Z"),
				"created_at": device.CreatedAt.Format("2006-01-02T15:04:05Z"),
				"updated_at": device.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})

			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	return err
}

// BulkAddLinks adds multiple links in a single transaction
func (r *Neo4jRepository) BulkAddLinks(ctx context.Context, links []topology.Link) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, link := range links {
			query := `
				MATCH (source:Device {id: $source_id})
				MATCH (target:Device {id: $target_id})
				CREATE (source)-[l:CONNECTED {
					id: $id,
					source_port: $source_port,
					target_port: $target_port,
					weight: $weight,
					status: $status,
					metadata: $metadata,
					last_seen: datetime($last_seen),
					created_at: datetime($created_at),
					updated_at: datetime($updated_at)
				}]->(target)
			`

			_, err := tx.Run(ctx, query, map[string]interface{}{
				"source_id":   link.SourceID,
				"target_id":   link.TargetID,
				"id":          link.ID,
				"source_port": link.SourcePort,
				"target_port": link.TargetPort,
				"weight":      link.Weight,
				"status":      link.Status,
				"metadata":    link.Metadata,
				"last_seen":   link.LastSeen.Format("2006-01-02T15:04:05Z"),
				"created_at":  link.CreatedAt.Format("2006-01-02T15:04:05Z"),
				"updated_at":  link.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})

			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	return err
}

// FindReachableDevices finds all devices reachable from a given device using graph algorithms
func (r *Neo4jRepository) FindReachableDevices(ctx context.Context, deviceID string, opts topology.ReachabilityOptions) ([]topology.Device, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	var query string
	params := map[string]interface{}{
		"start_id": deviceID,
	}

	if opts.MaxHops > 0 {
		// Limit by max hops
		query = fmt.Sprintf(`
			MATCH path = (start:Device {id: $start_id})-[:CONNECTED*1..%d]-(reachable:Device)
			WHERE start <> reachable
			RETURN DISTINCT reachable
		`, opts.MaxHops)
	} else {
		// No hop limit
		query = `
			MATCH path = (start:Device {id: $start_id})-[:CONNECTED*]-(reachable:Device)
			WHERE start <> reachable
			RETURN DISTINCT reachable
		`
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, params)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute reachability query: %w", err)
	}

	records, ok := result.(neo4j.ResultWithContext)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	var devices []topology.Device
	for records.Next(ctx) {
		record := records.Record()
		deviceNode, ok := record.Get("reachable")
		if !ok {
			continue
		}

		device, err := r.nodeToDevice(deviceNode.(neo4j.Node))
		if err != nil {
			continue // Skip invalid nodes
		}

		devices = append(devices, *device)
	}

	return devices, nil
}

// FindShortestPath finds the shortest path between two devices using Cypher's shortest path algorithm
func (r *Neo4jRepository) FindShortestPath(ctx context.Context, fromID, toID string, opts topology.PathOptions) (*topology.Path, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: r.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (from:Device {id: $from_id}), (to:Device {id: $to_id})
		MATCH path = shortestPath((from)-[:CONNECTED*]-(to))
		RETURN path
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{
			"from_id": fromID,
			"to_id":   toID,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute shortest path query: %w", err)
	}

	records, ok := result.(neo4j.ResultWithContext)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	if !records.Next(ctx) {
		return nil, fmt.Errorf("no path found between devices")
	}

	record := records.Record()
	pathValue, ok := record.Get("path")
	if !ok {
		return nil, fmt.Errorf("path not found in result")
	}

	path, err := r.neo4jPathToTopologyPath(pathValue.(neo4j.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to convert Neo4j path: %w", err)
	}

	return path, nil
}

// Helper methods for data conversion
func (r *Neo4jRepository) nodeToDevice(node neo4j.Node) (*topology.Device, error) {
	props := node.Props

	// Extract and validate required fields
	id, ok := props["id"].(string)
	if !ok {
		return nil, fmt.Errorf("device id is not a string")
	}

	device := &topology.Device{
		ID: id,
	}

	// Extract optional fields with safe type assertions
	if deviceType, ok := props["type"].(string); ok {
		device.Type = deviceType
	}
	if hardware, ok := props["hardware"].(string); ok {
		device.Hardware = hardware
	}
	if instance, ok := props["instance"].(string); ok {
		device.Instance = instance
	}
	if location, ok := props["location"].(string); ok {
		device.Location = location
	}
	if status, ok := props["status"].(string); ok {
		device.Status = status
	}
	if layer, ok := props["layer"].(int64); ok {
		device.Layer = int(layer)
	}

	// Handle metadata map
	if metadata, ok := props["metadata"].(map[string]interface{}); ok {
		device.Metadata = make(map[string]string)
		for k, v := range metadata {
			if str, ok := v.(string); ok {
				device.Metadata[k] = str
			}
		}
	}

	return device, nil
}

func (r *Neo4jRepository) neo4jPathToTopologyPath(neo4jPath neo4j.Path) (*topology.Path, error) {
	var devices []topology.Device
	var links []topology.Link
	var totalCost float64

	// Extract devices from path nodes
	for _, node := range neo4jPath.Nodes {
		device, err := r.nodeToDevice(node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert path node: %w", err)
		}
		devices = append(devices, *device)
	}

	// Extract links from path relationships
	for _, relationship := range neo4jPath.Relationships {
		link := topology.Link{
			ID:       fmt.Sprintf("path-link-%d", relationship.Id),
			SourceID: fmt.Sprintf("%v", relationship.StartId),
			TargetID: fmt.Sprintf("%v", relationship.EndId),
		}

		// Extract relationship properties
		if sourcePort, ok := relationship.Props["source_port"].(string); ok {
			link.SourcePort = sourcePort
		}
		if targetPort, ok := relationship.Props["target_port"].(string); ok {
			link.TargetPort = targetPort
		}
		if weight, ok := relationship.Props["weight"].(float64); ok {
			link.Weight = weight
			totalCost += weight
		}
		if status, ok := relationship.Props["status"].(string); ok {
			link.Status = status
		}

		links = append(links, link)
	}

	return &topology.Path{
		Devices:   devices,
		Links:     links,
		TotalCost: totalCost,
		HopCount:  len(devices) - 1,
	}, nil
}

// Stub implementations for interface compliance
func (r *Neo4jRepository) UpdateDevice(ctx context.Context, device topology.Device) error {
	return fmt.Errorf("UpdateDevice not implemented for Neo4j")
}

func (r *Neo4jRepository) UpdateLink(ctx context.Context, link topology.Link) error {
	return fmt.Errorf("UpdateLink not implemented for Neo4j")
}

func (r *Neo4jRepository) RemoveDevice(ctx context.Context, deviceID string) error {
	return fmt.Errorf("RemoveDevice not implemented for Neo4j")
}

func (r *Neo4jRepository) RemoveLink(ctx context.Context, linkID string) error {
	return fmt.Errorf("RemoveLink not implemented for Neo4j")
}

func (r *Neo4jRepository) GetLink(ctx context.Context, linkID string) (*topology.Link, error) {
	return nil, fmt.Errorf("GetLink not implemented for Neo4j")
}

func (r *Neo4jRepository) FindDevicesByType(ctx context.Context, deviceType string) ([]topology.Device, error) {
	return nil, fmt.Errorf("FindDevicesByType not implemented for Neo4j")
}

func (r *Neo4jRepository) FindDevicesByHardware(ctx context.Context, hardware string) ([]topology.Device, error) {
	return nil, fmt.Errorf("FindDevicesByHardware not implemented for Neo4j")
}

func (r *Neo4jRepository) FindDevicesByInstance(ctx context.Context, instance string) ([]topology.Device, error) {
	return nil, fmt.Errorf("FindDevicesByInstance not implemented for Neo4j")
}

func (r *Neo4jRepository) GetDeviceLinks(ctx context.Context, deviceID string) ([]topology.Link, error) {
	return nil, fmt.Errorf("GetDeviceLinks not implemented for Neo4j")
}

func (r *Neo4jRepository) FindLinksByPort(ctx context.Context, deviceID, port string) ([]topology.Link, error) {
	return nil, fmt.Errorf("FindLinksByPort not implemented for Neo4j")
}

func (r *Neo4jRepository) ExtractSubTopology(ctx context.Context, deviceID string, opts topology.SubTopologyOptions) ([]topology.Device, []topology.Link, error) {
	return nil, nil, fmt.Errorf("ExtractSubTopology not implemented for Neo4j")
}

func (r *Neo4jRepository) GetDevices(ctx context.Context, opts topology.PaginationOptions) ([]topology.Device, *topology.PaginationResult, error) {
	return nil, nil, fmt.Errorf("GetDevices not implemented for Neo4j")
}

// Health checks the health of the Neo4j connection
func (r *Neo4jRepository) Health(ctx context.Context) error {
	return r.driver.VerifyConnectivity(ctx)
}
