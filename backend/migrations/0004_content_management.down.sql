-- Down migration for content management
-- Removes content_pages and faqs tables

DROP TABLE IF EXISTS faqs CASCADE;
DROP TABLE IF EXISTS content_pages CASCADE;
