-- Update banner table schema to match backend expectations
-- This migration updates the existing banner table structure

-- Drop the existing banners table if it exists with old structure
DROP TABLE IF EXISTS banners CASCADE;

-- Create banners table with correct schema for backend
CREATE TABLE banners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    image_url TEXT NOT NULL,
    link_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for banners
CREATE INDEX idx_banners_is_active ON banners(is_active);
CREATE INDEX idx_banners_sort_order ON banners(sort_order);
CREATE INDEX idx_banners_starts_at ON banners(starts_at);
CREATE INDEX idx_banners_ends_at ON banners(ends_at);
CREATE INDEX idx_banners_active_dates ON banners(is_active, starts_at, ends_at);
