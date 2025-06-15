-- 012_redesign_devices_table.sql
-- 抜本的改革: デバイステーブル設計の刷新

-- まず既存の分類ルール名を新命名規則に変更
UPDATE classification_rules SET name = 'border-router-name-pattern' WHERE name = 'Border Router - Name Pattern';
UPDATE classification_rules SET name = 'border-router-hardware-type' WHERE name = 'Border Router - Hardware Type';
UPDATE classification_rules SET name = 'security-firewall-name-pattern' WHERE name = 'Security - Firewall Name Pattern';
UPDATE classification_rules SET name = 'security-firewall-hardware' WHERE name = 'Security - Firewall Hardware';
UPDATE classification_rules SET name = 'spine-switch-name-pattern' WHERE name = 'Spine Switch - Name Pattern';
UPDATE classification_rules SET name = 'spine-switch-hardware-type' WHERE name = 'Spine Switch - Hardware Type';
UPDATE classification_rules SET name = 'leaf-switch-name-pattern' WHERE name = 'Leaf Switch - Name Pattern';
UPDATE classification_rules SET name = 'leaf-switch-hardware-type' WHERE name = 'Leaf Switch - Hardware Type';
UPDATE classification_rules SET name = 'server-name-pattern' WHERE name = 'Server - Name Pattern';
UPDATE classification_rules SET name = 'server-hardware-type' WHERE name = 'Server - Hardware Type';
UPDATE classification_rules SET name = 'storage-name-pattern' WHERE name = 'Storage - Name Pattern';
UPDATE classification_rules SET name = 'storage-hardware-type' WHERE name = 'Storage - Hardware Type';
UPDATE classification_rules SET name = 'edge-leaf-switches' WHERE name = 'Edge Leaf';

-- 既存テーブルを削除
DROP TABLE IF EXISTS device_classifications CASCADE;
DROP TABLE IF EXISTS devices CASCADE;
DROP TABLE IF EXISTS links CASCADE;

-- 新しいデバイステーブル
CREATE TABLE devices (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    hardware VARCHAR(255),
    
    -- 分類情報（統合）
    layer_id INTEGER REFERENCES hierarchy_layers(id),
    device_type VARCHAR(100),
    classified_by VARCHAR(255),
    
    -- 運用情報
    metadata JSONB,
    last_seen TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- 制約
    CONSTRAINT valid_classified_by CHECK (
        classified_by IS NULL OR 
        classified_by ~ '^(rule:[a-zA-Z0-9\-]+|user:[a-zA-Z0-9\-]+|system:[a-zA-Z0-9\-]+)$'
    )
);

-- 新しいリンクステーブル
CREATE TABLE links (
    id VARCHAR(255) PRIMARY KEY,
    source_id VARCHAR(255) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    target_id VARCHAR(255) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    source_port VARCHAR(255),
    target_port VARCHAR(255),
    weight FLOAT DEFAULT 1.0,
    metadata JSONB,
    last_seen TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(source_id, target_id, source_port, target_port)
);

-- 分類ルールテーブルの更新（ルール名制約追加）
ALTER TABLE classification_rules 
ADD CONSTRAINT valid_rule_name CHECK (name ~ '^[a-zA-Z0-9\-]+$'),
ADD CONSTRAINT rule_name_length CHECK (length(name) >= 2 AND length(name) <= 50);

-- インデックス
CREATE INDEX idx_devices_layer_id ON devices(layer_id);
CREATE INDEX idx_devices_type ON devices(type);
CREATE INDEX idx_devices_classified_by ON devices(classified_by);
CREATE INDEX idx_links_source_id ON links(source_id);
CREATE INDEX idx_links_target_id ON links(target_id);

-- コメント
COMMENT ON TABLE devices IS '新設計: 分類情報統合型デバイステーブル';
COMMENT ON COLUMN devices.classified_by IS 'rule:xxx, user:xxx, system:xxx形式';
COMMENT ON TABLE links IS '新設計: シンプルなリンクテーブル';