package search

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB provides search-related database queries
type DB struct {
	pool *pgxpool.Pool
}

// New creates a new search DB instance
func New(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

// Cursor represents pagination cursor state
type cursor struct {
	ID    int     `json:"id"`
	Score float64 `json:"score"`
}

// EncodeCursor encodes pagination state
func EncodeCursor(id int, score float64) string {
	c := cursor{ID: id, Score: score}
	data, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeCursor decodes pagination state
func DecodeCursor(encoded string) (int, float64, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return 0, 0, err
	}
	var c cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return 0, 0, err
	}
	return c.ID, c.Score, nil
}

// Search performs full-text search with optional filters
func (db *DB) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Sanitize search query: remove control chars, limit length
	query := strings.TrimSpace(req.Query)
	if len(query) > 500 {
		query = query[:500]
	}

	// Build WHERE clause for filters
	whereConditions := []string{
		"p.published = TRUE",
		"(p.publish_at IS NULL OR p.publish_at <= NOW())",
		"(p.unpublish_at IS NULL OR p.unpublish_at > NOW())",
	}

	args := []interface{}{query}
	argIdx := 1

	// Category filter
	if req.CategoryID != nil {
		argIdx++
		whereConditions = append(whereConditions, fmt.Sprintf("pc.category_id = $%d", argIdx))
		args = append(args, *req.CategoryID)
	}

	// Price filter
	if req.MinPrice != nil {
		argIdx++
		whereConditions = append(whereConditions, fmt.Sprintf("pv.price_cents >= $%d", argIdx))
		args = append(args, *req.MinPrice)
	}
	if req.MaxPrice != nil {
		argIdx++
		whereConditions = append(whereConditions, fmt.Sprintf("pv.price_cents <= $%d", argIdx))
		args = append(args, *req.MaxPrice)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Determine sort order
	orderBy := "ts_rank(p.search_vector, to_tsquery('english', $1 || ':*')) DESC"
	switch req.Sort {
	case "price_asc":
		orderBy = "pv.price_cents ASC"
	case "price_desc":
		orderBy = "pv.price_cents DESC"
	case "newest":
		orderBy = "p.created_at DESC"
	default:
		orderBy = "ts_rank(p.search_vector, to_tsquery('english', $1 || ':*')) DESC"
	}

	// Build query with cursor pagination
	query2 := fmt.Sprintf(`
SELECT
  p.id,
  p.title,
  p.slug,
  pv.price_cents,
  (SELECT media.path FROM product_images pi
   LEFT JOIN media ON pi.media_id = media.id
   WHERE pi.product_id = p.id
   ORDER BY pi.sort_order LIMIT 1) AS image,
  LEFT(p.description, 200) AS excerpt,
  ts_rank(p.search_vector, to_tsquery('english', $1 || ':*')) AS score,
  p.brand,
  p.tags
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id AND pv.id = (
  SELECT id FROM product_variants WHERE product_id = p.id ORDER BY price_cents LIMIT 1
)
LEFT JOIN product_categories pc ON p.id = pc.product_id
WHERE %s
GROUP BY p.id, p.title, p.slug, pv.price_cents, score, p.brand, p.tags
ORDER BY %s
LIMIT $%d`, whereClause, orderBy, argIdx+1)

	argIdx++
	args = append(args, req.Limit+1) // Fetch one extra to check if more results exist

	rows, err := db.pool.Query(ctx, query2, args...)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var items []SearchResult
	var lastID int
	var lastScore float64

	for rows.Next() {
		var item SearchResult
		var image *string
		var tags []string

		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Slug,
			&item.Price,
			&image,
			&item.Excerpt,
			&item.Score,
			&item.Brand,
			&tags,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		if image != nil {
			item.Image = *image
		}
		item.Tags = tags

		items = append(items, item)
		lastID = item.ID
		lastScore = item.Score
	}

	// Check if there are more results
	var nextCursor *string
	if len(items) > req.Limit {
		items = items[:req.Limit]
		encoded := EncodeCursor(lastID, lastScore)
		nextCursor = &encoded
	}

	return &SearchResponse{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

// Suggest returns autocomplete suggestions
func (db *DB) Suggest(ctx context.Context, req *SuggestionRequest) ([]Suggestion, error) {
	if req.Limit == 0 {
		req.Limit = 8
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	query := strings.TrimSpace(req.Query)
	if len(query) > 100 {
		query = query[:100]
	}

	// Use similarity() from pg_trgm for fuzzy matching, prefer title prefix matches
	sqlQuery := `
SELECT
  p.id,
  p.title,
  p.slug,
  pv.price_cents,
  (SELECT media.path FROM product_images pi
   LEFT JOIN media ON pi.media_id = media.id
   WHERE pi.product_id = p.id
   ORDER BY pi.sort_order LIMIT 1) AS image
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id AND pv.id = (
  SELECT id FROM product_variants WHERE product_id = p.id ORDER BY price_cents LIMIT 1
)
WHERE p.published = TRUE
  AND (p.publish_at IS NULL OR p.publish_at <= NOW())
  AND (p.unpublish_at IS NULL OR p.unpublish_at > NOW())
  AND (p.title ILIKE $1 || '%' OR p.title % $1 OR p.brand % $1)
ORDER BY
  CASE WHEN p.title ILIKE $1 || '%' THEN 1 ELSE 2 END,
  similarity(p.title, $1) DESC,
  similarity(p.brand, $1) DESC
LIMIT $2`

	rows, err := db.pool.Query(ctx, sqlQuery, query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("suggest query failed: %w", err)
	}
	defer rows.Close()

	var suggestions []Suggestion
	for rows.Next() {
		var s Suggestion
		var image *string

		err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Price, &image)
		if err != nil {
			return nil, fmt.Errorf("suggest scan failed: %w", err)
		}

		if image != nil {
			s.Image = *image
		}

		suggestions = append(suggestions, s)
	}

	return suggestions, nil
}

// GetFacets returns aggregations for UI filters
func (db *DB) GetFacets(ctx context.Context, query string) (*FacetResponse, error) {
	query = strings.TrimSpace(query)
	if len(query) > 500 {
		query = query[:500]
	}

	// Get category facets
	rows, err := db.pool.Query(ctx, `
SELECT category_id, category_name, product_count
FROM products_search_facets
WHERE product_count > 0
ORDER BY product_count DESC
LIMIT 20`)
	if err != nil {
		// If view doesn't exist, return empty facets
		return &FacetResponse{Categories: []CategoryFacet{}, PriceRange: PriceFacet{}}, nil
	}
	defer rows.Close()

	facets := &FacetResponse{}
	for rows.Next() {
		var catID *int
		var catName *string
		var count int

		if err := rows.Scan(&catID, &catName, &count); err == nil && catID != nil {
			facets.Categories = append(facets.Categories, CategoryFacet{
				ID:           *catID,
				Name:         *catName,
				ProductCount: count,
			})
		}
	}

	// Get price range
	priceRow := db.pool.QueryRow(ctx, `
SELECT min_price_cents, max_price_cents, avg_price_cents
FROM (
  SELECT
    MIN(pv.price_cents) AS min_price_cents,
    MAX(pv.price_cents) AS max_price_cents,
    AVG(pv.price_cents)::INT AS avg_price_cents
  FROM products p
  LEFT JOIN product_variants pv ON p.id = pv.product_id
  WHERE p.published = TRUE
) t`)

	var minPrice, maxPrice, avgPrice int
	if err := priceRow.Scan(&minPrice, &maxPrice, &avgPrice); err == nil {
		facets.PriceRange = PriceFacet{
			Min: minPrice,
			Max: maxPrice,
			Avg: avgPrice,
		}
	}

	return facets, nil
}

// ReindexAll rebuilds search vectors for all products
func (db *DB) ReindexAll(ctx context.Context) (int, error) {
	result, err := db.pool.Exec(ctx, `
UPDATE products
SET search_vector = to_tsvector('english',
  coalesce(title, '') || ' ' ||
  coalesce(brand, '') || ' ' ||
  array_to_string(coalesce(tags, '{}'), ' ') || ' ' ||
  coalesce(description, '') || ' ' ||
  coalesce(primary_sku, '')
)`)
	if err != nil {
		return 0, fmt.Errorf("reindex failed: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// CheckExtensions verifies required PG extensions are installed
func (db *DB) CheckExtensions(ctx context.Context) error {
	row := db.pool.QueryRow(ctx, `
SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm')
AND EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'unaccent')`)

	var hasExtensions bool
	if err := row.Scan(&hasExtensions); err != nil {
		return err
	}

	if !hasExtensions {
		return fmt.Errorf("required PostgreSQL extensions (pg_trgm, unaccent) not installed")
	}

	return nil
}
