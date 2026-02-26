package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// BeginTxWithUser starts a transaction and sets the RLS context for the given user_id.
// This ensures all queries within the transaction are scoped to that user via Row Level Security.
// The transaction must be committed or rolled back by the caller.
func BeginTxWithUser(ctx context.Context, userID int) (pgx.Tx, error) {
	// Start a transaction
	tx, err := DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set the user context for RLS policies
	// SET LOCAL is transaction-scoped and automatically cleared when the transaction ends
	_, err = tx.Exec(ctx, "SET LOCAL app.user_id = $1", userID)
	if err != nil {
		// If setting the context fails, rollback the transaction
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("failed to set user context: %w", err)
	}

	return tx, nil
}
