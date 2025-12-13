package search

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductResult is returned for product search results
type ProductResult struct {
	ID      int64   `json:"id"`
	Title   string  `json:"title"`
	Slug    string  `json:"slug"`
	Price   float64 `json:"price,omitempty"`
	Image   string  `json:"image,omitempty"`
	Excerpt string  `json:"excerpt,omitempty"`
	Score   float32 `json:"score,omitempty"`
	Link    string  `json:"link"`
}

type CategoryResult struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
	Link  string `json:"link"`
}

type OfferResult struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

type BannerResult struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

// timeout used for DB search operations
const dbSearchTimeout = 2500 * time.Millisecond

// helper: basic normalization - trim is done by caller; we rely on DB unaccent/lower

// SearchProducts runs a simple search using SQL LIKE for immediate functionality
func SearchProducts(parentCtx context.Context, db *pgxpool.Pool, q string, limit int) ([]ProductResult, error) {
	if db == nil {
		return nil, errors.New("db pool is nil")
	}
	ctx, cancel := context.WithTimeout(parentCtx, dbSearchTimeout)
	defer cancel()

	var results []ProductResult

	// Simple search using SQL LIKE - works immediately without extensions
	simpleQuery := `
    SELECT 
      p.uuid_id, p.title, p.slug, p.description,
      COALESCE(MIN(v.price_cents), 0)::double precision as price,
      COALESCE(MIN(v.currency), 'INR') as currency,
      COALESCE(m.image_key, '') as image,
      1.0 as score
    FROM products p
    LEFT JOIN product_variants v ON v.product_id = p.uuid_id
    LEFT JOIN (
      SELECT DISTINCT ON (product_id) product_id, image_key
      FROM product_media
      ORDER BY product_id, sort_order
    ) m ON m.product_id = p.uuid_id
    WHERE LOWER(p.title) LIKE LOWER($1) OR LOWER(p.description) LIKE LOWER($1)
    GROUP BY p.uuid_id, p.title, p.slug, p.description, m.image_key
    ORDER BY p.title
    LIMIT $2`

	rows, err := db.Query(ctx, simpleQuery, "%"+q+"%", limit)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r ProductResult
			var price float64
			var currency string
			if err := rows.Scan(&r.ID, &r.Title, &r.Slug, &r.Excerpt, &price, &currency, &r.Image, &r.Score); err != nil {
				continue
			}
			r.Price = price
			r.Link = fmt.Sprintf("/product/%s", r.Slug)
			results = append(results, r)
		}
		if len(results) > 0 {
			return results, nil
		}
	}

	// Fallback: trigram similarity on title and slug
	trigramQuery := `
    SELECT id, title, slug, COALESCE(price,0)::double precision, COALESCE(image,'') as image,
      '' as excerpt, GREATEST(similarity(lower(title), lower($1)), similarity(lower(slug), lower($1))) as score
    FROM products
    WHERE lower(title) % lower($1) OR lower(slug) % lower($1) OR similarity(lower(title), lower($1)) > 0.15
    ORDER BY score DESC
    LIMIT $2` + ";"

	rows2, err2 := db.Query(ctx, trigramQuery, q, limit)
	if err2 != nil {
		return nil, err2
	}
	defer rows2.Close()
	for rows2.Next() {
		var r ProductResult
		var price float64
		if err := rows2.Scan(&r.ID, &r.Title, &r.Slug, &price, &r.Image, &r.Excerpt, &r.Score); err != nil {
			continue
		}
		r.Price = price
		r.Link = fmt.Sprintf("/product/%s", r.Slug)
		results = append(results, r)
	}

	return results, nil
}

// SearchCategories searches categories table using trigram similarity on name/description
func SearchCategories(parentCtx context.Context, db *pgxpool.Pool, q string, limit int) ([]CategoryResult, error) {
	if db == nil {
		return nil, errors.New("db pool is nil")
	}
	ctx, cancel := context.WithTimeout(parentCtx, dbSearchTimeout)
	defer cancel()

	query := `
    SELECT id, title, COALESCE(slug,'') as slug
    FROM categories
    WHERE lower(search_text) % lower($1) OR similarity(lower(search_text), lower($1)) > 0.15
    ORDER BY similarity(lower(search_text), lower($1)) DESC
    LIMIT $2` + ";"

	rows, err := db.Query(ctx, query, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CategoryResult
	for rows.Next() {
		var r CategoryResult
		if err := rows.Scan(&r.ID, &r.Title, &r.Slug); err != nil {
			continue
		}
		r.Link = fmt.Sprintf("/category/%s", r.Slug)
		out = append(out, r)
	}
	return out, nil
}

// SearchOffers searches offers table
func SearchOffers(parentCtx context.Context, db *pgxpool.Pool, q string, limit int) ([]OfferResult, error) {
	if db == nil {
		return nil, errors.New("db pool is nil")
	}
	ctx, cancel := context.WithTimeout(parentCtx, dbSearchTimeout)
	defer cancel()

	query := `
    SELECT id, title
    FROM offers
    WHERE lower(search_text) % lower($1) OR similarity(lower(search_text), lower($1)) > 0.15
    ORDER BY similarity(lower(search_text), lower($1)) DESC
    LIMIT $2` + ";"

	rows, err := db.Query(ctx, query, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OfferResult
	for rows.Next() {
		var r OfferResult
		if err := rows.Scan(&r.ID, &r.Title); err != nil {
			continue
		}
		r.Link = fmt.Sprintf("/offer/%d", r.ID)
		out = append(out, r)
	}
	return out, nil
}

// SearchBanners searches banners table
func SearchBanners(parentCtx context.Context, db *pgxpool.Pool, q string, limit int) ([]BannerResult, error) {
	if db == nil {
		return nil, errors.New("db pool is nil")
	}
	ctx, cancel := context.WithTimeout(parentCtx, dbSearchTimeout)
	defer cancel()

	query := `
    SELECT id, title, COALESCE(target,'') as target
    FROM banners
    WHERE lower(search_text) % lower($1) OR similarity(lower(search_text), lower($1)) > 0.15
    ORDER BY similarity(lower(search_text), lower($1)) DESC
    LIMIT $2` + ";"

	rows, err := db.Query(ctx, query, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BannerResult
	for rows.Next() {
		var r BannerResult
		var target string
		if err := rows.Scan(&r.ID, &r.Title, &target); err != nil {
			continue
		}
		r.Link = target
		out = append(out, r)
	}
	return out, nil
}

// FallbackTwoProducts returns top 3 bestsellers or newest products
func FallbackTwoProducts(parentCtx context.Context, db *pgxpool.Pool) ([]ProductResult, error) {
	if db == nil {
		return nil, errors.New("db pool is nil")
	}
	ctx, cancel := context.WithTimeout(parentCtx, dbSearchTimeout)
	defer cancel()

	query := `
    SELECT 
      p.uuid_id, p.title, p.slug,
      COALESCE(MIN(v.price_cents), 0)::double precision as price,
      COALESCE(MIN(v.currency), 'INR') as currency,
      COALESCE(m.image_key, '') as image
    FROM products p
    LEFT JOIN product_variants v ON v.product_id = p.uuid_id
    LEFT JOIN (
      SELECT DISTINCT ON (product_id) product_id, image_key
      FROM product_media
      ORDER BY product_id, sort_order
    ) m ON m.product_id = p.uuid_id
    GROUP BY p.uuid_id, p.title, p.slug, m.image_key
    ORDER BY p.title
    LIMIT 3`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProductResult
	for rows.Next() {
		var r ProductResult
		var price float64
		var currency string
		if err := rows.Scan(&r.ID, &r.Title, &r.Slug, &price, &currency, &r.Image); err != nil {
			continue
		}
		r.Price = price
		r.Link = fmt.Sprintf("/products/%s", r.Slug)
		out = append(out, r)
	}
	return out, nil
}
