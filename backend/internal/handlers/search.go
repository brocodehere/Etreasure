package handlers

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/etreasure/backend/internal/search"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// token bucket for rate limiting
type tokenBucket struct {
	tokens     int
	last       time.Time
	refillRate float64
	capacity   int
}

func (b *tokenBucket) allow() bool {
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.last = now
	refill := int(elapsed * b.refillRate)
	if refill > 0 {
		b.tokens += refill
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
	}
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

type SearchHandler struct {
	DB         *pgxpool.Pool
	limiterMu  sync.Mutex
	buckets    map[string]*tokenBucket
	minLetters int
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(db *pgxpool.Pool) *SearchHandler {
	return &SearchHandler{DB: db, buckets: make(map[string]*tokenBucket), minLetters: 3}
}

var alphaRe = regexp.MustCompile(`[A-Za-z]`)

func (h *SearchHandler) getBucket(ip string) *tokenBucket {
	h.limiterMu.Lock()
	defer h.limiterMu.Unlock()
	b, ok := h.buckets[ip]
	if !ok {
		b = &tokenBucket{tokens: 60, last: time.Now(), refillRate: 1.0, capacity: 60}
		h.buckets[ip] = b
	}
	return b
}

// Search handles GET /api/search?q=query&limit=12
func (h *SearchHandler) Search(c *gin.Context) {
	ip := c.ClientIP()
	b := h.getBucket(ip)
	if !b.allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	q := c.Query("q")
	limit := 12
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if len(q) < h.minLetters || !alphaRe.MatchString(q) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short"})
		return
	}

	start := time.Now()
	ctx := c.Request.Context()

	// 1) products
	products, err := search.SearchProducts(ctx, h.DB, q, limit)
	if err != nil {
		log.Printf("search.SearchProducts error: %v", err)
	}
	if len(products) > 0 {
		var out []gin.H
		for _, p := range products {
			out = append(out, gin.H{"type": "product", "id": p.ID, "title": p.Title, "slug": p.Slug, "image": p.Image, "price": p.Price, "excerpt": p.Excerpt, "link": p.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": out, "source": "products", "took_ms": took})
		return
	}

	// 2) categories
	cats, err := search.SearchCategories(ctx, h.DB, q, limit)
	if err != nil {
		log.Printf("search.SearchCategories error: %v", err)
	}
	if len(cats) > 0 {
		var out []gin.H
		for _, it := range cats {
			out = append(out, gin.H{"type": "category", "id": it.ID, "title": it.Title, "slug": it.Slug, "link": it.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": out, "source": "categories", "took_ms": took})
		return
	}

	// 3) offers
	offers, err := search.SearchOffers(ctx, h.DB, q, limit)
	if err != nil {
		log.Printf("search.SearchOffers error: %v", err)
	}
	if len(offers) > 0 {
		var out []gin.H
		for _, it := range offers {
			out = append(out, gin.H{"type": "offer", "id": it.ID, "title": it.Title, "link": it.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": out, "source": "offers", "took_ms": took})
		return
	}

	// 4) banners
	banners, err := search.SearchBanners(ctx, h.DB, q, limit)
	if err != nil {
		log.Printf("search.SearchBanners error: %v", err)
	}
	if len(banners) > 0 {
		var out []gin.H
		for _, it := range banners {
			out = append(out, gin.H{"type": "banner", "id": it.ID, "title": it.Title, "link": it.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": out, "source": "banners", "took_ms": took})
		return
	}

	// 5) fallback
	fallback, err := search.FallbackTwoProducts(ctx, h.DB)
	if err != nil {
		log.Printf("search.FallbackTwoProducts error: %v", err)
	}
	var out []gin.H
	for _, p := range fallback {
		out = append(out, gin.H{"type": "product", "id": p.ID, "title": p.Title, "slug": p.Slug, "image": p.Image, "price": p.Price, "excerpt": p.Excerpt, "link": p.Link, "note": "suggested"})
	}
	took := time.Since(start).Milliseconds()
	c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
	c.JSON(http.StatusOK, gin.H{"q": q, "results": out, "source": "fallback", "took_ms": took})
}

// Suggest handles GET /api/search/suggest?q=query&limit=8
func (h *SearchHandler) Suggest(c *gin.Context) {
	ip := c.ClientIP()
	b := h.getBucket(ip)
	if !b.allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	q := c.Query("q")
	limit := 8
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if len(q) < h.minLetters || !alphaRe.MatchString(q) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short"})
		return
	}

	start := time.Now()
	ctx := c.Request.Context()

	// For suggestions, use the same search but return with shorter cache
	products, err := search.SearchProducts(ctx, h.DB, q, limit)
	log.Printf("SearchProducts for query '%s': found %d products, error: %v", q, len(products), err)
	var out []gin.H
	for _, p := range products {
		log.Printf("Adding product to results: %s (ID: %d)", p.Title, p.ID)
		out = append(out, gin.H{"type": "product", "id": p.ID, "title": p.Title, "slug": p.Slug, "image": p.Image, "price": p.Price, "excerpt": p.Excerpt, "link": p.Link})
	}

	// If no products, try categories
	if len(out) == 0 {
		cats, _ := search.SearchCategories(ctx, h.DB, q, limit)
		for _, it := range cats {
			out = append(out, gin.H{"type": "category", "id": it.ID, "title": it.Title, "slug": it.Slug, "link": it.Link})
		}
	}

	// If still no results, add fallback
	if len(out) == 0 {
		fallback, err := search.FallbackTwoProducts(ctx, h.DB)
		log.Printf("FallbackTwoProducts: found %d products, error: %v", len(fallback), err)
		for _, p := range fallback {
			log.Printf("Adding fallback product: %s (ID: %d)", p.Title, p.ID)
			out = append(out, gin.H{"type": "product", "id": p.ID, "title": p.Title, "slug": p.Slug, "image": p.Image, "price": p.Price, "excerpt": p.Excerpt, "link": p.Link, "note": "suggested"})
		}
	}

	took := time.Since(start).Milliseconds()
	c.Header("Cache-Control", "public, max-age=30")
	c.JSON(http.StatusOK, gin.H{"q": q, "results": out, "took_ms": took})
}

// Facets handles GET /api/search/facets?q=query (optional)
func (h *SearchHandler) Facets(c *gin.Context) {
	q := c.Query("q")
	start := time.Now()

	// Return empty facet response for now (can be enhanced later)
	// Facets would typically include category counts, price ranges, etc.
	facets := gin.H{
		"categories": []gin.H{},
		"price_range": gin.H{
			"min": 0,
			"max": 100000,
		},
	}

	took := time.Since(start).Milliseconds()
	c.Header("Cache-Control", "public, max-age=600")
	c.JSON(http.StatusOK, gin.H{"q": q, "facets": facets, "took_ms": took})
}

// Health checks search infrastructure
func (h *SearchHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	start := time.Now()

	// Simple health check: verify DB connection is alive
	if err := h.DB.Ping(ctx); err != nil {
		took := time.Since(start).Milliseconds()
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"error":    err.Error(),
			"took_ms":  took,
			"pg_trgm":  "unknown",
			"unaccent": "unknown",
		})
		return
	}

	// Check if extensions are available
	var pgTrgm, unaccent bool
	h.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm')").Scan(&pgTrgm)
	h.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'unaccent')").Scan(&unaccent)

	took := time.Since(start).Milliseconds()
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"pg_trgm":   pgTrgm,
		"unaccent":  unaccent,
		"took_ms":   took,
		"timestamp": time.Now(),
	})
}

// Reindex handles POST /api/search/reindex (admin only)
func (h *SearchHandler) Reindex(c *gin.Context) {
	ctx := c.Request.Context()
	start := time.Now()

	// TODO: Add authentication middleware to ensure only admins can call this
	// For now, we'll skip auth, but in production add:
	// if !isAdmin(c) { c.JSON(http.StatusForbidden, ...); return }

	// Rebuild search_vector for all products
	result, err := h.DB.Exec(ctx, `
		UPDATE products
		SET search_vector = (
			setweight(to_tsvector('simple', coalesce(unaccent(title),'')), 'A') ||
			setweight(to_tsvector('simple', coalesce(unaccent(array_to_string(tags, ' '),''))), 'B') ||
			setweight(to_tsvector('simple', coalesce(unaccent(brand),'')), 'B') ||
			setweight(to_tsvector('simple', coalesce(unaccent(description),'')), 'C')
		)
	`)

	var count int64
	if err == nil {
		count = result.RowsAffected()
	}

	duration := time.Since(start)

	c.JSON(http.StatusOK, gin.H{
		"message":      "search index rebuilt successfully",
		"updatedCount": count,
		"durationMs":   duration.Milliseconds(),
		"timestamp":    time.Now(),
	})
}
