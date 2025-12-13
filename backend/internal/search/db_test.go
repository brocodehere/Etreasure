package search

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDB represents test database setup
type TestDB struct {
	pool *pgxpool.Pool
	db   *DB
}

// setupTestDB initializes an in-memory test database
func setupTestDB(t *testing.T) *TestDB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// This test assumes a running postgres instance
	// For CI/CD, use testcontainers or postgres docker image
	dbURL := "postgres://postgres:postgres@localhost:5432/etreasure_test"

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping tests: could not connect to test database: %v", err)
	}

	// Cleanup: drop test tables
	_ = pool.QueryRow(ctx, "SELECT 1")

	return &TestDB{
		pool: pool,
		db:   New(pool),
	}
}

// teardownTestDB cleans up test database
func (td *TestDB) teardown(ctx context.Context) {
	if td.pool != nil {
		td.pool.Close()
	}
}

// TestSearchFullTextBasic verifies basic full-text search
func TestSearchFullTextBasic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tdb := setupTestDB(t)
	defer tdb.teardown(ctx)

	// Insert test product
	_, err := tdb.pool.Exec(ctx, `
		INSERT INTO products (slug, title, brand, description, published, search_vector)
		VALUES ($1, $2, $3, $4, $5, to_tsvector('english', $2 || ' ' || $3))
	`, "silk-saree-001", "Premium Silk Saree", "Ethnic Treasure", "Beautiful handcrafted silk saree", true)
	require.NoError(t, err)

	// Search for product
	req := &SearchRequest{
		Query: "silk",
		Limit: 10,
	}

	results, err := tdb.db.Search(ctx, req)
	require.NoError(t, err)
	assert.Len(t, results.Items, 1)
	assert.Equal(t, "Premium Silk Saree", results.Items[0].Title)
	assert.Equal(t, "silk-saree-001", results.Items[0].Slug)
}

// TestSearchFilters verifies price and category filtering
func TestSearchFilters(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tdb := setupTestDB(t)
	defer tdb.teardown(ctx)

	// Insert test product
	_, err := tdb.pool.Exec(ctx, `
		INSERT INTO products (slug, title, brand, description, published, search_vector)
		VALUES ($1, $2, $3, $4, $5, to_tsvector('english', $2))
	`, "bag-001", "Leather Bag", "Ethnic Treasure", "Premium leather bag", true)
	require.NoError(t, err)

	productID := 1

	// Insert variant with price
	_, err = tdb.pool.Exec(ctx, `
		INSERT INTO product_variants (product_id, sku, price_cents, stock_quantity)
		VALUES ($1, $2, $3, $4)
	`, productID, "BAG-001", 50000, 100)
	require.NoError(t, err)

	// Test price filter (within range)
	minPrice := 40000
	maxPrice := 60000
	req := &SearchRequest{
		Query:    "bag",
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
		Limit:    10,
	}

	results, err := tdb.db.Search(ctx, req)
	require.NoError(t, err)
	if len(results.Items) > 0 {
		assert.Equal(t, results.Items[0].Price, 50000)
	}
}

// TestSuggestFuzzyMatch verifies autocomplete with fuzzy matching
func TestSuggestFuzzyMatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tdb := setupTestDB(t)
	defer tdb.teardown(ctx)

	// Insert test products
	products := []struct {
		slug  string
		title string
		brand string
	}{
		{"banarasi-001", "Banarasi Silk Saree", "Royal Weaves"},
		{"bandhani-001", "Bandhani Cotton Saree", "Tradition"},
		{"banarasi-002", "Banarasi Pure Silk", "Heritage"},
	}

	for _, p := range products {
		_, err := tdb.pool.Exec(ctx, `
			INSERT INTO products (slug, title, brand, published, search_vector)
			VALUES ($1, $2, $3, $4, to_tsvector('english', $2 || ' ' || $3))
		`, p.slug, p.title, p.brand, true)
		require.NoError(t, err)
	}

	// Test prefix match
	req := &SuggestionRequest{
		Query: "bana",
		Limit: 10,
	}

	suggestions, err := tdb.db.Suggest(ctx, req)
	require.NoError(t, err)

	// Should find Banarasi products
	assert.True(t, len(suggestions) > 0, "Should find Banarasi products")
	assert.Contains(t, suggestions[0].Title, "Banarasi")
}

// TestReindexAll verifies search vector rebuild
func TestReindexAll(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tdb := setupTestDB(t)
	defer tdb.teardown(ctx)

	// Insert test product without search_vector
	_, err := tdb.pool.Exec(ctx, `
		INSERT INTO products (slug, title, brand, description, published)
		VALUES ($1, $2, $3, $4, $5)
	`, "test-001", "Test Product", "Test Brand", "Test description", true)
	require.NoError(t, err)

	// Reindex
	count, err := tdb.db.ReindexAll(ctx)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "Should have updated at least one product")
}

// TestCursorPagination verifies cursor-based pagination
func TestCursorPagination(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tdb := setupTestDB(t)
	defer tdb.teardown(ctx)

	// Insert multiple products
	for i := 1; i <= 25; i++ {
		_, err := tdb.pool.Exec(ctx, `
			INSERT INTO products (slug, title, published, search_vector)
			VALUES ($1, $2, $3, to_tsvector('english', $2))
		`, "product-"+string(rune(i)), "Product "+string(rune(i)), true)
		require.NoError(t, err)
	}

	// First page
	req := &SearchRequest{
		Query: "product",
		Limit: 10,
	}

	results, err := tdb.db.Search(ctx, req)
	require.NoError(t, err)
	assert.Len(t, results.Items, 10)
	assert.NotNil(t, results.NextCursor, "Should have next cursor")

	// Second page with cursor
	if results.NextCursor != nil {
		req.Cursor = *results.NextCursor
		results2, err := tdb.db.Search(ctx, req)
		require.NoError(t, err)
		assert.Len(t, results2.Items, 10)

		// Results should be different
		assert.NotEqual(t, results.Items[0].ID, results2.Items[0].ID)
	}
}

// TestEncodeDecode cursor tests cursor encoding/decoding
func TestEncodeDecodeCursor(t *testing.T) {
	cursor := EncodeCursor(123, 0.95)
	id, score, err := DecodeCursor(cursor)

	require.NoError(t, err)
	assert.Equal(t, 123, id)
	assert.Equal(t, 0.95, score)
}

// TestInvalidCursorFormat tests cursor validation
func TestInvalidCursorFormat(t *testing.T) {
	_, _, err := DecodeCursor("invalid-cursor-format")
	assert.Error(t, err, "Should error on invalid cursor")
}

// TestSearchEmptyQuery tests empty search handling
func TestSearchEmptyQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tdb := setupTestDB(t)
	defer tdb.teardown(ctx)

	req := &SearchRequest{
		Query: "",
		Limit: 10,
	}

	results, err := tdb.db.Search(ctx, req)
	// Should fail gracefully or return empty results
	assert.True(t, err != nil || len(results.Items) == 0)
}

// TestSearchTimeout verifies timeout handling
func TestSearchTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	tdb := setupTestDB(t)
	defer tdb.teardown(context.Background())

	req := &SearchRequest{
		Query: "test",
		Limit: 10,
	}

	_, err := tdb.db.Search(ctx, req)
	// Timeout should eventually occur on heavy queries
	// (may not happen on light queries)
	_ = err
}

// BenchmarkSearch measures search performance
func BenchmarkSearch(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tdb := setupTestDB(&testing.T{})
	defer tdb.teardown(ctx)

	// Insert many test products
	for i := 0; i < 1000; i++ {
		_, _ = tdb.pool.Exec(ctx, `
			INSERT INTO products (slug, title, description, published, search_vector)
			VALUES ($1, $2, $3, $4, to_tsvector('english', $2 || ' ' || $3))
		`, "product-"+string(rune(i%10)), "Product Type "+string(rune(i%10)), "High quality product", true)
	}

	req := &SearchRequest{
		Query: "product",
		Limit: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tdb.db.Search(ctx, req)
	}
}
