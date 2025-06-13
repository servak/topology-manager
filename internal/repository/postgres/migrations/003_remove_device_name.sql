-- 003_remove_device_name.sql
-- Remove name column and related constraints from devices table

-- Drop the index on name column
DROP INDEX IF EXISTS idx_devices_name;

-- Drop the unique constraint on name column if it exists as a separate constraint
-- (The UNIQUE constraint is part of the column definition, so it will be dropped with the column)

-- Drop the name column
ALTER TABLE devices DROP COLUMN IF EXISTS name;