package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// GetOrCreateUserByAuthID looks up a user by their Supabase auth UUID.
// If the user doesn't exist, it creates a new user with the provided email.
// Returns the integer user_id for use with RLS and application logic.
func GetOrCreateUserByAuthID(ctx context.Context, authUserID string, email string) (int, error) {
	var userID int

	// Try to find existing user by auth_user_id
	err := DB.QueryRow(ctx,
		`SELECT id FROM users WHERE auth_user_id = $1`,
		authUserID).Scan(&userID)

	if err == nil {
		// User found
		log.Printf("[DB] Found existing user: id=%d auth_user_id=%s", userID, authUserID)
		return userID, nil
	}

	if err != pgx.ErrNoRows {
		// Unexpected database error
		return 0, fmt.Errorf("failed to query user: %w", err)
	}

	// User not found, create new user
	log.Printf("[DB] Creating new user: auth_user_id=%s email=%s", authUserID, email)

	err = DB.QueryRow(ctx,
		`INSERT INTO users (auth_user_id, email)
		 VALUES ($1, $2)
		 ON CONFLICT (email) DO UPDATE SET auth_user_id = EXCLUDED.auth_user_id
		 RETURNING id`,
		authUserID, email).Scan(&userID)

	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("[DB] Created new user: id=%d auth_user_id=%s email=%s", userID, authUserID, email)
	return userID, nil
}
