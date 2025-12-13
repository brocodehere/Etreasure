-- Down migration for banner schema update
-- This restores the original banner table structure

-- Drop the updated banners table
DROP TABLE IF EXISTS banners CASCADE;

-- Recreate the original banners table structure
CREATE TABLE banners (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    slot TEXT NOT NULL,
    target_url TEXT,
    priority INT NOT NULL DEFAULT 0,
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,
    desktop_media_id INT REFERENCES media(id) ON DELETE SET NULL,
    mobile_media_id INT REFERENCES media(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_banners_slot_priority ON banners(slot, priority);
