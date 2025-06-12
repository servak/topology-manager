-- 002_create_links.sql
CREATE TABLE IF NOT EXISTS links (
    id VARCHAR(255) PRIMARY KEY,
    source_id VARCHAR(255) NOT NULL,
    target_id VARCHAR(255) NOT NULL,
    source_port VARCHAR(100) NOT NULL,
    target_port VARCHAR(100) NOT NULL,
    weight DECIMAL(10,2) DEFAULT 1.0,
    status VARCHAR(50) DEFAULT 'unknown',
    metadata JSONB DEFAULT '{}',
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (source_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES devices(id) ON DELETE CASCADE,
    
    -- 同じポート間の重複リンクを防ぐ
    UNIQUE(source_id, source_port, target_id, target_port)
);

-- インデックス作成
CREATE INDEX IF NOT EXISTS idx_links_source_id ON links(source_id);
CREATE INDEX IF NOT EXISTS idx_links_target_id ON links(target_id);
CREATE INDEX IF NOT EXISTS idx_links_source_port ON links(source_id, source_port);
CREATE INDEX IF NOT EXISTS idx_links_target_port ON links(target_id, target_port);
CREATE INDEX IF NOT EXISTS idx_links_weight ON links(weight);
CREATE INDEX IF NOT EXISTS idx_links_last_seen ON links(last_seen);

-- 双方向検索用のインデックス
CREATE INDEX IF NOT EXISTS idx_links_bidirectional ON links(source_id, target_id);

-- updated_at の自動更新トリガー
CREATE TRIGGER update_links_updated_at BEFORE UPDATE ON links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();