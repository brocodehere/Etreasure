package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// First, add session_id column if it doesn't exist
	_, err = db.ExecContext(ctx, `
		ALTER TABLE wishlist 
		ADD COLUMN IF NOT EXISTS session_id TEXT
	`)
	if err != nil {
		log.Printf("Error adding session_id column (might already exist): %v", err)
	}

	// Create index for session_id if it doesn't exist
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_wishlist_session ON wishlist(session_id)
	`)
	if err != nil {
		log.Printf("Error creating index (might already exist): %v", err)
	}

	// Update existing records to use session_id for demo purposes
	// This converts existing customer_id records to session_id format for local development
	result, err := db.ExecContext(ctx, `
		UPDATE wishlist 
		SET session_id = 'demo-session-127.0.0.1' 
		WHERE session_id IS NULL AND customer_id IS NOT NULL
	`)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated %d wishlist records to use session_id\n", rowsAffected)

	// Verify the update
	var count int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM wishlist WHERE session_id = 'demo-session-127.0.0.1'
	`).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d wishlist records with session_id 'demo-session-127.0.0.1'\n", count)

	fmt.Println("Wishlist table successfully updated to support session_id!")
}
