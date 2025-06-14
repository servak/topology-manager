-- 009_add_created_by_to_classification_rules.sql
-- classification_rulesテーブルにcreated_byカラムを追加

ALTER TABLE classification_rules 
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) DEFAULT 'system';