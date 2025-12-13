package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContentPage represents a static page content (About, Policies, etc.)
type ContentPage struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // about, policy, etc.
	IsActive  bool      `json:"is_active"`
	MetaTitle string    `json:"meta_title"`
	MetaDesc  string    `json:"meta_description"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FAQ represents a frequently asked question
type FAQ struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Category  string    `json:"category"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateContentPage creates or updates a content page
func (h *Handler) CreateContentPage(c *gin.Context) {
	ctx := c.Request.Context()

	var page ContentPage
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate UUID if not provided
	if page.ID == "" {
		page.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	page.CreatedAt = now
	page.UpdatedAt = now

	// Insert or update page
	query := `
		INSERT INTO content_pages (id, title, slug, content, type, is_active, meta_title, meta_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			slug = EXCLUDED.slug,
			content = EXCLUDED.content,
			type = EXCLUDED.type,
			is_active = EXCLUDED.is_active,
			meta_title = EXCLUDED.meta_title,
			meta_description = EXCLUDED.meta_description,
			updated_at = EXCLUDED.updated_at
		RETURNING id, title, slug, content, type, is_active, meta_title, meta_description, created_at, updated_at
	`

	err := h.DB.QueryRow(ctx, query,
		page.ID, page.Title, page.Slug, page.Content, page.Type, page.IsActive,
		page.MetaTitle, page.MetaDesc, page.CreatedAt, page.UpdatedAt,
	).Scan(
		&page.ID, &page.Title, &page.Slug, &page.Content, &page.Type, &page.IsActive,
		&page.MetaTitle, &page.MetaDesc, &page.CreatedAt, &page.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save content page"})
		return
	}

	c.JSON(http.StatusOK, page)
}

// ListContentPages retrieves all content pages
func (h *Handler) ListContentPages(c *gin.Context) {
	ctx := c.Request.Context()

	query := `
		SELECT id, title, slug, content, type, is_active, meta_title, meta_description, created_at, updated_at
		FROM content_pages
		ORDER BY type, title
	`

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch content pages"})
		return
	}
	defer rows.Close()

	var pages []ContentPage
	for rows.Next() {
		var page ContentPage
		err := rows.Scan(
			&page.ID, &page.Title, &page.Slug, &page.Content, &page.Type, &page.IsActive,
			&page.MetaTitle, &page.MetaDesc, &page.CreatedAt, &page.UpdatedAt,
		)
		if err != nil {
			continue
		}
		pages = append(pages, page)
	}

	c.JSON(http.StatusOK, gin.H{"data": pages})
}

// GetContentPage retrieves a specific content page by slug
func (h *Handler) GetContentPage(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	var page ContentPage
	query := `
		SELECT id, title, slug, content, type, is_active, meta_title, meta_description, created_at, updated_at
		FROM content_pages
		WHERE slug = $1 AND is_active = true
	`

	err := h.DB.QueryRow(ctx, query, slug).Scan(
		&page.ID, &page.Title, &page.Slug, &page.Content, &page.Type, &page.IsActive,
		&page.MetaTitle, &page.MetaDesc, &page.CreatedAt, &page.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content page not found"})
		return
	}

	c.JSON(http.StatusOK, page)
}

// DeleteContentPage deletes a content page
func (h *Handler) DeleteContentPage(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	_, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	_, err = h.DB.Exec(ctx, "DELETE FROM content_pages WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete content page"})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateFAQ creates or updates an FAQ
func (h *Handler) CreateFAQ(c *gin.Context) {
	ctx := c.Request.Context()

	var faq FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate UUID if not provided
	if faq.ID == "" {
		faq.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	faq.CreatedAt = now
	faq.UpdatedAt = now

	// Insert or update FAQ
	query := `
		INSERT INTO faqs (id, question, answer, category, is_active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			question = EXCLUDED.question,
			answer = EXCLUDED.answer,
			category = EXCLUDED.category,
			is_active = EXCLUDED.is_active,
			sort_order = EXCLUDED.sort_order,
			updated_at = EXCLUDED.updated_at
		RETURNING id, question, answer, category, is_active, sort_order, created_at, updated_at
	`

	err := h.DB.QueryRow(ctx, query,
		faq.ID, faq.Question, faq.Answer, faq.Category, faq.IsActive, faq.SortOrder,
		faq.CreatedAt, faq.UpdatedAt,
	).Scan(
		&faq.ID, &faq.Question, &faq.Answer, &faq.Category, &faq.IsActive, &faq.SortOrder,
		&faq.CreatedAt, &faq.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save FAQ"})
		return
	}

	c.JSON(http.StatusOK, faq)
}

// ListFAQs retrieves all FAQs
func (h *Handler) ListFAQs(c *gin.Context) {
	ctx := c.Request.Context()

	query := `
		SELECT id, question, answer, category, is_active, sort_order, created_at, updated_at
		FROM faqs
		ORDER BY sort_order, category, question
	`

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FAQs"})
		return
	}
	defer rows.Close()

	var faqs []FAQ
	for rows.Next() {
		var faq FAQ
		err := rows.Scan(
			&faq.ID, &faq.Question, &faq.Answer, &faq.Category, &faq.IsActive,
			&faq.SortOrder, &faq.CreatedAt, &faq.UpdatedAt,
		)
		if err != nil {
			continue
		}
		faqs = append(faqs, faq)
	}

	c.JSON(http.StatusOK, gin.H{"data": faqs})
}

// DeleteFAQ deletes an FAQ
func (h *Handler) DeleteFAQ(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	_, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	_, err = h.DB.Exec(ctx, "DELETE FROM faqs WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete FAQ"})
		return
	}

	c.Status(http.StatusNoContent)
}
