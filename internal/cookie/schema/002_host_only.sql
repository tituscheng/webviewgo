-- Add host_only to databases created before this column existed.
-- The migration runner ignores the "duplicate column name" error this
-- statement raises on databases that already have the column.
ALTER TABLE cookies ADD COLUMN host_only INTEGER NOT NULL DEFAULT 0;
