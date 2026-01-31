-- Run this in the Supabase SQL Editor to create tables and a default user.
-- Project Settings -> Database -> SQL Editor -> New query

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS habits (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    color TEXT DEFAULT '#58cc02',
    frequency TEXT DEFAULT 'daily',
    frequency_details JSONB,
    locked BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS habit_logs (
    id SERIAL PRIMARY KEY,
    habit_id INTEGER REFERENCES habits(id) ON DELETE CASCADE,
    completed_at DATE NOT NULL,
    UNIQUE(habit_id, completed_at)
);

-- Seed one user for development (id=1). The backend uses user_id=1 until you add auth.
INSERT INTO users (id, email) VALUES (1, 'dev@example.com')
ON CONFLICT (id) DO NOTHING;

-- If your sequence is out of sync after manual inserts, run:
-- SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
