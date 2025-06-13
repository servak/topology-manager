-- 005_remove_ip_address.sql
-- Remove ip_address column completely from devices table

-- Drop any indexes on ip_address column if they exist
DROP INDEX IF EXISTS idx_devices_ip_address;

-- Remove ip_address column
ALTER TABLE devices DROP COLUMN IF EXISTS ip_address;