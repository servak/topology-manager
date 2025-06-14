-- Device Classifications Table
CREATE TABLE IF NOT EXISTS device_classifications (
    id UUID PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL UNIQUE,
    layer INTEGER NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    is_manual BOOLEAN NOT NULL DEFAULT false,
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraint to devices table
    CONSTRAINT fk_device_classifications_device_id 
        FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

-- Classification Rules Table  
CREATE TABLE IF NOT EXISTS classification_rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    field VARCHAR(50) NOT NULL, -- name, hardware, ip_address, type
    operator VARCHAR(20) NOT NULL, -- contains, starts_with, ends_with, equals, regex
    value VARCHAR(255) NOT NULL,
    layer INTEGER NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Classification Suggestions Table
CREATE TABLE IF NOT EXISTS classification_suggestions (
    id UUID PRIMARY KEY,
    rule_id UUID NOT NULL,
    confidence DECIMAL(5,4) NOT NULL, -- 0.0000 to 1.0000
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, accepted, rejected
    affected_devices TEXT[], -- Array of device IDs
    based_on_devices TEXT[], -- Array of device IDs used to generate suggestion
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_classification_suggestions_rule_id 
        FOREIGN KEY (rule_id) REFERENCES classification_rules(id) ON DELETE CASCADE
);

-- Hierarchy Layers Table
CREATE TABLE IF NOT EXISTS hierarchy_layers (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    order_index INTEGER NOT NULL,
    color VARCHAR(7) NOT NULL, -- Hex color code like #FF0000
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_device_classifications_device_id ON device_classifications(device_id);
CREATE INDEX IF NOT EXISTS idx_device_classifications_layer ON device_classifications(layer);
CREATE INDEX IF NOT EXISTS idx_classification_rules_active ON classification_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_classification_rules_priority ON classification_rules(priority DESC);
CREATE INDEX IF NOT EXISTS idx_classification_suggestions_status ON classification_suggestions(status);
CREATE INDEX IF NOT EXISTS idx_hierarchy_layers_order ON hierarchy_layers(order_index);

-- Insert default hierarchy layers
INSERT INTO hierarchy_layers (id, name, description, order_index, color) VALUES
    (0, 'Internet Gateway', 'External internet connection point', 0, '#e74c3c'),
    (1, 'Firewall', 'Security appliances', 1, '#e67e22'),
    (2, 'Core Router', 'Core network routing', 2, '#f39c12'),
    (3, 'Distribution', 'Distribution layer switches', 3, '#3498db'),
    (4, 'Access', 'Access layer switches', 4, '#2ecc71'),
    (5, 'Server', 'End devices and servers', 5, '#95a5a6')
ON CONFLICT (id) DO NOTHING;