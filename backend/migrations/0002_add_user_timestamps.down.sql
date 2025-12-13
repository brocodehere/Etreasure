-- Drop the trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop the index
DROP INDEX IF EXISTS idx_users_deleted_at;

-- Remove the timestamp columns
ALTER TABLE users 
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS deleted_at;
