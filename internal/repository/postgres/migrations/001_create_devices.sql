-- 001_create_devices.sql
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(100) NOT NULL,
    hardware VARCHAR(255),
    instance VARCHAR(255),
    ip_address INET,
    location VARCHAR(255),
    status VARCHAR(50) DEFAULT 'unknown',
    layer INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- インデックス作成
CREATE INDEX IF NOT EXISTS idx_devices_name ON devices(name);
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type);
CREATE INDEX IF NOT EXISTS idx_devices_hardware ON devices(hardware);
CREATE INDEX IF NOT EXISTS idx_devices_instance ON devices(instance);
CREATE INDEX IF NOT EXISTS idx_devices_layer ON devices(layer);
CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(last_seen);

-- updated_at の自動更新トリガー
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_devices_updated_at BEFORE UPDATE ON devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();