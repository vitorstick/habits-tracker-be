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

6. **Production API**
   - The API is also deployed at: `https://habits-tracker-be.onrender.com/api/`

5. **Test & API Usage**
   - For quick testing, you can use the provided [REST Client](api_tests.http) file (requires VS Code + REST Client extension).
   - For scripting, see the [Python Client Sample](scripts/api_client.py).
   - For a full guide on using this as a standalone API (including authentication), see [API_USAGE.md](API_USAGE.md).


   ```bash
   # Quick health check (no auth)
   curl http://localhost:8080/api/auth/me
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
| GET | `/api/habits` | List habits for a specific day. Query: `?date=YYYY-MM-DD` (default: today). Returns computed status and streak relative to the date. |
| POST | `/api/habits` | Create a habit (JSON: `title`, optional `icon`, `color`, `frequency`). |
| POST | `/api/habits/{id}/log` | Toggle completion for a habit on a date. Query: `?date=YYYY-MM-DD` (default: today). |
| DELETE| `/api/habits/{id}` | Delete a habit and its logs. |

## Debugging

The code uses `log.Println` and `log.Printf` with prefixes like `[GetHabits]` and `[DB]`. Watch your terminal while calling the API to trace requests and errors.

## Next steps (from instructions.md)

- **Auth**: Use Supabase Auth and filter habits by `user_id` from the JWT.
- **Validation**: Validate input (e.g. non-empty title, hex color).
- **Errors**: Centralize JSON error responses.
