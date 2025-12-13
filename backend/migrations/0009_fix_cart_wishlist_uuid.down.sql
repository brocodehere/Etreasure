-- Migration: Remove fixed cart and wishlist tables

DROP TABLE IF EXISTS wishlist CASCADE;
DROP TABLE IF EXISTS cart CASCADE;
