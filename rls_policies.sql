-- Row Level Security (RLS) Setup for Habit Tracker
-- Run this in Supabase SQL Editor after running schema.sql
--
-- IMPORTANT: This uses your current integer-based user_id system.
-- When you integrate Supabase Auth, you'll need to update these policies
-- to use auth.uid() instead (see commented sections at bottom).

-- ============================================================
-- ENABLE RLS ON ALL TABLES
-- ============================================================

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE habits ENABLE ROW LEVEL SECURITY;
ALTER TABLE habit_logs ENABLE ROW LEVEL SECURITY;

-- ============================================================
-- POLICIES FOR USERS TABLE
-- ============================================================

-- Users can view their own profile
CREATE POLICY "Users can view their own profile"
  ON users FOR SELECT
  USING (id = current_setting('app.user_id', true)::INTEGER);

-- Users can update their own profile
CREATE POLICY "Users can update their own profile"
  ON users FOR UPDATE
  USING (id = current_setting('app.user_id', true)::INTEGER);

-- Allow inserting new users (for registration)
CREATE POLICY "Allow user registration"
  ON users FOR INSERT
  WITH CHECK (true);

-- ============================================================
-- POLICIES FOR HABITS TABLE
-- ============================================================

-- Users can view only their own habits
CREATE POLICY "Users can view their own habits"
  ON habits FOR SELECT
  USING (user_id = current_setting('app.user_id', true)::INTEGER);

-- Users can insert their own habits
CREATE POLICY "Users can insert their own habits"
  ON habits FOR INSERT
  WITH CHECK (user_id = current_setting('app.user_id', true)::INTEGER);

-- Users can update their own habits
CREATE POLICY "Users can update their own habits"
  ON habits FOR UPDATE
  USING (user_id = current_setting('app.user_id', true)::INTEGER);

-- Users can delete their own habits
CREATE POLICY "Users can delete their own habits"
  ON habits FOR DELETE
  USING (user_id = current_setting('app.user_id', true)::INTEGER);

-- ============================================================
-- POLICIES FOR HABIT_LOGS TABLE
-- ============================================================

-- Users can view logs for their own habits
CREATE POLICY "Users can view their own habit logs"
  ON habit_logs FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM habits
      WHERE habits.id = habit_logs.habit_id
      AND habits.user_id = current_setting('app.user_id', true)::INTEGER
    )
  );

-- Users can insert logs for their own habits
CREATE POLICY "Users can insert their own habit logs"
  ON habit_logs FOR INSERT
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM habits
      WHERE habits.id = habit_logs.habit_id
      AND habits.user_id = current_setting('app.user_id', true)::INTEGER
    )
  );

-- Users can update logs for their own habits
CREATE POLICY "Users can update their own habit logs"
  ON habit_logs FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM habits
      WHERE habits.id = habit_logs.habit_id
      AND habits.user_id = current_setting('app.user_id', true)::INTEGER
    )
  );

-- Users can delete logs for their own habits
CREATE POLICY "Users can delete their own habit logs"
  ON habit_logs FOR DELETE
  USING (
    EXISTS (
      SELECT 1 FROM habits
      WHERE habits.id = habit_logs.habit_id
      AND habits.user_id = current_setting('app.user_id', true)::INTEGER
    )
  );

-- ============================================================
-- VERIFY RLS IS ENABLED
-- ============================================================

-- Run this to verify RLS is enabled on all tables:
-- SELECT tablename, rowsecurity
-- FROM pg_tables
-- WHERE schemaname = 'public'
-- AND tablename IN ('users', 'habits', 'habit_logs');

-- ============================================================
-- FUTURE: SUPABASE AUTH INTEGRATION
-- ============================================================
-- When you integrate Supabase Auth (JWT tokens), replace the policies above with these:
--
-- For habits table example:
-- CREATE POLICY "Users can view their own habits"
--   ON habits FOR SELECT
--   USING (user_id = (auth.uid()::text)::INTEGER);
--
-- Or if you migrate user_id to UUID:
-- CREATE POLICY "Users can view their own habits"
--   ON habits FOR SELECT
--   USING (user_id = auth.uid());
--
-- Note: auth.uid() returns the authenticated user's UUID from the JWT token
-- You'll need to either:
-- 1. Change user_id column to UUID type, OR
-- 2. Create a mapping table between auth.users.id (UUID) and your users.id (INTEGER)
