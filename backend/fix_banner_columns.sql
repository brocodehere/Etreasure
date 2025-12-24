-- Add multiple image columns to banners table for responsive images
-- This script fixes the missing columns that are causing the 500 error

ALTER TABLE banners ADD COLUMN IF NOT EXISTS desktop_image_url TEXT;
ALTER TABLE banners ADD COLUMN IF NOT EXISTS laptop_image_url TEXT;
ALTER TABLE banners ADD COLUMN IF NOT EXISTS mobile_image_url TEXT;

-- Create indexes for the new image columns
CREATE INDEX IF NOT EXISTS idx_banners_desktop_image ON banners(desktop_image_url) WHERE desktop_image_url IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_banners_laptop_image ON banners(laptop_image_url) WHERE laptop_image_url IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_banners_mobile_image ON banners(mobile_image_url) WHERE mobile_image_url IS NOT NULL;

-- Update existing banners to have desktop_image_url populated from image_url
UPDATE banners SET desktop_image_url = image_url WHERE desktop_image_url IS NULL AND image_url IS NOT NULL;
