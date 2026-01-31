# Habit Tracker Backend (Go + Supabase)

Backend API for the habit tracker app. Built with Go, [chi](https://github.com/go-chi/chi) router, and Supabase (PostgreSQL).

## Quick start

1. **Install Go 1.22+** and ensure `go` is in your PATH.

2. **Supabase setup**
   - Create a project at [supabase.com](https://supabase.com).
   - In the SQL Editor, run the contents of `schema.sql` to create tables and a default user.

3. **Environment**
   - Copy `.env.example` to `.env`.
   - Set `DATABASE_URL` to your Supabase connection string (Project Settings → Database → Connection string URI).

4. **Install dependencies and run**
   ```bash
   go mod tidy
   go run cmd/server/main.go
   ```
   You should see: `Connected to Supabase successfully` and `Server starting on http://localhost:8080`.

5. **Test**
   ```bash
   # Create a habit
   curl -X POST http://localhost:8080/api/habits -H "Content-Type: application/json" -d "{\"title\":\"Drink Water\",\"color\":\"#58cc02\"}"

   # List habits
   curl http://localhost:8080/api/habits
   ```

## Running tests

```bash
# Unit tests only (no DB): computeStreak and invalid-input validation
go test ./internal/handlers/... -v -short

# All tests including integration (requires DATABASE_URL in .env)
go test ./internal/handlers/... -v
```

- **Unit tests** (e.g. `TestComputeStreak`, invalid-input cases) run without a database.
- **Integration tests** (GET/POST habits, toggle log) require `DATABASE_URL`; they are skipped if unset.

## Project layout

- `cmd/server/main.go` – Entry point; loads config, connects DB, starts server.
- `internal/server/router.go` – Router and middleware (shared by main and tests).
- `internal/database/db.go` – Supabase connection pool.
- `internal/models/habit.go` – Structs for habits and requests.
- `internal/handlers/habit_handler.go` – HTTP handlers for habits and logs.
- `internal/handlers/habit_handler_test.go` – Unit and integration tests.
- `schema.sql` – SQL to run in Supabase (tables + seed user).

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/habits` | List all habits (with completed dates, status, streak). |
| POST | `/api/habits` | Create a habit (JSON: `title`, optional `icon`, `color`, `frequency`). |
| POST | `/api/habits/{id}/log` | Toggle completion for a habit on a date. Query: `?date=YYYY-MM-DD` (default: today). |

## Debugging

The code uses `log.Println` and `log.Printf` with prefixes like `[GetHabits]` and `[DB]`. Watch your terminal while calling the API to trace requests and errors.

## Next steps (from instructions.md)

- **Auth**: Use Supabase Auth and filter habits by `user_id` from the JWT.
- **Validation**: Validate input (e.g. non-empty title, hex color).
- **Errors**: Centralize JSON error responses.
