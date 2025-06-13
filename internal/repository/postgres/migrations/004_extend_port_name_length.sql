-- 004_extend_port_name_length.sql
-- Extend port name length from VARCHAR(100) to VARCHAR(255) to accommodate longer port names

-- Extend source_port column
ALTER TABLE links ALTER COLUMN source_port TYPE VARCHAR(255);

-- Extend target_port column  
ALTER TABLE links ALTER COLUMN target_port TYPE VARCHAR(255);