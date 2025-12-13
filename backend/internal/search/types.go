package search

import "time"

// SearchRequest represents a product search query
type SearchRequest struct {
	Query      string `form:"q" binding:"required"`
	CategoryID *int   `form:"category"`
	MinPrice   *int   `form:"min_price"`
	MaxPrice   *int   `form:"max_price"`
	Sort       string `form:"sort" binding:"omitempty,oneof=relevance price_asc price_desc newest"`
	Cursor     string `form:"cursor"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Fields     string `form:"fields"` // comma-separated list of fields to return
}

// SuggestionRequest represents an autocomplete query
type SuggestionRequest struct {
	Query string `form:"q" binding:"required"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=50"`
}

// SearchResult represents a single product in search results
type SearchResult struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	Slug    string   `json:"slug"`
	Price   int      `json:"price"` // price_cents from primary variant
	Image   string   `json:"image"`
	Excerpt string   `json:"excerpt"`
	Score   float64  `json:"score"`
	Brand   string   `json:"brand,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	SKU     string   `json:"sku,omitempty"`
}

// Suggestion represents an autocomplete suggestion
type Suggestion struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Price     int    `json:"price"`
	Image     string `json:"image"`
	Highlight string `json:"highlight,omitempty"`
}

// SearchResponse wraps search results with pagination
type SearchResponse struct {
	Items      []SearchResult `json:"items"`
	NextCursor *string        `json:"nextCursor,omitempty"`
	TotalCount *int           `json:"totalCount,omitempty"`
}

// FacetResponse provides aggregations for filtering
type FacetResponse struct {
	Categories []CategoryFacet `json:"categories"`
	PriceRange PriceFacet      `json:"priceRange"`
}

// CategoryFacet represents category filter options
type CategoryFacet struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ProductCount int    `json:"productCount"`
}

// PriceFacet represents price range options
type PriceFacet struct {
	Min int `json:"min"`
	Max int `json:"max"`
	Avg int `json:"avg"`
}

// ReindexStats tracks search index rebuild progress
type ReindexStats struct {
	UpdatedCount int           `json:"updatedCount"`
	Duration     time.Duration `json:"duration"`
	Timestamp    time.Time     `json:"timestamp"`
}
