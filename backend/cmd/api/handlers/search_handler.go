package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/etreasure/backend/internal/search"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Simple in-memory token bucket for rate limiting per IP
type tokenBucket struct {
	tokens     int
	last       time.Time
	refillRate float64 // tokens per second
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

// NewSearchHandler creates a search handler bound to a db pool
func NewSearchHandler(db *pgxpool.Pool) *SearchHandler {
	return &SearchHandler{
		DB:         db,
		buckets:    make(map[string]*tokenBucket),
		minLetters: 3,
	}
}

var alphaRe = regexp.MustCompile(`[A-Za-z]`)

// getBucket returns or creates a token bucket for ip
func (h *SearchHandler) getBucket(ip string) *tokenBucket {
	h.limiterMu.Lock()
	defer h.limiterMu.Unlock()
	b, ok := h.buckets[ip]
	if !ok {
		// 60 req per minute -> 1 req per second with burst up to 60
		b = &tokenBucket{tokens: 60, last: time.Now(), refillRate: 1.0, capacity: 60}
		h.buckets[ip] = b
	}
	return b
}

// SearchHandler handles GET /api/search?q=&limit=
func (h *SearchHandler) SearchHandler(c *gin.Context) {
	ip := c.ClientIP()
	b := h.getBucket(ip)
	if !b.allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	q := c.Query("q")
	limit := 12
	if l := c.Query("limit"); l != "" {
		// try parse
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if len(q) < h.minLetters || !alphaRe.MatchString(q) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short"})
		return
	}

	// normalize: trim already minimal; DB will use unaccent
	q = q

	start := time.Now()
	ctx := c.Request.Context()

	// 1) products
	products, err := search.SearchProducts(ctx, h.DB, q, limit)
	if err != nil {
		// If extension missing, search functions may return errors; fallback to ILIKE could be implemented
	}
	if len(products) > 0 {
		var res []gin.H
		for _, p := range products {
			res = append(res, gin.H{"type": "product", "id": p.ID, "title": p.Title, "slug": p.Slug, "image": p.Image, "price": p.Price, "excerpt": p.Excerpt, "link": p.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": res, "source": "products", "took_ms": took})
		return
	}

	// 2) categories
	cats, err := search.SearchCategories(ctx, h.DB, q, limit)
	if err == nil && len(cats) > 0 {
		var res []gin.H
		for _, it := range cats {
			res = append(res, gin.H{"type": "category", "id": it.ID, "title": it.Title, "slug": it.Slug, "link": it.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": res, "source": "categories", "took_ms": took})
		return
	}

	// 3) offers
	offers, err := search.SearchOffers(ctx, h.DB, q, limit)
	if err == nil && len(offers) > 0 {
		var res []gin.H
		for _, it := range offers {
			res = append(res, gin.H{"type": "offer", "id": it.ID, "title": it.Title, "link": it.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": res, "source": "offers", "took_ms": took})
		return
	}

	// 4) banners
	banners, err := search.SearchBanners(ctx, h.DB, q, limit)
	if err == nil && len(banners) > 0 {
		var res []gin.H
		for _, it := range banners {
			res = append(res, gin.H{"type": "banner", "id": it.ID, "title": it.Title, "link": it.Link})
		}
		took := time.Since(start).Milliseconds()
		c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
		c.JSON(http.StatusOK, gin.H{"q": q, "results": res, "source": "banners", "took_ms": took})
		return
	}

	// 5) fallback
	fallback, err := search.FallbackTwoProducts(ctx, h.DB)
	var res []gin.H
	if err == nil && len(fallback) > 0 {
		for _, p := range fallback {
			res = append(res, gin.H{"type": "product", "id": p.ID, "title": p.Title, "slug": p.Slug, "image": p.Image, "price": p.Price, "excerpt": p.Excerpt, "link": p.Link, "note": "suggested"})
		}
	}
	took := time.Since(start).Milliseconds()
	c.Header("Cache-Control", "public, max-age=30, stale-while-revalidate=60")
	c.JSON(http.StatusOK, gin.H{"q": q, "results": res, "source": "fallback", "took_ms": took})
}
