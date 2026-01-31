// Package database provides the Supabase (PostgreSQL) connection pool.
// In Go, a "package" is like a namespace: all files in the same directory share it.
package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is a global connection pool. The pool reuses connections instead of
// opening a new one per request, which is much more efficient.
// We use a global here for simplicity; in larger apps you might inject it via handlers.
var DB *pgxpool.Pool

// Connect establishes a connection to the database using the given connection string.
// It parses the config, creates a pool, and pings the DB to verify connectivity.
// If any step fails, it calls log.Fatal which exits the program (appropriate for startup).
func Connect(connString string) {
	log.Println("[DB] Parsing connection string...")

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatal("[DB] Unable to parse DB config:", err)
	}

	// Context with timeout: we don't want to hang forever if the DB is unreachable.
	// In Go, context.Context is used for cancellation and deadlines across async operations.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Always call cancel when done to release resources

	log.Println("[DB] Creating connection pool...")

	DB, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("[DB] Unable to connect to database:", err)
	}

	if err := DB.Ping(ctx); err != nil {
		log.Fatal("[DB] Unable to ping database - check your .env DATABASE_URL:", err)
	}

	log.Println("[DB] Connected to Supabase successfully")
}
