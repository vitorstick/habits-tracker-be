-- Run this if you already have the habits table and need to add the new UX fields.
-- Safe to run multiple times (IF NOT EXISTS / do nothing on conflict).

ALTER TABLE habits ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE habits ADD COLUMN IF NOT EXISTS frequency_details JSONB;
ALTER TABLE habits ADD COLUMN IF NOT EXISTS locked BOOLEAN DEFAULT false;
