package sqlite

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SQLite-specific migrations
// Note: SQLite uses TEXT for JSON data instead of JSONB

const createDevicesTable = `
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    type TEXT,
    hardware TEXT,
    ip_address TEXT,
    
    -- Classification information (integrated)
    layer_id INTEGER,
    device_type TEXT,
    classified_by TEXT, -- "user:username", "rule:ruleName", "system:auto"
    
    -- Metadata and timestamps
    metadata TEXT, -- JSON data stored as TEXT in SQLite
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CHECK (classified_by IS NULL OR 
           classified_by LIKE 'user:%' OR 
           classified_by LIKE 'rule:%' OR 
           classified_by = 'system:auto')
);`

const createLinksTable = `
CREATE TABLE IF NOT EXISTS links (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    source_port TEXT,
    target_port TEXT,
    weight REAL DEFAULT 1.0,
    metadata TEXT, -- JSON data stored as TEXT
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (source_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES devices(id) ON DELETE CASCADE,
    
    -- Unique constraint for device-port pairs
    UNIQUE(source_id, target_id, source_port, target_port)
);`

const createHierarchyLayersTable = `
CREATE TABLE IF NOT EXISTS hierarchy_layers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    order_index INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(name)
);`

const createClassificationRulesTable = `
CREATE TABLE IF NOT EXISTS classification_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    conditions TEXT, -- JSON array of conditions
    logic_operator TEXT DEFAULT 'AND', -- 'AND' or 'OR'
    layer INTEGER NOT NULL,
    device_type TEXT NOT NULL,
    priority INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    confidence REAL,
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CHECK (logic_operator IN ('AND', 'OR')),
    CHECK (priority >= 0),
    CHECK (confidence IS NULL OR (confidence >= 0.0 AND confidence <= 1.0)),
    CHECK (name LIKE 'rule:%' OR name NOT LIKE '%:%'), -- Allow both prefixed and non-prefixed names
    
    UNIQUE(name)
);`

const createClassificationSuggestionsTable = `
CREATE TABLE IF NOT EXISTS classification_suggestions (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    affected_devices TEXT, -- JSON array of device IDs
    based_on_devices TEXT, -- JSON array of device IDs this suggestion is based on
    confidence REAL,
    status TEXT DEFAULT 'pending', -- 'pending', 'accepted', 'rejected'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CHECK (status IN ('pending', 'accepted', 'rejected')),
    CHECK (confidence IS NULL OR (confidence >= 0.0 AND confidence <= 1.0)),
    
    FOREIGN KEY (rule_id) REFERENCES classification_rules(id) ON DELETE CASCADE
);`

const createIndexes = `
-- Device indexes for better performance
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type);
CREATE INDEX IF NOT EXISTS idx_devices_hardware ON devices(hardware);
CREATE INDEX IF NOT EXISTS idx_devices_layer_id ON devices(layer_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_type ON devices(device_type);
CREATE INDEX IF NOT EXISTS idx_devices_classified_by ON devices(classified_by);
CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(last_seen);

-- Link indexes
CREATE INDEX IF NOT EXISTS idx_links_source_id ON links(source_id);
CREATE INDEX IF NOT EXISTS idx_links_target_id ON links(target_id);
CREATE INDEX IF NOT EXISTS idx_links_source_port ON links(source_id, source_port);
CREATE INDEX IF NOT EXISTS idx_links_target_port ON links(target_id, target_port);
CREATE INDEX IF NOT EXISTS idx_links_last_seen ON links(last_seen);

-- Classification rule indexes
CREATE INDEX IF NOT EXISTS idx_classification_rules_active ON classification_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_classification_rules_priority ON classification_rules(priority);
CREATE INDEX IF NOT EXISTS idx_classification_rules_layer ON classification_rules(layer);
CREATE INDEX IF NOT EXISTS idx_classification_rules_device_type ON classification_rules(device_type);

-- Classification suggestion indexes
CREATE INDEX IF NOT EXISTS idx_classification_suggestions_status ON classification_suggestions(status);
CREATE INDEX IF NOT EXISTS idx_classification_suggestions_rule_id ON classification_suggestions(rule_id);
CREATE INDEX IF NOT EXISTS idx_classification_suggestions_confidence ON classification_suggestions(confidence);

-- Hierarchy layer indexes
CREATE INDEX IF NOT EXISTS idx_hierarchy_layers_order_index ON hierarchy_layers(order_index);`

// insertDefaultHierarchyLayers inserts default hierarchy layers
const insertDefaultHierarchyLayers = `
INSERT OR IGNORE INTO hierarchy_layers (id, name, description, order_index) VALUES
(1, 'Core', 'Core network layer - backbone switches and routers', 1),
(2, 'Distribution', 'Distribution layer - aggregation switches', 2),  
(3, 'Access', 'Access layer - edge switches connecting end devices', 3),
(4, 'Server', 'Server layer - physical and virtual servers', 4),
(5, 'Unknown', 'Unclassified devices', 99);`

// RunMigrations executes all SQLite migrations
func RunMigrations(db *sqlx.DB) error {
	migrations := []string{
		createDevicesTable,
		createLinksTable,
		createHierarchyLayersTable,
		createClassificationRulesTable,
		createClassificationSuggestionsTable,
		createIndexes,
		insertDefaultHierarchyLayers,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
		}
	}

	return nil
}
