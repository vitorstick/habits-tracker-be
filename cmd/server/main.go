// Package main is the entry point for the habit tracker backend server.
// In Go, the executable is built from the package named "main" with a func main().
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"habit-tracker-be/internal/database"
	"habit-tracker-be/internal/server"
)

func main() {
	// 1. Load config from .env file (if present). godotenv.Load() is safe to call
	// even when .env doesn't exist - it just returns an error we can ignore.
	if err := godotenv.Load(); err != nil {
		log.Println("[main] No .env file found (using system env vars)")
	} else {
		log.Println("[main] Loaded .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("[main] Using PORT=%s", port)

	// 2. Connect to Supabase (PostgreSQL). DATABASE_URL must be set.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("[main] DATABASE_URL is not set - add it to .env (see .env.example)")
	}
	database.Connect(dbURL)
	defer database.DB.Close() // Close pool when main exits

	// 3. Setup router (middleware + routes). Router() is in internal/server so tests can use it.
	r := server.Router()

	// 4. Start the HTTP server. ListenAndServe blocks until the server fails.
	log.Printf("[main] Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("[main] Server failed: %v", err)
	}
}
