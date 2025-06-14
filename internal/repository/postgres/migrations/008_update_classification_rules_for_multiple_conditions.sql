-- 008_update_classification_rules_for_multiple_conditions.sql
-- 分類ルールテーブルを複数条件対応に更新

-- 新しいカラムを追加
ALTER TABLE classification_rules 
ADD COLUMN logic_operator VARCHAR(10) DEFAULT 'AND',
ADD COLUMN conditions JSONB,
ADD COLUMN created_by VARCHAR(255) DEFAULT 'system';

-- 既存データを新しい構造に移行
UPDATE classification_rules 
SET conditions = jsonb_build_array(
    jsonb_build_object(
        'field', field,
        'operator', operator,
        'value', value
    )
),
logic_operator = 'AND'
WHERE conditions IS NULL;

-- 古いカラムを削除
ALTER TABLE classification_rules 
DROP COLUMN field,
DROP COLUMN operator,
DROP COLUMN value;

-- インデックス追加（JSONB検索用）
CREATE INDEX IF NOT EXISTS idx_classification_rules_conditions ON classification_rules USING gin(conditions);