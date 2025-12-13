package handlers

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestDBOrSkip(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping handler integration tests")
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

func TestHandlerShortQuery(t *testing.T) {
	db := getTestDBOrSkip(t)
	h := NewSearchHandler(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/search", h.SearchHandler)

	req := httptest.NewRequest("GET", "/api/search?q=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400 for short query, got %d", w.Code)
	}
}
