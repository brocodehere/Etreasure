package search

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// These are integration-style tests and require a real Postgres database.
// Set TEST_DATABASE_URL in the environment to run them.

func getTestDB(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration tests")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	cfg.MaxConns = 2
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return db
}

func TestSearchProductsBasic(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()

	// Insert a test product; assumes products table has at least (title,slug,description,price)
	// We try to be non-destructive: create product only if not exists.
	title := "Test Saree Integration"
	slug := "test-saree-integration"
	_, _ = db.Exec(ctx, `INSERT INTO products (title, slug, description, price, created_at)
        VALUES ($1,$2,$3,$4, now()) ON CONFLICT (slug) DO NOTHING`, title, slug, "integration test product", 999.0)

	// Give DB a moment to backfill triggers
	time.Sleep(500 * time.Millisecond)

	res, err := SearchProducts(ctx, db, "Saree Integration", 5)
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("expected at least one result for query, got 0")
	}
}
