-- 007_create_hierarchy_layers_table.sql
-- 階層レイヤーテーブルの作成

CREATE TABLE IF NOT EXISTS hierarchy_layers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    order_index INTEGER NOT NULL DEFAULT 0,
    color VARCHAR(7) NOT NULL DEFAULT '#3498db',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- インデックス作成
CREATE INDEX IF NOT EXISTS idx_hierarchy_layers_order ON hierarchy_layers(order_index);

-- デフォルトデータの挿入
INSERT INTO hierarchy_layers (id, name, description, order_index, color) VALUES
(0, 'Internet Gateway', 'External internet connection point', 0, '#e74c3c'),
(1, 'Firewall', 'Security appliances', 1, '#e67e22'),
(2, 'Core Router', 'Core network routing', 2, '#f39c12'),
(3, 'Distribution', 'Distribution layer switches', 3, '#3498db'),
(4, 'Access', 'Access layer switches', 4, '#2ecc71'),
(5, 'Server', 'End devices and servers', 5, '#95a5a6')
ON CONFLICT (id) DO NOTHING;

-- IDシーケンスを適切に設定（シーケンスが存在する場合のみ）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_sequences WHERE sequencename = 'hierarchy_layers_id_seq') THEN
        PERFORM setval('hierarchy_layers_id_seq', COALESCE((SELECT MAX(id) FROM hierarchy_layers), 0) + 1, false);
    END IF;
END
$$;

-- updated_atトリガーの作成
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_hierarchy_layers_updated_at 
    BEFORE UPDATE ON hierarchy_layers 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();