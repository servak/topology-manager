-- 013_fix_classified_by_constraint.sql
-- Fix classified_by constraint to allow empty string

-- Drop the existing constraint
ALTER TABLE devices DROP CONSTRAINT IF EXISTS valid_classified_by;

-- Add new constraint that allows empty string and NULL
ALTER TABLE devices ADD CONSTRAINT valid_classified_by CHECK (
    classified_by = '' OR 
    classified_by IS NULL OR 
    classified_by ~ '^(rule:[a-zA-Z0-9\-]+|user:[a-zA-Z0-9\-]+|system:[a-zA-Z0-9\-]+)$'
);