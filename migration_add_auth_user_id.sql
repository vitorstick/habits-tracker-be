-- Migration: Add auth_user_id column to users table
-- Run this in Supabase SQL Editor after schema.sql
--
-- This allows mapping Supabase Auth UUIDs to our integer user_id

ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_user_id UUID UNIQUE;

-- Create an index for fast lookups by auth_user_id
CREATE INDEX IF NOT EXISTS idx_users_auth_user_id ON users(auth_user_id);

-- Update the existing dev user to have a placeholder UUID (optional)
-- UPDATE users SET auth_user_id = '00000000-0000-0000-0000-000000000001' WHERE id = 1 AND auth_user_id IS NULL;
